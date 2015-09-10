// Alexandria
//
// Copyright (C) 2015  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	flag "github.com/ogier/pflag"
)

const (
	NAME    = "Alexandria"
	VERSION = "0.1"
)

func initConfig() {
	config.quality = 90
	config.dpi = 137

	dir := os.Getenv("HOME") + "/.alexandria/"

	config.alexandriaDirectory = dir
	config.knowledgeDirectory = dir + "library/"
	config.cacheDirectory = dir + "cache/"
	config.templateDirectory = dir + "templates/"
	config.tempDirectory = dir + "tmp/"

	config.swishConfig = dir + "swish++.conf"
}

func printStats() {
	n, size := getDirSize(config.knowledgeDirectory)
	fmt.Printf("The library contains %v scrolls with a total size of %.1f kiB.\n", n, float32(size)/1024.0)
}

func main() {
	var index, profile, stats, version bool
	flag.BoolVarP(&index, "index", "i", false, "\tUpdate the index")
	flag.BoolVarP(&stats, "stats", "S", false, "\tPrint some statistics")
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	initConfig()

	if profile {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	switch {
	case index:
		generateIndex()
	case stats:
		printStats()
	case version:
		fmt.Println(NAME, VERSION)
	default:
		num, ids, err := findScrolls(os.Args[1:])
		if err != nil {
			panic(err)
		}
		fmt.Printf("There are %d matching scrolls.\n", num)
		for _, id := range ids {
			fmt.Println("file://" + config.cacheDirectory + id + ".png")
		}
	}
}
