// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"strings"
)

// Return the lines of the last LaTeX comment block without the leading %.  In
// addition, leading and trailing whitespace is removed.
func findMetadataLines(doc string) []string {
	var metadata []string
	for _, line := range strings.Split(doc, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// If the current line does not begin with a LaTeX comment, we
		// have not yet encountered the last block of comments, so we
		// should discard all data gathered so far.
		if trimmedLine[0] != '%' {
			metadata = make([]string, 0, 100) // FIXME magic constant
			continue
		}
		// Remove any leading % and whitespaces
		trimmedLine = strings.TrimLeft(trimmedLine, "%  \t")
		if trimmedLine != "" {
			metadata = append(metadata, trimmedLine)
		}
	}
	return metadata
}

// Parse a comma separated list of tags into a slice, tranfsorming everything
// to lower case.
func parseTags(line string) []string {
	var tags []string
	for _, tag := range strings.Split(line, ",") {
		tmp := strings.TrimSpace(tag)
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
func parse(id, doc string) Scroll {
	// TODO Handle different types of tags: @source, @doctype, @keywords, and normal tags.
	var source []string
	var hidden []string
	var scrollType string
	var tags []string
	var otherLines []string

	for _, line := range findMetadataLines(doc) {
		switch {
		case strings.HasPrefix(line, "@hidden "):
			hidden = append(hidden, parseTags(strings.TrimPrefix(line, "@hidden "))...)
		case strings.HasPrefix(line, "@source "):
			source = append(source, strings.TrimSpace(strings.TrimPrefix(line, "@source ")))
		case strings.HasPrefix(line, "@type "):
			tmp := strings.TrimSpace(strings.TrimPrefix(line, "@type "))
			for _, typ := range strings.Split(tmp, ",") {
				// Ignore all but the first type, the other
				// ones are just for searching
				if scrollType == "" {
					scrollType = strings.TrimSpace(typ)
					break
				}
			}
		case strings.HasPrefix(line, "@"):
			// Do not strip the @[a-zA-Z0-9\-]** prefix, otherwise
			// there is no way to tell what the line signifies.
			otherLines = append(otherLines, line)
		default:
			tags = append(tags, parseTags(line)...)
		}
	}
	content := stripComments(doc)

	return Scroll{ID: ID(id), Content: content, Type: scrollType,
		SourceLines: source, Tags: tags, Hidden: hidden,
		OtherLines: otherLines}
}

// Remove all lines that only contain a LaTeX comment.  This removes all the
// medatata from a scroll.
func stripComments(doc string) string {
	var content string
	for _, line := range strings.Split(doc, "\n\n") {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) > 0 && trimmedLine[0] == '%' {
			continue
		}
		content += trimmedLine + "\n\n"
	}
	return strings.TrimSpace(content)
}
