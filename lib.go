// This file contains most of the public interface of Alexandria (the library).
// The remainder is the types defined in datastructures.go.

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"github.com/yzhs/alexandria/backends/latex"
	"github.com/yzhs/alexandria/common"
)

// Programm name and version
const (
	NAME    = "Alexandria"
	VERSION = "0.1"
)

type scrollType int

const (
	latexScroll scrollType = iota
	markdownScroll
)

type (
	Backend    = common.Backend
	ID         = common.ID
	Scroll     = common.Scroll
	Statistics = common.Statistics
)

var (
	Assets = common.Assets
	Config = common.Config
)

func NewBackend() common.Backend {
	return latex.LatexToPngBackend{Renderer: &latex.XelatexImagemagickRenderer{}}
}

func LoadScrolls(ids []ID) ([]common.Scroll, error) {
	return common.LoadScrolls(NewBackend(), ids)
}

func UpdateIndex() error {
	return common.UpdateIndex(NewBackend())
}

func FindMatchingScrolls(query string) ([]ID, int, error) {
	return common.FindMatchingScrolls(query)
}

func ComputeStatistics() (Statistics, error) {
	return common.ComputeStatistics()
}
