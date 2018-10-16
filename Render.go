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

type RenderBackend interface {
	// Create a LaTeX file from the content of the given scroll together
	// with all the appropriate templates.  The resulting file stored in
	// the temp directory.
	scrollToLatex(id ID)

	// Compile a LaTeX file with the given id to produce a PDF file.  Both
	// input and output files are in the temp directory.
	latexToPdf(id ID)

	// Convert PDF to PNG, storing the result in the cache directory.  From
	// there, it can be served by the web server or displayed to the user
	// via some other user interface.
	pdfToPng(id ID)

	deleteTemporaryFiles(id ID)

	err() error
	message() string
}

type XelatexImagemagickRenderer struct {
	error error
	msg   string
}

func (x XelatexImagemagickRenderer) scrollToLatex(id ID) {
	var e errTemplateReader

	scrollText, err := readScroll(id)
	if err != nil {
		if os.IsNotExist(err) {
			err = RemoveFromIndex(id)
			if err != nil {
				LogError(err)
			}
			x.error = NoSuchScrollError
			return
		}
		LogError(err)
		x.error = err
		return
	}
	scroll := Parse(string(id), scrollText)

	e.readTemplate("header")
	e.readTemplate(scroll.DocumentType() + "_header")
	e.doc += scroll.Content
	e.readTemplate(scroll.DocumentType() + "_footer")
	e.readTemplate("footer")

	if e.err != nil {
		x.error = e.err
		return
	}
	x.error = writeTemp(id, e.doc)
}

func (x XelatexImagemagickRenderer) latexToPdf(id ID) {
	if x.error != nil {
		return
	}

	msg, err := exec.Command("xelatex", "-interaction", "nonstopmode",
		"-output-directory", Config.TempDirectory,
		Config.TempDirectory+string(id)).CombinedOutput()
	x.error = err
	if err != nil {
		x.msg = string(msg)
	}
}

func (x XelatexImagemagickRenderer) pdfToPng(i ID) {
	if x.error != nil {
		return
	}

	id := string(i)
	x.error = exec.Command("convert", "-trim",
		"-quality", strconv.Itoa(Config.Quality),
		"-density", strconv.Itoa(Config.Dpi),
		Config.TempDirectory+id+".pdf", Config.CacheDirectory+id+".png").Run()

}

func (x XelatexImagemagickRenderer) deleteTemporaryFiles(id ID) {
}

func (x XelatexImagemagickRenderer) err() error {
	return x.error
}

func (x XelatexImagemagickRenderer) message() string {
	return x.msg
}

// Generate a PNG image from a given scroll, if there is no up-to-date image.
func renderScroll(id ID, renderer RenderBackend) error {
	if isUpToDate(id) {
		return nil
	}

	renderer.scrollToLatex(id)
	renderer.latexToPdf(id)
	renderer.pdfToPng(id)
	renderer.deleteTemporaryFiles(id)

	err := renderer.err()
	if err != nil {
		log.Panic("Error: ", err, "\n", renderer.message())
	}
	return err
}

func RenderListOfScrolls(ids []Scroll, renderer RenderBackend) int {
	numScrolls := 0

	for _, foo := range ids {
		id := foo.ID
		err := renderScroll(id, renderer)
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

func RenderAllScrolls(renderer RenderBackend) int {
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
			id := ID(strings.TrimSuffix(file.Name(), ".tex"))
			if err := renderScroll(id, renderer); err != nil && err != NoSuchScrollError {
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
