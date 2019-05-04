// This file contains most of the public interface of Alexandria (the library).
// The remainder is the types defined in datastructures.go.

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

//go:generate go run contrib/generate_assets.go

package alexandria

import (
	"fmt"
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

type Backend interface {
	RenderAllScrolls() (numScrolls int, errors []error)

	RenderScrollsByID(ids []ID) (renderedScrollIDs []ID, errors []error)

	FindMatchingScrolls(query string) ([]ID, int, error)

	// IDsForAllScrolls goes through the library directory and creates a list of
	// IDs of all available scrolls.
	IDsForAllScrolls() []ID

	UpdateIndex() error

	Statistics() (Statistics, error)

	LoadScrolls(ids []ID) ([]Scroll, error)
}

type LatexToPngBackend struct {
	renderer latexToPngRenderer
}

type scrollType int

const (
	latexScroll scrollType = iota
	markdownScroll
)

func NewBackend() Backend {
	return LatexToPngBackend{renderer: &xelatexImagemagickRenderer{}}
}

func (b LatexToPngBackend) RenderAllScrolls() (numScrolls int, errors []error) {
	ids := b.IDsForAllScrolls()
	renderedIDs, errors := b.RenderScrollsByID(ids)
	return len(renderedIDs), errors
}

func (b LatexToPngBackend) RenderScrollsByID(ids []ID) (renderedScrollIDs []ID, errors []error) {
	renderedScrollIDs = make([]ID, len(ids))
	numScrolls := 0
	for _, id := range ids {
		err := renderScroll(id, b.renderer)
		if err != nil {
			errors = append(errors, fmt.Errorf("rendering %v failed", id))
		} else {
			renderedScrollIDs[numScrolls] = id
			numScrolls++
		}
	}

	return renderedScrollIDs[:numScrolls], errors
}

func (b LatexToPngBackend) FindMatchingScrolls(query string) ([]ID, int, error) {
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
func (b LatexToPngBackend) IDsForAllScrolls() []ID {
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

func (b LatexToPngBackend) UpdateIndex() error {
	return updateIndex()
}

func (b LatexToPngBackend) Statistics() (Statistics, error) {
	return computeStatistics()
}

func (b LatexToPngBackend) LoadScrolls(ids []ID) ([]Scroll, error) {
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
