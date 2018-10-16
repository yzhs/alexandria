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

import "os"

// Programm name and version
const (
	NAME    = "Alexandria"
	VERSION = "0.1"
)

// Stats contains the size of the library, in number of scrolls and in terms of
// file size.
type Stats struct {
	numScrolls int
	fileSize   int64
}

// NumberOfScrolls returns the number of scrolls in the library.
func (s Stats) NumberOfScrolls() int {
	return s.numScrolls
}

// TotalSize returns the total size of the files in the library.
func (s Stats) TotalSize() int64 {
	return s.fileSize
}

// ID holds UUID identifying a scroll.
type ID string

// Backend provides access to a full-text search system.
type Backend interface {
	Search(query []string) ([]ID, error)
	GenerateIndex() error
	ComputeStatistics() Statistics
}

// Renderer allows you to render a scroll of a certain file type.
type Renderer interface {
	Extension() string
	Render(id ID) error
}

// Statistics computes data about how large your library is.
type Statistics interface {
	NumberOfScrolls() int
	TotalSize() int64
}

// Configuration data of Alexandria
type Configuration struct {
	// The setting passed to ImageMagick when generating the PNG files
	Quality int
	// The resolution of the generated PNG files
	Dpi int
	// How many processes may run in parallel when rendering
	MaxProcs int

	// How many results are to be processed at once
	MaxResults int

	AlexandriaDirectory string
	KnowledgeDirectory  string
	CacheDirectory      string
	TemplateDirectory   string
	TempDirectory       string
}

// Scroll contains all the data contained in a document.
type Scroll struct {
	ID      ID     `json:"id"`
	Content string `json:"content"`
	// Type is the type of document we are dealing with.  This might be
	// something 'definition', 'lemma', etc.  It is used to select the
	// appropriate template when rendering.
	Type        string   `json:"type"`
	SourceLines []string `json:"source"`
	Tags        []string `json:"tag"`
	Hidden      []string `json:"hidden"`
	OtherLines  []string `json:"other"`
}

// Results holds both the IDs of the first n scrolls matching a query, and the
// number of matches.
type Results struct {
	Ids []Scroll
	// How many results there were all in all; can be significantly larger than len(Ids).
	Total int
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
