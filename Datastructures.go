package main

// Configuration data of Alexandria
type Config struct {
	quality int
	dpi     int

	knowledgeDirectory  string
	cacheDirectory      string
	templateDirectory   string
	tempDirectory       string
	alexandriaDirectory string

	swishConfig string
}

// The metadata contained in a scroll
type Metadata struct {
	source string
	tags   []string
}
