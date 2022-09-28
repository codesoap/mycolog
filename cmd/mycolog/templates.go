package main

import (
	"embed"
	"html/template"
	"regexp"
)

//go:embed tmpl
var tmplFS embed.FS

var reComponentID = regexp.MustCompile("^[0-9]+")

type overviewTmplData struct {
	IntroSelected  bool
	SporesSelected bool
	MycSelected    bool
	SpawnSelected  bool
	GrowSelected   bool
}

var tmpls = make(map[string]*template.Template)

func init() {
	tmplMap := map[string][]string{
		"intro":         []string{"base", "overview", "intro"},
		"list":          []string{"base", "overview", "list"},
		"error":         []string{"base", "error"},
		"add":           []string{"base", "register_change_script", "add"},
		"details":       []string{"base", "register_change_script", "details"},
		"delete":        []string{"base", "delete"},
		"changeSpecies": []string{"base", "register_change_script", "change_species"},
	}
	for k, v := range tmplMap {
		filenames := make([]string, len(v))
		for i, name := range v {
			filenames[i] = "tmpl/" + name + ".html"
		}
		tmpls[k] = template.Must(template.ParseFS(tmplFS, filenames...))
	}
}
