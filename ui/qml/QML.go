// Alexandria
//
// Copyright (C) 2015  Colin Benner
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
	"os"
	"runtime/pprof"
	"strings"

	flag "github.com/ogier/pflag"

	. "github.com/yzhs/alexandria"
	render "github.com/yzhs/alexandria/render/xelatex"

	"gopkg.in/qml.v1"
)

func printStats() string {
	stats := render.ComputeStatistics()
	n := stats.Num()
	size := float32(stats.Size()) / 1024.0
	return fmt.Sprintf("The library contains %v scrolls with a total size of %.1f kiB.\n", n, size)
}

type SearchEngine struct {
	resultList qml.Object
}

func (s SearchEngine) Search(query string) {
	ids, err := render.FindScrolls(strings.Split(query, " "))
	if err != nil {
		panic(err)
	}
	lst := s.resultList
	results := make([]string, len(ids))
	for i := range ids {
		results[i] = string(ids[i])
	}
	lst.Call("setIds", results)
}

func run() error {
	engine := qml.NewEngine()
	context := engine.Context()
	searchEngine := SearchEngine{}
	context.SetVar("searchEngine", &searchEngine)

	component, err := engine.LoadFile(Config.AlexandriaDirectory + "qml/main.qml")
	if err != nil {
		return err
	}
	win := component.CreateWindow(nil)
	searchEngine.resultList = win.Root().ObjectByName("resultList")
	win.Show()
	win.Wait()
	return nil
}

func main() {
	var profile, version bool
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	InitConfig()

	if profile {
		f, err := os.Create("alexandria.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// TODO run when there is something new render.GenerateIndex()

	if version {
		fmt.Println(NAME, VERSION)
		return
	}

	err := qml.Run(run)
	if err != nil {
		panic(err)
	}
}
