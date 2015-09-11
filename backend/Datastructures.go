// Alexandria
//
// Copyright (C) 2015  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package backend

import "os"

type Statistics struct {
	Num  int
	Size int64
}

// Configuration data of Alexandria
type Configuration struct {
	Quality int
	Dpi     int

	AlexandriaDirectory string
	KnowledgeDirectory  string
	CacheDirectory      string
	TemplateDirectory   string
	TempDirectory       string

	SwishConfig string
}

// The metadata contained in a scroll
type Metadata struct {
	source string
	tags   []string
}

var Config Configuration

func InitConfig() {
	Config.Quality = 90
	Config.Dpi = 137

	dir := os.Getenv("HOME") + "/.alexandria/"

	Config.AlexandriaDirectory = dir
	Config.KnowledgeDirectory = dir + "library/"
	Config.CacheDirectory = dir + "cache/"
	Config.TemplateDirectory = dir + "templates/"
	Config.TempDirectory = dir + "tmp/"

	Config.SwishConfig = dir + "swish++.conf"
}
