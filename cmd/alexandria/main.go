// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"

	flag "github.com/ogier/pflag"

	"github.com/yzhs/alexandria"
)

func main() {
	var index, profile, stats, version bool
	flag.BoolVarP(&index, "index", "i", false, "\tUpdate the index")
	flag.BoolVarP(&stats, "stats", "S", false, "\tPrint some statistics")
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	alexandria.Config.MaxResults = 1e9
	if profile {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	b := alexandria.NewBackend()

	switch {
	case index:
		b.UpdateIndex()
	case stats:
		printStats(b)
	case version:
		fmt.Println(alexandria.NAME, alexandria.VERSION)
	case len(os.Args) <= 1:
		fmt.Fprintln(os.Stderr, "Nothing to do")
	case len(os.Args) == 2 && os.Args[1] == "all":
		renderEverything(b)
	default:
		i := 1
		if os.Args[1] == "--" {
			i++
		}
		renderMatchesForQuery(b, strings.Join(os.Args[i:], " "))
	}
}

func renderEverything(b alexandria.Backend) {
	numScrolls, errors := b.RenderAllScrolls()
	fmt.Printf("Rendered all %d scrolls.\n", numScrolls)
	if len(errors) != 0 {
		printErrors(errors)
		os.Exit(1)
	}
}

func printStats(b alexandria.Backend) {
	stats, err := b.Statistics()
	if err != nil {
		panic(err)
	}
	n := stats.NumberOfScrolls()
	size := float32(stats.TotalSize()) / 1024.0
	fmt.Printf("The library contains %v scrolls with a total size of %.1f kiB.\n", n, size)
}

func renderMatchesForQuery(b alexandria.Backend, query string) {
	ids, _, err := b.FindMatchingScrolls(query)
	if err != nil {
		panic(err)
	}
	renderedIDs, errors := b.RenderScrollsByID(ids)
	fmt.Printf("There are %d matching scrolls.\n", len(renderedIDs))
	for _, id := range renderedIDs {
		fmt.Println("file://" + alexandria.Config.CacheDirectory + string(id) + ".png")
	}
	if len(errors) != 0 {
		printErrors(errors)
		os.Exit(1)
	}
}

func printErrors(errors []error) {
	fmt.Fprintf(os.Stderr, "The following errors occurred:\n")
	for _, err := range errors {
		fmt.Fprintln(os.Stderr, err)
	}
}
