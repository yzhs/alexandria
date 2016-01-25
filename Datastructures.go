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

import "os"

const (
	NAME    = "Alexandria"
	VERSION = "0.1"
)

// Statistics concerning the size of the library.
type Stats struct {
	num_scrolls int
	file_size   int64
}

func (s Stats) Num() int {
	return s.num_scrolls
}

func (s Stats) Size() int64 {
	return s.file_size
}

type Id string

type Backend interface {
	Search(query []string) ([]Id, error)
	GenerateIndex() error
	ComputeStatistics() Statistics
}

type Renderer interface {
	Extension() string
	Render(id Id) error
}

type Statistics interface {
	Num() int
	Size() int64
}

// Configuration data of Alexandria
type Configuration struct {
	Quality    int
	Dpi        int
	MaxResults int

	AlexandriaDirectory string
	KnowledgeDirectory  string
	CacheDirectory      string
	TemplateDirectory   string
	TempDirectory       string

	// TODO move code specific to the swish backend elsewhere
	SwishConfig string
}

// The metadata contained in a scroll
type Metadata struct {
	Type   string
	Source []string
	Tags   []string
}

var Config Configuration

func InitConfig() {
	Config.Quality = 90
	Config.Dpi = 137
	Config.MaxResults = 1000

	dir := os.Getenv("HOME") + "/.alexandria/"

	Config.AlexandriaDirectory = dir
	Config.KnowledgeDirectory = dir + "library/"
	Config.CacheDirectory = dir + "cache/"
	Config.TemplateDirectory = dir + "templates/"
	Config.TempDirectory = dir + "tmp/"

	Config.SwishConfig = dir + "swish++.conf"
}
