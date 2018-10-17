// Alexandria
//
// Copyright (C) 2015-2018  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

	alexandria.InitConfig()
	alexandria.Config.MaxResults = 1e9

	if profile {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	b := alexandria.Backend{alexandria.XelatexImagemagickRenderer{}}

	switch {
	case index:
		b.UpdateIndex()
	case stats:
		printStats(b)
	case version:
		fmt.Println(alexandria.NAME, alexandria.VERSION)
	case len(os.Args) <= 1:
		println("Nothing to do")
	case len(os.Args) == 2 && os.Args[1] == "all":
		numScrolls, errors := b.RenderAllScrolls()
		fmt.Printf("Rendered all %d scrolls.\n", numScrolls)
		if len(errors) != 0 {
			printErrors(errors)
			os.Exit(1)
		}

	default:
		i := 1
		if os.Args[1] == "--" {
			i += 1
		}
		renderMatchesForQuery(b, strings.Join(os.Args[i:], " "))
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
	ids, err := b.FindMatchingScrolls(query)
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
