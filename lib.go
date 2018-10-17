// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Backend struct {
	renderer latexToPngRenderer
}

func NewBackend() Backend {
	return Backend{renderer: xelatexImagemagickRenderer{}}
}

func (b *Backend) RenderAllScrolls() (numScrolls int, errors []error) {
	ids := b.IDsForAllScrolls()
	renderedIDs, errors := b.RenderScrollsByID(ids)
	return len(renderedIDs), errors
}

func (b *Backend) RenderScrollsByID(ids []ID) (renderedScrollIDs []ID, errors []error) {
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

func (b *Backend) FindMatchingScrolls(query string) ([]ID, int, error) {
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
func (b *Backend) IDsForAllScrolls() []ID {
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

func (b *Backend) UpdateIndex() error {
	return updateIndex()
}

func (b *Backend) Statistics() (Statistics, error) {
	return computeStatistics()
}

func (b *Backend) LoadScrolls(ids []ID) ([]Scroll, error) {
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
var Config Configuration

// InitConfig initializes Config with reasonable default values.
func InitConfig() {
	Config.Quality = 90
	Config.Dpi = 160
	Config.MaxResults = 1000
	Config.MaxProcs = 4

	dir := os.Getenv("HOME") + "/.alexandria/"

	Config.AlexandriaDirectory = dir
	Config.KnowledgeDirectory = dir + "library/"
	Config.CacheDirectory = dir + "cache/"
	Config.TemplateDirectory = dir + "templates/"
	Config.TempDirectory = dir + "tmp/"
}
