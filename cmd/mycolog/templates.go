package main

import (
	"embed"
	"html/template"
	"math"
	"regexp"
	"strconv"
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
	funcMap := map[string]any{
		"getYieldString": getYieldString,
	}
	for k, v := range tmplMap {
		filenames := make([]string, len(v))
		for i, name := range v {
			filenames[i] = "tmpl/" + name + ".html"
		}
		t := template.New("base.html").Funcs(funcMap)
		tmpls[k] = template.Must(t.ParseFS(tmplFS, filenames...))
	}
}

func getYieldString(yields map[int64]float64, compID int64) string {
	yield, found := yields[compID]
	if found {
		if math.Abs(yield-math.Round(yield)) >= 0.001 {
			return "~" + strconv.FormatFloat(yield, 'f', 0, 64)
		}
		return strconv.FormatFloat(yield, 'f', 0, 64)
	}
	return ""
}
