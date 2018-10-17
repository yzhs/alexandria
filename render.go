// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

// ErrNoSuchScroll is used when a query returns a scroll ID that is no longer
// in the library, i.e. has been deleted but RemoveFromIndex has not yet been
// called on it.
var ErrNoSuchScroll = errors.New("No such scroll")

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
	e.err = errors.Wrapf(err, "read template %v", name)
	e.doc += tmp
}

// latexToPngRenderer describes a LaTeX->PDF->PNG pipeline.
type latexToPngRenderer interface {
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
}

// xelatexImagemagickRenderer uses xelatex to handle the LaTeX-to-PDF
// translation, ImageMagick to convert the PDF to a PNG.
type xelatexImagemagickRenderer struct {
	error error
}

func (x *xelatexImagemagickRenderer) scrollToLatex(id ID) {
	var e errTemplateReader

	scrollText, err := readScroll(id)
	if err != nil {
		if os.IsNotExist(err) {
			err = removeFromIndex(id)
			if err != nil {
				logError(err)
			}
			x.error = ErrNoSuchScroll
			return
		}
		x.error = err
		return
	}
	scroll := parse(string(id), scrollText)

	e.readTemplate("header")
	e.readTemplate(scroll.Type + "_header")
	e.doc += scroll.Content
	e.readTemplate(scroll.Type + "_footer")
	e.readTemplate("footer")

	if e.err != nil {
		x.error = errors.Wrapf(e.err, "producing latex file for scroll %v", id)
		return
	}
	err = writeTemp(id, e.doc)
	x.error = errors.Wrapf(err, "writing latex file %v.tex to temporary directory", id)
}

func (x *xelatexImagemagickRenderer) latexToPdf(id ID) {
	if x.error != nil {
		return
	}

	msg, err := exec.Command("xelatex",
		"-halt-on-error", "-output-directory", Config.TempDirectory,
		Config.TempDirectory+string(id)).CombinedOutput()
	x.error = errors.Wrapf(err, "XeLaTeX build: %v", string(msg))
}

func (x *xelatexImagemagickRenderer) pdfToPng(i ID) {
	if x.error != nil {
		return
	}

	id := string(i)
	x.error = exec.Command("convert", "-trim",
		"-quality", strconv.Itoa(Config.Quality),
		"-density", strconv.Itoa(Config.Dpi),
		Config.TempDirectory+id+".pdf", Config.CacheDirectory+id+".png").Run()

}

func (x *xelatexImagemagickRenderer) deleteTemporaryFiles(id ID) {
	if x.error != nil {
		return
	}

	files, err := filepath.Glob(Config.TempDirectory + string(id) + ".*")
	if err != nil {
		logError(err)
		return
	}
	for _, file := range files {
		tryLogError(os.Remove(file))
	}
}

func (x *xelatexImagemagickRenderer) err() error {
	return x.error
}

// renderScroll takes a scroll ID and a renderer to create a PNG image from
// that scroll.
func renderScroll(id ID, renderer latexToPngRenderer) error {
	if isUpToDate(id) {
		return nil
	}

	renderer.scrollToLatex(id)
	renderer.latexToPdf(id)
	renderer.pdfToPng(id)
	tryLogError(renderer.err())
	renderer.deleteTemporaryFiles(id)

	return errors.Wrap(renderer.err(), "rendering")
}
