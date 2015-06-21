package main

import (
	"strings"
)

// Figure out what type of a document we have
func documentType(m Metadata) string {
	return m.tags[0]
}

func parseTags(doc string) Metadata {
	var source string
	var tags []string
	var found_source bool

	for _, line_ := range strings.Split(doc, "\n") {
		line := strings.TrimSpace(strings.TrimPrefix(line_, "%"))
		if line == "" {
			continue
		}

		// Ignore all but the last block of comments
		if line[0] != '%' {
			tags = []string{}
			source = ""
			found_source = false
		}

		if !found_source && strings.HasPrefix(line, "@source") {
			source = line
			found_source = true
		} else {
			tags_temp := strings.Split(line, ",")
			for i := range tags_temp {
				tags_temp[i] = strings.TrimSpace(tags_temp[i])
			}
			tags = append(tags, tags_temp...)
		}
	}
	return Metadata{source, tags}
}
