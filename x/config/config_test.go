package config_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/flowdev/spaghetti-cutter/x/config"
)

func testPatternList(t *testing.T) {
	specs := []struct {
		name              string
		givenPatterns     []string
		expectedMatches   []string
		expectedNoMatches []string
		expectedString    string
	}{
		{
			name:              "one-simple",
			givenPatterns:     []string{"a"},
			expectedMatches:   []string{"a"},
			expectedNoMatches: []string{"b", "aa"},
			expectedString:    "`a`",
		}, {
			name:              "many-simple",
			givenPatterns:     []string{"a", "be", "do", "ra"},
			expectedMatches:   []string{"a", "be", "do", "ra"},
			expectedNoMatches: []string{"aa", "b", "bd", "od", "ra*"},
			expectedString:    "`a`, `be`, `do`, `ra`",
		}, {
			name:              "one-star-1",
			givenPatterns:     []string{"a/*/b"},
			expectedMatches:   []string{"a/bla/b", "a//b", "a/*/b"},
			expectedNoMatches: []string{"a/bla/blue/b", "a/bla//b", "a//bla/b"},
			expectedString:    "`a/*/b`",
		}, {
			name:              "one-star-2",
			givenPatterns:     []string{"*/b"},
			expectedMatches:   []string{"bla/b", "/b", "*/b"},
			expectedNoMatches: []string{"bla/blue/b", "bla//b", "/bla/b"},
			expectedString:    "`*/b`",
		}, {
			name:              "one-star-3",
			givenPatterns:     []string{"a/*"},
			expectedMatches:   []string{"a/bla", "a/", "a/*"},
			expectedNoMatches: []string{"a/bla/blue", "a/bla/", "a//bla", "a//"},
			expectedString:    "`a/*`",
		}, {
			name:              "one-star-4",
			givenPatterns:     []string{"a/b*"},
			expectedMatches:   []string{"a/bla", "a/b", "a/b*"},
			expectedNoMatches: []string{"a/bla/blue", "a/bla/"},
			expectedString:    "`a/b*`",
		}, {
			name:              "multiple-single-stars-1",
			givenPatterns:     []string{"a/*/b/*/c"},
			expectedMatches:   []string{"a/foo/b/bar/c", "a//b//c", "a/*/b/*/c"},
			expectedNoMatches: []string{"a/foo//b/bar/c", "a/foo/b//bar/c", "a/bla/b///c"},
			expectedString:    "`a/*/b/*/c`",
		}, {
			name:              "multiple-single-stars-2",
			givenPatterns:     []string{"a/*b/c*/d"},
			expectedMatches:   []string{"a/foob/candy/d", "a/b/c/d"},
			expectedNoMatches: []string{"a/foo/candy/d", "a/foob/c/de"},
			expectedString:    "`a/*b/c*/d`",
		}, {
			name:              "double-stars",
			givenPatterns:     []string{"a/**"},
			expectedMatches:   []string{"a/foob/candy/d", "a/b/c/d/..."},
			expectedNoMatches: []string{"a/foo/candy\nd", "b/foo/b/c/d"},
			expectedString:    "`a/**`",
		}, {
			name:              "all-stars",
			givenPatterns:     []string{"a/*/b/*/c/**"},
			expectedMatches:   []string{"a/foo/b/bar/c/d/e/f", "a/foo/b/bar/c/d/**/f", "a//b//c/"},
			expectedNoMatches: []string{},
			expectedString:    "`a/*/b/*/c/**`",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			pl := &config.PatternList{}
			for _, s := range spec.givenPatterns {
				//pl.Set(s)
				_ = s
			}
			for _, s := range spec.expectedMatches {
				if !pl.MatchString(s) {
					t.Errorf("%q should match one of the patterns %q", s, spec.givenPatterns)
				}
			}
			for _, s := range spec.expectedNoMatches {
				if pl.MatchString(s) {
					t.Errorf("%q should NOT match any of the patterns %q", s, spec.givenPatterns)
				}
			}
			if spec.expectedString != pl.String() {
				t.Errorf("expected string representation %q but got: %q", spec.expectedString, pl.String())
			}
		})
	}
}

