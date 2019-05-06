package main

import (
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	opts := vfsgen.Options{PackageName: "common", VariableName: "Assets"}

	templates := http.Dir("../templates")
	err := vfsgen.Generate(templates, opts)
	if err != nil {
		panic(err)
	}

}
