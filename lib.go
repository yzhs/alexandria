// This file contains most of the public interface of Alexandria (the library).
// The remainder is the types defined in datastructures.go.

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

//go:generate go run contrib/generate_assets.go

package alexandria

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Programm name and version
const (
	NAME    = "Alexandria"
	VERSION = "0.1"
)

type scrollType int

const (
	latexScroll scrollType = iota
	markdownScroll
)

type Backend interface {
	// RenderAllScrolls does what it says on the tin: it pre-renders all
	// scrolls, returning the number of scrolls rendered and an array of
	// all the errors that occurred in the process.
	RenderAllScrolls() (numScrolls int, errors []error)

	// RenderScrollsById renders all the scrolls specified, returning the
	// IDs of all scrolls that were successfully rendered, and a list of
	// the errors that were encountered in the process.
	RenderScrollsByID(ids []ID) (renderedScrollIDs []ID, errors []error)
}

func NewBackend() Backend {
	return LatexToPngBackend{renderer: &xelatexImagemagickRenderer{}}
}

// FindMatchingScrolls asks the storage backend for all scrolls matching the
// given query. It the IDs of the first of those scrolls, the total number of
// matches (which can be much greater than the number of IDs returned), and an
// error, if any occurred.
func FindMatchingScrolls(query string) ([]ID, int, error) {
	index, err := openExistingIndex()
	if err != nil {
		return []ID{}, 0, err
	}
	defer index.Close()

	newQuery := translatePlusMinusTildePrefixes(query)
	searchResults, err := performQuery(index, newQuery)
	totalMatches := int(searchResults.Total)
	if err != nil {
		if err.Error() == "syntax error" {
			err = errors.Wrapf(err, "invalid query string: '%v'", newQuery)
		} else {
			err = errors.Wrap(err, "perform query")
		}
		return []ID{}, totalMatches, err
	}

	var ids []ID
	for _, match := range searchResults.Hits {
		id := ID(match.ID)
		ids = append(ids, id)
	}

	return ids, totalMatches, nil
}

// IDsForAllScrolls goes through the library directory and creates a list of
// IDs of all available scrolls.
func (LatexToPngBackend) IDsForAllScrolls() []ID {
	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		panic(err)
	}

	var ids []ID
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".tex") {
			continue
		}
		id := ID(strings.TrimSuffix(file.Name(), ".tex"))
		ids = append(ids, id)
	}
	return ids
}

func UpdateIndex() error {
	return updateIndex()
}

func ComputeStatistics() (Statistics, error) {
	return computeStatistics()
}

func LoadScrolls(ids []ID) ([]Scroll, error) {
	result := make([]Scroll, len(ids))
	for i, id := range ids {
		scroll, err := loadAndParseScrollContentByID(id)
		if err != nil {
			return result, err
		}
		result[i] = scroll
	}
	return result, nil
}

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
