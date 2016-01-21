// Alexandria
//
// Copyright (C) 2015  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package alexandria

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func LogError(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
}

func TryLogError(err interface{}) {
	if err != nil {
		LogError(err)
	}
}

func GenerateIndex() error {
	return exec.Command("index++", "-c", Config.SwishConfig, Config.KnowledgeDirectory).Run()
}

func readScroll(id Id) (string, error) {
	result, err := ioutil.ReadFile(Config.KnowledgeDirectory + string(id) + ".tex")
	return string(result), err
}

func readTemplate(filename string) (string, error) {
	result, err := ioutil.ReadFile(Config.TemplateDirectory + filename + ".tex")
	return string(result), err
}

// Write a TeX file with the given name and content to Alexandria's temp
// directory.
func writeTemp(id Id, data string) error {
	return ioutil.WriteFile(Config.TempDirectory+string(id)+".tex", []byte(data), 0644)
}

// Compute the combined size of all files in a given directory.
func getDirSize(dir string) (int, int64) {
	directory, err := os.Open(dir)
	TryLogError(err)
	defer directory.Close()
	fileInfo, err := directory.Readdir(0)
	if err != nil {
		panic(err)
	}
	result := int64(0)
	for _, file := range fileInfo {
		result += file.Size()
	}
	return len(fileInfo), result
}

func ComputeStatistics() Statistics {
	num, size := getDirSize(Config.KnowledgeDirectory)
	return Stats{num, size}
}

// Get the time a given file was last modified as a Unix time.
func getModTime(file string) (int64, error) {
	info, err := os.Stat(file)
	if err != nil {
		return -1, err
	}
	return info.ModTime().Unix(), nil
}

// Cache the newest modification of any of the template files as a Unix time
// (i.e. seconds since 1970-01-01).
var templatesModTime int64 = -1

// All recognized template files
// TODO Generate the list⁈
var templateFiles []string = []string{"algorithm_footer.tex",
	"algorithm_header.tex", "axiom_footer.tex", "axiom_header.tex",
	"corollary_footer.tex", "corollary_header.tex", "definition_footer.tex",
	"definition_header.tex", "example_footer.tex", "example_header.tex",
	"footer.tex", "header.tex", "lemma_footer.tex", "lemma_header.tex",
	"proposition_footer.tex", "proposition_header.tex",
	"remark_header.tex", "remark_footer.tex", "theorem_footer.tex",
	"theorem_header.tex"}

// Check whether a given scroll has to be recompiled
func isUpToDate(id Id) bool {
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
