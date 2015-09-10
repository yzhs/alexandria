package main

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

func run(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}

func generateIndex() error {
	return run("index++", "-c", config.swishConfig, config.knowledgeDirectory)
}

func readScroll(id string) (string, error) {
	result, err := ioutil.ReadFile(config.knowledgeDirectory + id + ".tex")
	return string(result), err
}

func readTemplate(filename string) (string, error) {
	result, err := ioutil.ReadFile(config.templateDirectory + filename + ".tex")
	return string(result), err
}

// Write a TeX file with the given name and content to Alexandria's temp
// directory.
func writeTemp(id, data string) error {
	return ioutil.WriteFile(config.tempDirectory+id+".tex", []byte(data), 0644)
}

// Get the file size in bytes.
func getFileSize(file string) int64 {
	f, err := os.Open(file)
	if err != nil {
		LogError(err)
		return 0
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		LogError(err)
		return 0
	}
	return stat.Size()
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
	"proposition_footer.tex", "proposition_header.tex", "theorem_footer.tex",
	"theorem_header.tex"}

// Check whether a given scroll has to be recompiled
func isUpToDate(id string) bool {
	if templatesModTime == -1 {
		// Check template for modification times
		templatesModTime = 0

		for _, file := range templateFiles {
			foo, err := getModTime(config.templateDirectory + file)
			if err != nil {
				break
			}
			if foo > templatesModTime {
				templatesModTime = foo
			}
		}
	}

	info, err := os.Stat(config.cacheDirectory + id + ".png")
	if err != nil {
		return false
	}
	imageTime := info.ModTime().Unix()

	if imageTime < templatesModTime {
		return false
	}

	info, err = os.Stat(config.knowledgeDirectory + id + ".tex")
	if err != nil {
		return false // When in doubt, recompile
	}
	scrollTime := info.ModTime().Unix()

	return imageTime > scrollTime
}
