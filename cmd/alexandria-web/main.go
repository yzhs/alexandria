// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"

	flag "github.com/ogier/pflag"

	"github.com/yzhs/alexandria"
)

const (
	LISTEN_ON   = "127.0.0.1:41665"
	MAX_RESULTS = 100
)

// Send the statistics page to the client.
func statsHandler(b alexandria.LatexToPngBackend) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := b.Statistics()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		n := stats.NumberOfScrolls()
		size := float32(stats.TotalSize()) / 1024.0
		fmt.Fprintf(w, "The library contains %v scrolls with a total size of %.1f kiB.\n", n, size)
	}
}

// Handle the edit-link, causing the browser to open that scroll in an editor.
func editHandler(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers["Content-Type"] = []string{"application/x-alexandria-edit"}
	id := r.FormValue("id")
	fmt.Fprintf(w, id)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, alexandria.Config.TemplateDirectory+"html/main.html")
}

type result struct {
	Query        string
	Matches      []alexandria.Scroll
	NumMatches   int
	TotalMatches int
}

func renderTemplate(w http.ResponseWriter, templateFile string, resultData result) {
	err := loadTemplate("search").Execute(w, resultData)
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Handle a query and serve the results.
func queryHandler(b alexandria.LatexToPngBackend) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.FormValue("q")
		if query == "" {
			mainHandler(w, r)
			return
		}
		ids, totalMatches, err := b.FindMatchingScrolls(query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ids, errors := b.RenderScrollsByID(ids)
		if len(errors) != 0 {
			errorStrings := make([]string, len(errors))
			for i, err := range errors {
				errorStrings[i] = err.Error()
				fmt.Fprintf(os.Stderr, "%v;\n", errorStrings[i])
			}
			http.Error(w, strings.Join(errorStrings, ";\n"), http.StatusInternalServerError)
			return
		}
		numMatches := len(ids)
		results, err := b.LoadScrolls(ids)
		data := result{Query: query, NumMatches: numMatches, Matches: results[:min(20, numMatches)], TotalMatches: totalMatches}
		renderTemplate(w, "search", data)
	}
}

func serveDirectory(prefix string, directory string) {
	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(directory))))
}

// Load gets a parsed template.Template, whether from cache or from disk.
func loadTemplate(name string) *template.Template {
	relPath := "html/" + name + ".html"
	path := alexandria.Config.TemplateDirectory + relPath
	template, err := template.ParseFiles(path)
	if os.IsNotExist(err) {
		template, err = loadTemplateFromAssets(relPath)
	}
	if err != nil {
		panic(err)
	}
	return template
}

func loadTemplateFromAssets(relPath string) (*template.Template, error) {
	file, err := alexandria.Assets.Open(relPath)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return template.New(relPath).Parse(string(content))
}

func main() {
	var profile, version bool
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	alexandria.Config.MaxResults = MAX_RESULTS

	if profile {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if version {
		fmt.Println(alexandria.NAME, alexandria.VERSION)
		return
	}

	b := alexandria.NewBackend()
	b.UpdateIndex()

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/stats", statsHandler(b))
	http.HandleFunc("/search", queryHandler(b))
	http.HandleFunc("/alexandria.edit", editHandler)
	serveDirectory("/images/", alexandria.Config.CacheDirectory)
	http.Handle("/static/", http.FileServer(alexandria.Assets))
	err := http.ListenAndServe(LISTEN_ON, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
}
