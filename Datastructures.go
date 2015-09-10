package main

// Configuration data of Alexandria
type Config struct {
	quality int
	dpi     int

	alexandriaDirectory string
	knowledgeDirectory  string
	cacheDirectory      string
	templateDirectory   string
	tempDirectory       string

	swishConfig string
}

var config Config

// The metadata contained in a scroll
type Metadata struct {
	source string
	tags   []string
}
