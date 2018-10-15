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

	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
)

func touch(file string) error {
	now := time.Now()
	return os.Chtimes(file, now, now)
}

// Add all documents to the index that have been created or modified since the
// last time this function was executed.
//
// Note that this function does *not* remove deleted documents from the index.
// See `RemoveFromIndex`.
func UpdateIndex() error {
	index, isNewIndex, err := openOrCreateIndex()
	if err != nil {
		return err
	}
	defer index.Close()

	indexUpdateFile := Config.AlexandriaDirectory + "index_updated"
	indexUpdateTime, err := getModTime(indexUpdateFile)
	if err != nil {
		LogError(err)
		return nil
	}
	// Save the time of this indexing operation
	err = touch(indexUpdateFile)
	TryLogError(err)

	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		return err
	}

	batch := index.NewBatch()
	for _, file := range files {
		if !isNewIndex && isOlderThan(file, indexUpdateTime) {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".tex")
		scroll, err := loadAndParseScrollContent(id, file)
		if err == nil {
			batch.Index(id, scroll)
		}
	}
	index.Batch(batch)

	return nil
}

func openOrCreateIndex() (bleve.Index, bool, error) {
	isNewIndex := false

	index, err := openIndex()
	if err != nil {
		index, err = createNewIndex()
		isNewIndex = true
	}

	return index, isNewIndex, err
}

func openIndex() (bleve.Index, error) {
	return bleve.Open(Config.AlexandriaDirectory + "bleve")
}

func createNewIndex() (bleve.Index, error) {
	enTextMapping := bleve.NewTextFieldMapping()
	enTextMapping.Analyzer = "en"

	simpleMapping := bleve.NewTextFieldMapping()
	simpleMapping.Analyzer = simple.Name

	typeMapping := bleve.NewTextFieldMapping()
	typeMapping.Analyzer = keyword.Name

	scrollMapping := bleve.NewDocumentMapping()
	scrollMapping.AddFieldMappingsAt("id", simpleMapping)
	scrollMapping.AddFieldMappingsAt("content", enTextMapping)
	scrollMapping.AddFieldMappingsAt("type", typeMapping)
	scrollMapping.AddFieldMappingsAt("source", enTextMapping)
	scrollMapping.AddFieldMappingsAt("tag", enTextMapping)
	scrollMapping.AddFieldMappingsAt("hidden", enTextMapping)
	scrollMapping.AddFieldMappingsAt("other", enTextMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "en"
	mapping.DefaultMapping = scrollMapping

	return bleve.New(Config.AlexandriaDirectory+"bleve", mapping)
}

func isOlderThan(file os.FileInfo, indexUpdateTime int64) bool {
	modTime, err := getModTime(Config.KnowledgeDirectory + file.Name())
	if err != nil {
		LogError(err)
		return true
	} else {
		return modTime < indexUpdateTime
	}
}

func loadAndParseScrollContent(id string, file os.FileInfo) (Scroll, error) {
	contentBytes, err := ioutil.ReadFile(Config.KnowledgeDirectory + file.Name())
	TryLogError(err)
	content := string(contentBytes)
	scroll := Parse(id, content)
	return scroll, err
}

// Remove the specified document from the index. This is necessary as
// `UpdateIndex` has no way of knowing if a document was deleted.
func RemoveFromIndex(id Id) error {
	index, err := openIndex()
	if err != nil {
		return err
	}
	defer index.Close()
	return index.Delete(string(id))
}

// Get a list of scrolls matching the query.
func FindScrolls(query string) (Results, error) {
	results, err := searchBleve(query)
	if err != nil {
		return Results{}, err
	}
	var x XelatexImagemagickRenderer
	n := RenderListOfScrolls(results.Ids, x)
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
