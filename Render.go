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
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var NoSuchScrollError = errors.New("No such scroll")

const hashes = "############################################################"

type errTemplateReader struct {
	doc string
	err error
}

// Load a template file from disk and propagate errors
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

	scrollText, err := readScroll(id)
	if err != nil {
		if os.IsNotExist(err) {
			err = RemoveFromIndex(id)
			if err != nil {
				LogError(err)
			}
			return NoSuchScrollError
		}
		LogError(err)
		return err
	}
	scroll := Parse(string(id), scrollText)

	e.readTemplate("header")
	e.readTemplate(scroll.DocumentType() + "_header")
	e.doc += scroll.Content
	e.readTemplate(scroll.DocumentType() + "_footer")
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
	}
	return err
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
func renderScroll(id Id) error {
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

func RenderListOfScrolls(ids []Scroll) int {
	numScrolls := 0

	for _, foo := range ids {
		id := foo.Id
		err := renderScroll(id)
		if err != nil {
			if err == NoSuchScrollError {
				continue
			} else {
				log.Panic("An error ocurred when processing scroll ", id, ": ", err)
			}
		} else {
			numScrolls += 1
		}
	}

	return numScrolls
}

func RenderAllScrolls() int {
	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		panic(err)
	}
	var errors []error
	limitGoroutines := make(chan bool, Config.MaxProcs)
	for i := 0; i < Config.MaxProcs; i++ {
		limitGoroutines <- true
	}
	ch := make(chan int, len(files))
	for _, file := range files {
		go func(file os.FileInfo) {
			<-limitGoroutines
			if !strings.HasSuffix(file.Name(), ".tex") {
				ch <- 0
				return
			}
			id := Id(strings.TrimSuffix(file.Name(), ".tex"))
			if err := renderScroll(id); err != nil && err != NoSuchScrollError {
				log.Printf("%s\nERROR\n%s\n%v\n%s\n", hashes, hashes, err, hashes)
			}
			ch <- 1
		}(file)
	}
	counter := 0
	for i := 0; i < len(files); i++ {
		counter += <-ch
		limitGoroutines <- true
	}
	for _, err := range errors {
		log.Printf("Error: %v\n", err)
	}
	return counter
}
