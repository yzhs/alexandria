package main

import (
	"fmt"
	"os"
	"runtime/pprof"
)

func usage() {
	os.Exit(1)
}

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
	if len(os.Args) <= 1 {
		usage()
	}
	initConfig()

	if len(os.Args) == 2 && os.Args[1] == "-i" {
		generateIndex()
	} else if os.Args[1] == "-I" {
		printStats()
	} else {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

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
