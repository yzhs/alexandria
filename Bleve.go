// Alexandria
//
// Copyright (C) 2015-2016  Colin Benner
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
	"os"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

func touch(file string) error {
	now := time.Now()
	return os.Chtimes(file, now, now)
}

// Run index++ to generate a (new) swish++ index file.
func GenerateIndex() error {
	// TODO handle multiple fields, i.e. the main text, @source, @type, tags, etc.
	newIndex := false
	// Try to open an existing index or create a new one if none exists.
	index, err := openIndex()
	if err != nil {
		mapping := bleve.NewIndexMapping()
		mapping.DefaultAnalyzer = "en"
		index, err = bleve.New(Config.AlexandriaDirectory+"bleve", mapping)
		if err != nil {
			return err
		}
		newIndex = true
	}
	defer index.Close()

	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		return err
	}

	indexUpdateFile := Config.AlexandriaDirectory + "index_updated"
	indexUpdateTime, err := getModTime(indexUpdateFile)
	if err != nil {
		LogError(err)
		return nil
	}
	// Save the time of this indexing operation
	_ = touch(Config.AlexandriaDirectory + "index_updated")

	batch := index.NewBatch()
	for _, file := range files {
		// Check whether the scroll is newer than the index.
		modTime, err := getModTime(Config.KnowledgeDirectory + file.Name())
		if err != nil {
			LogError(err)
			continue
		}
		id := strings.TrimSuffix(file.Name(), ".tex")
		if modTime < indexUpdateTime && !newIndex {
			continue
		}

		// Load and parse the scroll content
		contentBytes, err := ioutil.ReadFile(Config.KnowledgeDirectory + file.Name())
		TryLogError(err)
		content := string(contentBytes)
		scroll := Parse(id, content)

		batch.Index(id, scroll)
	}
	index.Batch(batch)

	return nil
}

func RemoveFromIndex(id Id) error {
	index, err := openIndex()
	if err != nil {
		return err
	}
	defer index.Close()
	return index.Delete(string(id))
}

// Open the bleve index
func openIndex() (bleve.Index, error) {
	return bleve.Open(Config.AlexandriaDirectory + "bleve")
}

// Search the swish index for a given query.
func searchBleve(queryString string) (Results, error) {
	index, err := openIndex()
	if err != nil {
		LogError(err)
		return Results{}, err
	}
	defer index.Close()

	newQueryString := ""
	for _, tmp := range strings.Split(queryString, " ") {
		word := strings.TrimSpace(tmp)
		if word[0] == '-' || word[0] == '+' {
			newQueryString += " " + word
		} else if word[0] == '~' {
			// Remove prefix to make term optional
			newQueryString += " " + word[1:]
		} else {
			newQueryString += " +" + word
		}
	}

	query := bleve.NewQueryStringQuery(newQueryString[1:]) // Remove leading space
	search := bleve.NewSearchRequest(query)
	search.Size = Config.MaxResults
	searchResults, err := index.Search(search)
	if err != nil {
		println("Invalid query string: '" + newQueryString[1:] + "'")
		LogError(err)
		return Results{}, err
	}

	var ids []Scroll
	for _, match := range searchResults.Hits {
		id := Id(match.ID)
		content, err := readScroll(id)
		TryLogError(err)
		scroll := Parse(string(id), content)
		ids = append(ids, scroll)
	}

	return Results{ids[:len(searchResults.Hits)], int(searchResults.Total)}, nil
}

// Get a list of scrolls matching the query.
func FindScrolls(query string) (Results, error) {
	results, err := searchBleve(query)
	if err != nil {
		return Results{}, err
	}
	n := ProcessScrolls(results.Ids)
	ids := make([]Scroll, n)
	i := 0
	for _, id := range results.Ids {
		if _, err := os.Stat(Config.KnowledgeDirectory + string(id.Id) + ".tex"); os.IsNotExist(err) {
			continue
		}
		ids[i] = Scroll{Id: id.Id}
		i += 1
	}
	results.Total = n // The number of hits can be wrong if scrolls have been deleted

	return results, nil
}

func ComputeStatistics() Statistics {
	index, err := openIndex()
	if err != nil {
		LogError(err)
	}
	defer index.Close()

	num, size := getDirSize(Config.KnowledgeDirectory)
	if err == nil {
		tmp, err := index.DocCount()
		if err != nil {
			LogError(err)
		} else {
			num = int(tmp)
		}
	}

	return Stats{num, size}
}
