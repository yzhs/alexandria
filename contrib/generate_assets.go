package main

import (
	"net/http"

	"github.com/shurcooL/vfsgen"
	"github.com/yzhs/alexandria"
)

func main() {
	opts := vfsgen.Options{PackageName: "alexandria", VariableName: "Assets"}

	templates := http.Dir("./templates")
	err := vfsgen.Generate(templates, opts)
	if err != nil {
		panic(err)
	}

}
