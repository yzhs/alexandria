// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package latex

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"

	"github.com/yzhs/alexandria/common"
)

// errNoSuchScroll is used when a query returns a scroll ID that is no longer
// in the library, i.e. has been deleted but RemoveFromIndex has not yet been
// called on it.
var errNoSuchScroll = errors.New("No such scroll")

type errTemplateReader struct {
	doc string
	err error
}

// Load a template file from disk and propagate errors
func (e *errTemplateReader) readTemplate(name string) {
	if e.err != nil {
		return
	}

	tmp, err := common.ReadTemplate(name)
	e.err = errors.Wrapf(err, "read template %v", name)
	e.doc += tmp
}

// latexToPngRenderer describes a LaTeX->PDF->PNG pipeline.
type latexToPngRenderer interface {
	// Create a LaTeX file from the content of the given scroll together
	// with all the appropriate templates.  The resulting file stored in
	// the temp directory.
	scrollToLatex(id common.ID)

	// Compile a LaTeX file with the given id to produce a PDF file.  Both
	// input and output files are in the temp directory.
	latexToPdf(id common.ID)

	// Convert PDF to PNG, storing the result in the cache directory.  From
	// there, it can be served by the web server or displayed to the user
	// via some other user interface.
	pdfToPng(id common.ID)

	deleteTemporaryFiles(id common.ID)

	err() error
}

// XelatexImagemagickRenderer uses xelatex to handle the LaTeX-to-PDF
// translation, ImageMagick to convert the PDF to a PNG.
type XelatexImagemagickRenderer struct {
	error error
}

func (x *XelatexImagemagickRenderer) scrollToLatex(id common.ID) {
	var e errTemplateReader

	scrollText, err := common.ReadScroll(id)
	if err != nil {
		if os.IsNotExist(err) {
			err = common.RemoveFromIndex(id)
			if err != nil {
				common.LogError(err)
			}
			x.error = errNoSuchScroll
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
	err = common.WriteTemp(id, e.doc)
	x.error = errors.Wrapf(err, "writing latex file %v.tex to temporary directory", id)
}

func (x *XelatexImagemagickRenderer) latexToPdf(id common.ID) {
	if x.error != nil {
		return
	}

	msg, err := exec.Command("xelatex",
		"-halt-on-error", "-output-directory", common.Config.TempDirectory,
		common.Config.TempDirectory+string(id)).CombinedOutput()
	x.error = errors.Wrapf(err, "XeLaTeX build: %v", string(msg))
}

func (x *XelatexImagemagickRenderer) pdfToPng(i common.ID) {
	if x.error != nil {
		return
	}

	id := string(i)
	x.error = exec.Command("convert", "-trim",
		"-quality", strconv.Itoa(common.Config.Quality),
		"-density", strconv.Itoa(common.Config.Dpi),
		common.Config.TempDirectory+id+".pdf", common.Config.CacheDirectory+id+".png").Run()

}

func (x *XelatexImagemagickRenderer) deleteTemporaryFiles(id common.ID) {
	if x.error != nil {
		return
	}

	files, err := filepath.Glob(common.Config.TempDirectory + string(id) + ".*")
	if err != nil {
		common.LogError(err)
		return
	}
	for _, file := range files {
		common.TryLogError(os.Remove(file))
	}
}

func (x *XelatexImagemagickRenderer) err() error {
	return x.error
}

// renderScroll takes a scroll ID and a renderer to create a PNG image from
// that scroll.
func renderScroll(id common.ID, renderer latexToPngRenderer) error {
	if common.IsUpToDate(id) {
		return nil
	}

	renderer.scrollToLatex(id)
	renderer.latexToPdf(id)
	renderer.pdfToPng(id)
	common.TryLogError(renderer.err())
	renderer.deleteTemporaryFiles(id)

	return errors.Wrap(renderer.err(), "rendering")
}
