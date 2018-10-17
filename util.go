// Alexandria
//
// Copyright (C) 2015-2018  Colin Benner
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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// logError writes things to stderr.
func logError(err interface{}) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
}

// tryLogError checks whether an error occurred, and logs it if necessary.
func tryLogError(err interface{}) {
	if err != nil {
		logError(err)
	}
}

// Load the content of a given scroll from disk.
func readScroll(id ID) (string, error) {
	result, err := ioutil.ReadFile(Config.KnowledgeDirectory + string(id) + ".tex")
	return string(result), errors.Wrapf(err, "read scroll %v", id)
}

// Load the content of a template file with the given name.
func readTemplate(filename string) (string, error) {
	result, err := ioutil.ReadFile(Config.TemplateDirectory + "tex/" + filename + ".tex")
	return string(result), errors.Wrapf(err, "read template %v", filename)
}

// Write a TeX file with the given name and content to Alexandria's temp
// directory.
func writeTemp(id ID, data string) error {
	err := ioutil.WriteFile(Config.TempDirectory+string(id)+".tex", []byte(data), 0644)
	return errors.Wrapf(err, "write %v.tex to temporary directory", id)
}

// Compute the combined size of all files in a given directory.
func getDirSize(dir string) (int, int64, error) {
	directory, err := os.Open(dir)
	tryLogError(err)
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
func isUpToDate(id ID) bool {
	if templatesModTime == -1 {
		// Check template for modification times
		templatesModTime = 0

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
