// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

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
