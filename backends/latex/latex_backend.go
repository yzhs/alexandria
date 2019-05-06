// This file contains most of the public interface of Alexandria (the library).
// The remainder is the types defined in datastructures.go.

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package latex

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/yzhs/alexandria/common"
)

type ID = common.ID

type LatexToPngBackend struct {
	Renderer latexToPngRenderer
}

func (b LatexToPngBackend) RenderAllScrolls() (numScrolls int, errors []error) {
	ids := b.IDsForAllScrolls()
	renderedIDs, errors := b.RenderScrollsByID(ids)
	return len(renderedIDs), errors
}

func (b LatexToPngBackend) RenderScrollsByID(ids []common.ID) (renderedScrollIDs []common.ID, errors []error) {
	renderedScrollIDs = make([]common.ID, len(ids))
	numScrolls := 0
	for _, id := range ids {
		err := renderScroll(id, b.Renderer)
		if err != nil {
			errors = append(errors, fmt.Errorf("rendering %v failed", id))
		} else {
			renderedScrollIDs[numScrolls] = id
			numScrolls++
		}
	}

	return renderedScrollIDs[:numScrolls], errors
}

// IDsForAllScrolls goes through the library directory and creates a list of
// the IDs of all available LaTeX scrolls.
func (LatexToPngBackend) IDsForAllScrolls() []ID {
	files, err := ioutil.ReadDir(common.Config.KnowledgeDirectory)
	if err != nil {
		panic(err)
	}

	var ids []ID
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".tex") {
			continue
		}
		id := ID(strings.TrimSuffix(file.Name(), ".tex"))
		ids = append(ids, id)
	}
	return ids
}
