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

package alexandria

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

type Backend struct {
	renderer LatexToPngRenderer
}

func NewBackend() Backend {
	return Backend{renderer: XelatexImagemagickRenderer{}}
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
		err := RenderScroll(id, b.renderer)
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
	return UpdateIndex()
}

func (b *Backend) Statistics() (Statistics, error) {
	return ComputeStatistics()
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
