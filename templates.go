package alexandria

import (
	"html/template"
)

type CompiledTemplates struct {
	cache map[string]*template.Template
}

var Templates CompiledTemplates

func (self CompiledTemplates) Load(typ, name string) *template.Template {
	fullName := name + "." + typ
	template, ok := self.cache[fullName]
	if ok {
		return template
	}

	path := Config.TemplateDirectory + typ + "/" + name + "." + typ
	template, err := template.ParseFiles(path)
	if err != nil {
		panic(err)
	}
	self.cache[fullName] = template
	return template
}
