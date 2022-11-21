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
		"intro":         {"base", "overview", "intro"},
		"list":          {"base", "overview", "list"},
		"error":         {"base", "error"},
		"add":           {"base", "register_change_script", "add"},
		"details":       {"base", "register_change_script", "details"},
		"delete":        {"base", "delete"},
		"changeSpecies": {"base", "register_change_script", "change_species"},
	}
	for k, v := range tmplMap {
		filenames := make([]string, len(v))
		for i, name := range v {
			filenames[i] = "tmpl/" + name + ".html"
		}
		tmpls[k] = template.Must(template.ParseFS(tmplFS, filenames...))
	}
}
