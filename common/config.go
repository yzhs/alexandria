// This, together with lib.go, forms the public interface of Alexandria (the
// library).

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package common

import (
	"os"
)

// Config holds all the configuration of Alexandria.
var Config = initConfig()

func initConfig() Configuration {
	var config Configuration

	config.Quality = 90
	config.Dpi = 160
	config.MaxResults = 1000
	config.MaxProcs = 4

	dir := os.Getenv("HOME") + "/.alexandria/"

	config.AlexandriaDirectory = dir
	config.KnowledgeDirectory = dir + "library/"
	config.CacheDirectory = dir + "cache/"
	config.TemplateDirectory = dir + "templates/"
	config.TempDirectory = dir + "tmp/"

	return config
}