func testPatternMap(t *testing.T) {
	specs := []struct {
		name                string
		givenPairs          []string
		givenLeftPattern    string
		expectedLeftMatch   bool
		expectedRightString string
	}{
		{
			name:                "simple-pair",
			givenPairs:          []string{"a b"},
			givenLeftPattern:    "a",
			expectedLeftMatch:   true,
			expectedRightString: "`b`",
		}, {
			name:                "multiple-pairs",
			givenPairs:          []string{"a b", "a c", "a do", "a foo"},
			givenLeftPattern:    "a",
			expectedLeftMatch:   true,
			expectedRightString: "`b`, `c`, `do`, `foo`",
		}, {
			name:                "one-pair-many-stars",
			givenPairs:          []string{"a/*/b/** c/*/d/**"},
			givenLeftPattern:    "a/foo/b/bar/doo",
			expectedLeftMatch:   true,
			expectedRightString: "`c/*/d/**`",
		}, {
			name:                "all-complexity",
			givenPairs:          []string{"*/*a/** */*b/**", "*/*a/** b*/c*d/**"},
			givenLeftPattern:    "foo/bara/doo/ey",
			expectedLeftMatch:   true,
			expectedRightString: "`*/*b/**`, `b*/c*d/**`",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			pm := &config.PatternMap{}
			for _, s := range spec.givenPairs {
				//pm.Set(s)
				_ = s
			}

			pl := pm.MatchingList(spec.givenLeftPattern)

			if spec.expectedLeftMatch && pl == nil {
				t.Fatalf("expected left match for pattern %q in map %v", spec.givenLeftPattern, pm)
			} else if !spec.expectedLeftMatch && pl != nil {
				t.Fatalf("expected NO left match for pattern %q in map %v but got: %v", spec.givenLeftPattern, pm, pl)
			}
			if !spec.expectedLeftMatch {
				return
			}

			if spec.expectedRightString != pl.String() {
				t.Errorf("expected right string representation %q but got: %q", spec.expectedRightString, pl.String())
			}
		})
	}
}

func TestParse(t *testing.T) {
	specs := []struct {
		name                 string
		givenConfigFile      string
		expectedConfigString string
	}{
		{
			name:            "all-empty",
			givenConfigFile: "all-empty.json",
			expectedConfigString: "{" +
				"..... ..... ... ... `main` " +
				" " +
				"2048 false false" +
				"}",
		}, {
			name:            "config-only",
			givenConfigFile: "config-only.json",
			expectedConfigString: "{" +
				"..... " +
				"`a`: `b` " +
				"`x/**` " +
				"`pkg/db/*` " +
				"`main` " +
				"dir/bla " +
				"3072 " +
				"false " +
				"true" +
				"}",
		}, {
			name:            "args-and-config",
			givenConfigFile: "args-and-config.json",
			expectedConfigString: "{" +
				"`a`: `b` ; `c`: `d` " +
				"`pkg/mysupertool`, `pkg/x/**` " +
				"`pkg/db`, `pkg/entities` " +
				"`main`, `pkg/service` " +
				"dir/blue " +
				"4096 " +
				"true " +
				"true" +
				"}",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			cfgFile := spec.givenConfigFile
			if cfgFile != "" {
				cfgFile = filepath.Join("testdata", cfgFile)
			}
			actualConfig, err := config.Parse(cfgFile)
			if err != nil {
				t.Fatalf("got unexpected error: %v", err)
			}
			actualConfigString := fmt.Sprint(actualConfig)
			if actualConfigString != spec.expectedConfigString {
				t.Errorf("expected configuration %v, actual %v",
					spec.expectedConfigString, actualConfigString)
			}
		})
	}
}
