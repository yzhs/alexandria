// This, together with lib.go, forms the public interface of Alexandria (the
// library).

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package common

import (
	"github.com/pkg/errors"
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

	Parse(id, doc string) Scroll
}

// FindMatchingScrolls asks the storage backend for all scrolls matching the
// given query. It the IDs of the first of those scrolls, the total number of
// matches (which can be much greater than the number of IDs returned), and an
// error, if any occurred.
func FindMatchingScrolls(query string) ([]ID, int, error) {
	index, err := OpenExistingIndex()
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

func UpdateIndex(b Backend) error {
	return updateIndex(b)
}

func ComputeStatistics() (Statistics, error) {
	return computeStatistics()
}

func LoadScrolls(b Backend, ids []ID) ([]Scroll, error) {
	result := make([]Scroll, len(ids))
	for i, id := range ids {
		scroll, err := loadAndParseScrollContentByID(b, id)
		if err != nil {
			return result, err
		}
		result[i] = scroll
	}
	return result, nil
}
