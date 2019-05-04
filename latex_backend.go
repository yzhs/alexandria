// This file contains most of the public interface of Alexandria (the library).
// The remainder is the types defined in datastructures.go.

// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

type LatexToPngBackend struct {
	renderer latexToPngRenderer
}

func (b LatexToPngBackend) RenderAllScrolls() (numScrolls int, errors []error) {
	ids := b.IDsForAllScrolls()
	renderedIDs, errors := b.RenderScrollsByID(ids)
	return len(renderedIDs), errors
}

func (b LatexToPngBackend) RenderScrollsByID(ids []ID) (renderedScrollIDs []ID, errors []error) {
	renderedScrollIDs = make([]ID, len(ids))
	numScrolls := 0
	for _, id := range ids {
		err := renderScroll(id, b.renderer)
		if err != nil {
			errors = append(errors, fmt.Errorf("rendering %v failed", id))
		} else {
			renderedScrollIDs[numScrolls] = id
			numScrolls++
		}
	}

	return renderedScrollIDs[:numScrolls], errors
}
