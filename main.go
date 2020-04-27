package main

import (
	"log"
	"os"

	"github.com/flowdev/spaghetti-cutter/deps"
	"github.com/flowdev/spaghetti-cutter/dirs"
	"github.com/flowdev/spaghetti-cutter/parse"
	"github.com/flowdev/spaghetti-cutter/size"
	"github.com/flowdev/spaghetti-cutter/x/config"
)

func main() {
	var err error

	cfg := config.Parse(os.Args[1:])
	cfg.God["main"] = config.Value // the main package can always access everything

	cfg.Root, err = dirs.FindRoot(cfg.Root)
	if err != nil {
		log.Printf("FATAL -  %v", err)
		os.Exit(2)
	}
	log.Printf("INFO - configuration God: %#v", cfg.God)
	log.Printf("INFO - configuration Tool: %#v", cfg.Tool)
	log.Printf("INFO - configuration DB: %#v", cfg.DB)
	log.Printf("INFO - configuration Allow: %#v", cfg.Allow)
	log.Printf("INFO - configuration Size: %d", cfg.Size)
	log.Printf("INFO - configuration Root: %s", cfg.Root)

	pkgs, err := parse.DirTree(cfg.Root)
	if err != nil {
		log.Printf("FATAL - %v", err)
		os.Exit(3)
	}

	var errs []error
	rootPkg := parse.RootPkg(pkgs)
	log.Printf("INFO - root package: %s", rootPkg)
	for _, pkg := range pkgs {
		errs = addErrors(errs, deps.Check(pkg, rootPkg, cfg))
		errs = addErrors(errs, size.Check(pkg, rootPkg, cfg.Size))
	}

	if len(errs) > 0 {
		for _, err = range errs {
			log.Printf("ERROR - %v", err)
		}
		os.Exit(1)
	}
}

func addErrors(errs []error, newErrs []error) []error {
	return append(errs, newErrs...)
}
