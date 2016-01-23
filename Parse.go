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

package alexandria

import (
	"fmt"
	"strings"
)

// Figure out what type of a document we have
func DocumentType(m Metadata) string {
	return m.Type
}

// Return the lines of the last LaTeX comment block without the leading %.  In
// addition, leading and trailing whitespace is removed.
func findMetadataLines(doc string) []string {
	var metadata []string
	for _, line_ := range strings.Split(doc, "\n") {
		line := strings.TrimSpace(line_)
		if line == "" {
			continue
		}

		// If the current line does not begin with a LaTeX comment, we
		// have not yet encountered the last block of comments.
		if line[0] != '%' {
			metadata = make([]string, 0, 10)
			continue
		}
		// Remove any leading % and whitespaces
		line = strings.TrimLeft(line, "%  \t")
		if line != "" {
			metadata = append(metadata, line)
		}
	}
	return metadata
}

// Parse a comma separated list of tags into a slice, tranfsorming everything
// to lower case.
func parseTags(line string) []string {
	var tags []string
	for _, tag := range strings.Split(line, ",") {
		tmp := strings.ToLower(strings.TrimSpace(tag))
		if tmp != "" {
			tags = append(tags, tmp)
		}
	}
	return tags
}

// Parse the tags in the given scroll content.  The format of a scroll is
// generally of the following form:
//
//	\LaTeX\ code ...
//
//	% @source Author: Title
//	% @source Lemma 3.2, p. 41
//	% @type proposition
//	% counter-example, analysis, TopOloGY, Weierstraß
//
// In this example, the scroll contains a proposition and is tagged with
// 'counter-example', 'analysis', 'topology' and 'weierstraß'.  It can be found
// in Author: Title as Lemma 3.2 on pase 41.  All the metadata is stored in the
// final block of LaTeX comments.  Also, we simply ignore any empty lines.
func ParseMetadata(doc string) Metadata {
	// TODO Handle different types of tags: @source, @doctype, @keywords, and normal tags.
	var source []string
	var scroll_type string
	var tags []string

	for _, line := range findMetadataLines(doc) {
		switch {
		case strings.HasPrefix(line, "@source "):
			source = append(source, strings.TrimSpace(strings.TrimPrefix(line, "@source ")))
		case strings.HasPrefix(line, "@type "):
			tmp := strings.TrimSpace(strings.TrimPrefix(line, "@type "))
			if scroll_type != "" {
				panic(fmt.Sprintf("Scroll has two different types: %s and %s", scroll_type, tmp))
			} else {
				scroll_type = tmp
			}
		default:
			tags = append(tags, parseTags(line)...)
		}
	}

	// Handle the old format
	if scroll_type == "" {
		scroll_type = tags[0]
	}
	return Metadata{scroll_type, source, tags}
}
