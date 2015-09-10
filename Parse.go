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
