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

import (
	"log"
	"os/exec"
	"strconv"
)

type errTemplateReader struct {
	doc string
	err error
}

func (e *errTemplateReader) readTemplate(name string) {
	if e.err != nil {
		return
	}

	tmp, err := readTemplate(name)
	e.err = err
	e.doc += tmp
}

// Create a LaTeX file from the content of the given scroll together with all
// the appropriate templates.  The resulting file stored in the temp directory.
func scrollToLatex(id Id) error {
	var e errTemplateReader

	scroll, err := readScroll(id)
	if err != nil {
		log.Fatal(err)
		return err
	}
	tags := ParseMetadata(scroll)

	e.readTemplate("header")
	e.readTemplate(DocumentType(tags) + "_header")
	e.doc += scroll
	e.readTemplate(DocumentType(tags) + "_footer")
	e.readTemplate("footer")

	if e.err != nil {
		return e.err
	}
	return writeTemp(id, e.doc)
}

// Compile a LaTeX file with the given id to produce a PDF file.  Both input
// and output files are in the temp directory.
func latexToPdf(id Id) error {
	msg, err := exec.Command("xelatex", "-interaction", "nonstopmode",
		"-output-directory", Config.TempDirectory,
		Config.TempDirectory+string(id)).CombinedOutput()
	if err != nil {
		log.Fatal(string(msg))
		return err
	}
	return nil
}

// Convert PDF to PNG, storing the result in the cache directory.  From there,
// it can be served by the web server or displayed to the user via some other
// user interface.
func pdfToPng(i Id) error {
	id := string(i)
	return exec.Command("convert", "-trim",
		"-quality", strconv.Itoa(Config.Quality),
		"-density", strconv.Itoa(Config.Dpi),
		Config.TempDirectory+id+".pdf", Config.CacheDirectory+id+".png").Run()
}

// Generate a PNG image from a given scroll, if there is no up-to-date image.
func ProcessScroll(id Id) error {
	if isUpToDate(id) {
		return nil
	}

	err := scrollToLatex(id)
	if err != nil {
		return err
	}

	err = latexToPdf(id)
	if err != nil {
		return err
	}

	return pdfToPng(id)
}

// Generate images for a list of scrolls.
func ProcessScrolls(ids []Id) {
	for _, id := range ids {
		err := ProcessScroll(id)
		if err != nil {
			log.Panic("An error ocurred when processing scroll ", id, ": ", err)
		}
	}
}
