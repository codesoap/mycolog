package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mycolog/cmd/mycolog/graphviz"
	"mycolog/graph"
	"mycolog/store"
)

type componentTmplData struct {
	ID        int64
	Parents   string // User readable representation of the parents.
	Spores    bool
	Myc       bool
	Spawn     bool
	Grow      bool
	Species   string
	Token     string
	CreatedAt string
	Notes     string
	Gone      bool

	Updated bool

	FullGraph bool
	Graph     template.HTML
}

func handleComponent(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	if len(pathSplit) != 3 {
		showError(w, fmt.Errorf("invalid component URL"), r.URL.Path)
		return
	}
	id, err := strconv.ParseInt(pathSplit[len(pathSplit)-1], 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	comp, err := db.GetComponent(id)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	if r.Method == http.MethodPost {
		handleUpdateComponent(w, r, comp)
		return
	}
	handleGetComponent(w, r, comp)
}

func handleUpdateComponent(w http.ResponseWriter, r *http.Request, comp store.Component) {
	createdAt, err := time.Parse("2006-01-02", r.FormValue("createdAt"))
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	gone := r.FormValue("gone") == "true"
	err = db.UpdateComponent(comp.ID, createdAt, r.FormValue("notes"), gone)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	http.Redirect(w, r, r.URL.Path+"?updated=true", http.StatusSeeOther)
}

func handleGetComponent(w http.ResponseWriter, r *http.Request, comp store.Component) {
	parents, err := getParentsString(comp.ID)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	graph, err := getGraph(comp.ID, r.FormValue("fullgraph") == "true")
	if err != nil {
		// Continue without graph.
		graph = ""
		log.Println("Could not render graph:", err.Error())
	}
	w.Header().Add("Content-Type", "text/html")
	data := componentTmplData{
		Updated:   r.FormValue("updated") == "true",
		ID:        comp.ID,
		Parents:   parents,
		Spores:    comp.Type == store.TypeSpores,
		Myc:       comp.Type == store.TypeMycelium,
		Spawn:     comp.Type == store.TypeSpawn,
		Grow:      comp.Type == store.TypeGrow,
		Species:   comp.Species,
		Token:     comp.Token,
		CreatedAt: comp.CreatedAt.Format("2006-01-02"),
		Notes:     comp.Notes,
		Gone:      comp.Gone,
		FullGraph: r.FormValue("fullgraph") == "true",
		Graph:     template.HTML(graph),
	}
	if err := tmpls["details"].Execute(w, data); err != nil {
		log.Println(err.Error())
	}
}

func getParentsString(id int64) (string, error) {
	parents, err := db.GetParents(id)
	if err != nil {
		return "", err
	}
	var parentsString string
	for i, parent := range parents {
		if i == 0 {
			parentsString += fmt.Sprint("#", parent)
		} else {
			parentsString += fmt.Sprint(", #", parent)
		}
	}
	if len(parentsString) == 0 {
		return "none", nil
	}
	return parentsString, nil
}

func getGraph(id int64, fullgraph bool) (string, error) {
	var err error
	var relatives []graph.Relative
	if fullgraph {
		relatives, err = graph.GetFullLineage(db, id)
	} else {
		relatives, err = graph.GetCloseLineage(db, id)
	}
	if err != nil {
		return "", err
	}
	return graphviz.Render(relatives, id)
}
