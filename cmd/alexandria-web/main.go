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
func statsHandler(b alexandria.Backend) func(w http.ResponseWriter, r *http.Request) {
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

// Serve the search page.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	html, err := loadHTMLTemplate("main")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	fmt.Fprintln(w, string(html))
}

func loadHTMLTemplate(name string) ([]byte, error) {
	return ioutil.ReadFile(alexandria.Config.TemplateDirectory + "html/" + name + ".html")
}

type result struct {
	Query        string
	Matches      []alexandria.Scroll
	NumMatches   int
	TotalMatches int
}

func renderTemplate(w http.ResponseWriter, templateFile string, resultData result) {
	t, err := template.ParseFiles(alexandria.Config.TemplateDirectory + "html/" + templateFile + ".html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	err = t.Execute(w, resultData)
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
func queryHandler(b alexandria.Backend) func(http.ResponseWriter, *http.Request) {
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
				fmt.Fprintf(os.Stderr, "%v;\n", err.Error())
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

func main() {
	var profile, version bool
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	alexandria.InitConfig()
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
	serveDirectory("/static/", alexandria.Config.TemplateDirectory+"static")
	err := http.ListenAndServe(LISTEN_ON, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
}
