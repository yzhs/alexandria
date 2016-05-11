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
		// have not yet encountered the last block of comments, so we
		// should discard all data gathered so far.
		if line[0] != '%' {
			metadata = make([]string, 0, 100) // FIXME magic constant
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
//	% @type proposition, definition
//	% counter-example, analysis, TopOloGY, Weierstraß
//
// In this example, the scroll contains a proposition and is tagged with
// 'counter-example', 'analysis', 'topology' and 'weierstraß'.  It can be found
// in Author: Title as Lemma 3.2 on pase 41.  All the metadata is stored in the
// final block of LaTeX comments.  Also, we simply ignore any empty lines.
func ParseMetadata(doc string) Metadata {
	// TODO Handle different types of tags: @source, @doctype, @keywords, and normal tags.
	var source []string
	var hidden []string
	var scroll_type string
	var tags []string
	var other_lines []string

	for _, line := range findMetadataLines(doc) {
		switch {
		case strings.HasPrefix(line, "@hidden "):
			hidden = append(hidden, parseTags(strings.TrimPrefix(line, "@hidden "))...)
		case strings.HasPrefix(line, "@source "):
			source = append(source, strings.TrimSpace(strings.TrimPrefix(line, "@source ")))
		case strings.HasPrefix(line, "@type "):
			tmp := strings.TrimSpace(strings.TrimPrefix(line, "@type "))
			for _, type_ := range strings.Split(tmp, ",") {
				// Ignore all but the first type, the other ones are just for searching
				if scroll_type == "" {
					scroll_type = strings.TrimSpace(type_)
					break
				}
			}
		case strings.HasPrefix(line, "@"):
			// Do not strip the @[a-zA-Z0-9\-]** prefix, otherwise
			// there is no way to tell what the line signifies.
			other_lines = append(other_lines, line)
		default:
			tags = append(tags, parseTags(line)...)
		}
	}

	return Metadata{Type: scroll_type, SourceLines: source, Tags: tags,
		Hidden: hidden, OtherLines: other_lines}
}

func StripComments(doc string) string {
	var content string
	for _, line_ := range strings.Split(doc, "\n\n") {
		line := strings.TrimSpace(line_)
		if len(line) > 0 && line[0] == '%' {
			continue
		}
		content += line + "\n"
	}
	return strings.TrimSpace(content)
}
