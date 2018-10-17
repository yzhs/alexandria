// This file is part of Alexandria which is released under AGPLv3.
// Copyright (C) 2015-2018 Colin Benner
// See LICENSE or go to https://github.com/yzhs/alexandria/LICENSE for full
// license details.

package alexandria

import (
	"html/template"
)

// CompiledTemplates holds precompiled HTML templates.
type CompiledTemplates struct {
	cache map[string]*template.Template
}

var Templates CompiledTemplates

// Load gets a parsed template.Template, whether from cache or from disk.
func (templates CompiledTemplates) Load(typ, name string) *template.Template {
	fullName := name + "." + typ
	template, ok := templates.cache[fullName]
	if ok {
		return template
	}

	path := Config.TemplateDirectory + typ + "/" + name + "." + typ
	template, err := template.ParseFiles(path)
	if err != nil {
		panic(err)
	}
	templates.cache[fullName] = template
	return template
}
