// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

//go:generate go run ../contrib/generate_assets.go

package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// LogError writes things to stderr.
func LogError(err interface{}) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
}

// TryLogError checks whether an error occurred, and logs it if necessary.
func TryLogError(err interface{}) {
	if err != nil {
		LogError(err)
	}
}

// Load the content of a given scroll from disk.
func ReadScroll(id ID) (string, error) {
	result, err := ioutil.ReadFile(Config.KnowledgeDirectory + string(id) + ".tex")
	return string(result), errors.Wrapf(err, "read scroll %v", id)
}

// Load the content of a template file with the given name.
func ReadTemplate(filename string) (string, error) {
	path := "tex/" + filename + ".tex"
	file, err := Assets.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "opening asset %v", path)
	}
	result, err := ioutil.ReadAll(file)
	return string(result), errors.Wrap(err, "readall")
}

// Write a TeX file with the given name and content to Alexandria's temp
// directory.
func WriteTemp(id ID, data string) error {
	err := ioutil.WriteFile(Config.TempDirectory+string(id)+".tex", []byte(data), 0644)
	return errors.Wrapf(err, "write %v.tex to temporary directory", id)
}

// Compute the combined size of all files in a given directory.
func getDirSize(dir string) (int, int64, error) {
	directory, err := os.Open(dir)
	TryLogError(err)
	defer directory.Close()
	fileInfo, err := directory.Readdir(0)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "read directory %v", directory)
	}
	result := int64(0)
	for _, file := range fileInfo {
		result += file.Size()
	}
	return len(fileInfo), result, nil
}

// Get the time a given file was last modified as a Unix time.
func getModTime(file string) (int64, error) {
	info, err := os.Stat(file)
	if err != nil {
		return -1, errors.Wrapf(err, "stat %v", file)
	}
	return info.ModTime().Unix(), nil
}

// Cache the newest modification of any of the template files as a Unix time
// (i.e. seconds since 1970-01-01).
var templatesModTime int64 = -1

// All recognized template files
// TODO Generate the list⁈
var templateFiles = []string{
	"header.tex", "footer.tex",
	"algorithm_header.tex", "algorithm_footer.tex",
	"axiom_header.tex", "axiom_footer.tex",
	"corollary_header.tex", "corollary_footer.tex",
	"definition_header.tex", "definition_footer.tex",
	"example_header.tex", "example_footer.tex",
	"exercise_header.tex", "exercise_footer.tex",
	"lemma_header.tex", "lemma_footer.tex",
	"proof_header.tex", "proof_footer.tex",
	"proposition_header.tex", "proposition_footer.tex",
	"remark_header.tex", "remark_footer.tex",
	"theorem_header.tex", "theorem_footer.tex"}

// Check whether a given scroll has to be recompiled
func IsUpToDate(id ID) bool {
	if templatesModTime == -1 {
		// Check template for modification times
		templatesModTime = 0
		foo, err := getModTime(os.Args[0])
		if err != nil && foo > templatesModTime {
			templatesModTime = foo
		}

		for _, file := range templateFiles {
			foo, err := getModTime(Config.TemplateDirectory + file)
			if err != nil {
				break
			}
			if foo > templatesModTime {
				templatesModTime = foo
			}
		}
	}

	info, err := os.Stat(Config.CacheDirectory + string(id) + ".png")
	if err != nil {
		return false
	}
	imageTime := info.ModTime().Unix()

	if imageTime < templatesModTime {
		return false
	}

	info, err = os.Stat(Config.KnowledgeDirectory + string(id) + ".tex")
	if err != nil {
		return false // When in doubt, recompile
	}
	scrollTime := info.ModTime().Unix()

	return imageTime > scrollTime
}
