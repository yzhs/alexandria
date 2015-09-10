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

	config.knowledgeDirectory = dir + "library/"
	config.cacheDirectory = dir + "cache/"
	config.templateDirectory = dir + "templates/"
	config.tempDirectory = dir + "tmp/"
	config.alexandriaDirectory = dir

	config.swishConfig = dir + "swish++.conf"
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
