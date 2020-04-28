package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/peterbourgon/ff/v3"
)

// File is the name of the configuration file
const File = ".spaghetti-cutter.json"

var (
	// Value is the set value for own maps that are really sets.
	Value = struct{}{}
)

// Pattern combines the original pattern string with a compiled regular
// expression ready for efficient evaluation.
type Pattern struct {
	Pattern string
	Regexp  *regexp.Regexp
}

// PatternList is a flag.Value that collects each Set string
// into a slice, allowing for repeated flags.
type PatternList []Pattern

// Set implements flag.Value and appends a pattern to the slice.
func (pl *PatternList) Set(s string) error {
	for _, p := range *pl {
		if p.Pattern == s {
			return nil // deduplication
		}
	}
	re, err := regexpForPattern(s)
	if err != nil {
		return fmt.Errorf("unable to set pattern `%s`: %w", s, err)
	}
	*pl = append(*pl, Pattern{Pattern: s, Regexp: re})
	return nil
}

// String implements flag.Value and returns the list of
// patterns, or "..." if no patterns have been added.
func (pl *PatternList) String() string {
	if len(*pl) <= 0 {
		return "..."
	}
	var b strings.Builder
	for i, p := range *pl {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("`")
		b.WriteString(p.Pattern)
		b.WriteString("`")
	}
	return b.String()
}

// PatternGroup groups a parent/left pattern with a list of children/right
// patterns.
type PatternGroup struct {
	Left  Pattern
	Right *PatternList
}

// PatternMap is a flag.Value that collects each Set string
// into PatternGroups, allowing for repeated flags.
type PatternMap map[string]PatternGroup

// Set implements flag.Value and adds left and right patterns
// (seperated by space in s) to the PatternMap.
func (pm *PatternMap) Set(s string) error {
	var left, right string
	_, err := fmt.Sscan(s, &left, &right)
	if err != nil {
		return fmt.Errorf("unable to split pattern group '%s' into left and right patterns: %w", s, err)
	}

	group := (*pm)[left]
	if group == (PatternGroup{}) {
		re, err := regexpForPattern(s)
		if err != nil {
			return fmt.Errorf("unable to set left pattern `%s`: %w", s, err)
		}
		list := PatternList(make([]Pattern, 0, 16))
		group = PatternGroup{
			Left:  Pattern{Pattern: left, Regexp: re},
			Right: &list,
		}
		(*pm)[left] = group
	}
	return group.Right.Set(right)
}

// String implements flag.Value and returns the map of
// string sets, or "..." if no strings have been added.
func (pm *PatternMap) String() string {
	if len(*pm) <= 0 {
		return "....."
	}
	var b strings.Builder
	for left, group := range *pm {
		b.WriteString(left)
		b.WriteString(": ")
		b.WriteString(group.Right.String())
		b.WriteString(" ; ")
	}
	s := b.String()
	return s[:len(s)-3]
}

func regexpForPattern(pattern string) (*regexp.Regexp, error) {
	i := strings.Index(pattern, "**")
	n2 := len(pattern) - 2
	if i >= 0 && i < n2 {
		return nil, errors.New("illegal pattern `" + pattern + "` contains `**` before the end")
	}
	if i >= 0 {
		pattern = pattern[:n2] + ".*"
	}
	pattern = strings.ReplaceAll(pattern, `*`, `[^/]*`)
	return regexp.Compile("^" + pattern + "$")
}

// Config contains the parsed configuration.
type Config struct {
	Allow PatternMap
	Tool  PatternList
	DB    PatternList
	God   PatternList
	Root  string
	Size  uint
}

// Parse parses command line arguments and configuration file
func Parse(args []string) Config {
	const (
		usageAllow = "allowed package dependency (e.g. 'pkg/a/uses pkg/x/util')"
		usageTool  = "tool package (leave package) (e.g. 'pkg/x/**'; '**' matches anything including a '/'"
		usageDB    = "common domain/database package (can only depend on tools) " +
			"(e.g. 'pkg/*/db'; '*' matches anything except for a '/')"
		usageGod    = "god package that can see everything (package 'main' is always god)"
		usageRoot   = "root directory"
		usageSize   = "maximum size of a package in \"lines\""
		defaultSize = 4096
	)
	cfg := Config{
		Allow: make(map[string]PatternGroup),
		Tool:  make([]Pattern, 0, 16),
		DB:    make([]Pattern, 0, 16),
		God:   make([]Pattern, 0, 16),
	}
	fs := flag.NewFlagSet("spaghetti-cutter", flag.ExitOnError)
	fs.Var(&cfg.Allow, "allow", usageAllow)
	fs.Var(&cfg.Tool, "tool", usageTool)
	fs.Var(&cfg.DB, "db", usageDB)
	fs.Var(&cfg.God, "god", usageGod)
	fs.StringVar(&cfg.Root, "root", "", usageRoot)
	fs.UintVar(&cfg.Size, "size", defaultSize, usageSize)

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("SPAGHETTI_CUTTER"),
		ff.WithConfigFile(File),
		ff.WithConfigFileParser(ff.JSONParser),
	)
	if err != nil {
		log.Fatalf("FATAL - Unable to parse command line arguments or configuration file: %v", err)
	}

	//fmt.Println("Parsed config:", cfg)
	return cfg
}