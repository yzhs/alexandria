// Alexandria
//
// Copyright (C) 2015  Colin Benner
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
	"io/ioutil"
	"strings"

	"github.com/blevesearch/bleve"
)

// Run index++ to generate a (new) swish++ index file.
func GenerateIndex() error {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(Config.AlexandriaDirectory+"Alexandria.bleve", mapping)
	if err != nil {
		return err
	}
	defer index.Close()

	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		return err
	}

	for _, f := range files {
		id := strings.TrimSuffix(f.Name(), ".tex")
		contentBytes, err := ioutil.ReadFile(Config.KnowledgeDirectory + f.Name())
		TryLogError(err)
		content := string(contentBytes)
		metadata := ParseMetadata(content)
		scroll := Scroll{metadata, Id(id), content}

		index.Index(id, scroll)
	}

	return nil
}

// Search the swish index for a given query.
func searchBleve(queryString string) ([]Id, error) {
	index, err := bleve.Open(Config.AlexandriaDirectory + "Alexandria.bleve")
	if err != nil {
		LogError(err)
		return nil, err
	}
	defer index.Close()

	query := bleve.NewQueryStringQuery(queryString)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		LogError(err)
		return nil, err
	}

	results := make([]Id, Config.MaxResults)
	for i, match := range searchResults.Hits {
		results[i] = Id(match.ID)
	}

	return results[:len(searchResults.Hits)], nil
}

// Get a list of scrolls matching the query.
func FindScrolls(query string) ([]Id, error) {
	ids, err := searchBleve(query)
	if err != nil {
		return nil, err
	}
	ProcessScrolls(ids)

	return ids, nil
}
