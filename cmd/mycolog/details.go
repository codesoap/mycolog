package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codesoap/mycolog/cmd/mycolog/graphviz"
	"github.com/codesoap/mycolog/graph"
	"github.com/codesoap/mycolog/store"
)

type componentTmplData struct {
	ID        int64
	Parents   string // User readable representation of the parents.
	Transfers *int   // Transfers since spores.
	Spores    bool
	Myc       bool
	Spawn     bool
	Grow      bool
	Species   string
	Token     string
	CreatedAt string
	Notes     string
	Gone      bool

	Yield        *float64
	YieldComment string

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
	if comp.Type == store.TypeGrow {
		if err = updateGrowInfo(r, comp.ID); err != nil {
			showError(w, err, r.URL.Path)
			return
		}
	}
	http.Redirect(w, r, r.URL.Path+"?updated=true", http.StatusSeeOther)
}

func updateGrowInfo(r *http.Request, compID int64) error {
	yieldStr := r.FormValue("yield")
	yieldComment := r.FormValue("yieldComment")
	if yieldStr == "" && yieldComment == "" {
		_, err := db.DeleteGrowInfoIfPresent(compID)
		return err
	}
	var yield *int
	if yieldStr != "" {
		yieldFloat, err := strconv.ParseFloat(yieldStr, 64)
		if err != nil {
			return err
		}
		yieldVal := int(math.Round(yieldFloat * 1_000))
		yield = &yieldVal
	}
	return db.AttachGrowInfo(compID, yield, yieldComment)
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
		Transfers: getTransfersSinceSpores(comp),
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
	if comp.Type == store.TypeGrow {
		if err := fillYield(&data, comp.ID); err != nil {
			log.Println(err.Error())
		}
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

// getTransfersSinceSpores determines how many transfers have happened
// from spores to comp. This gives a rough estimate of how old the DNA
// is.
//
// If there is no parent that are spores, there are multiple parents
// somewhere in the lineage or there was an error querying the database,
// nil will be returned.
func getTransfersSinceSpores(comp store.Component) *int {
	for i := 0; ; i++ {
		if comp.Type == store.TypeSpores {
			return &i
		}
		parents, err := db.GetParents(comp.ID)
		if err != nil || len(parents) != 1 {
			return nil
		}
		if comp, err = db.GetComponent(parents[0]); err != nil {
			return nil
		}
	}
}

func fillYield(data *componentTmplData, compID int64) error {
	growInfo, err := db.GetGrowInfo(compID)
	if err != nil {
		return err
	}
	if growInfo.Yield != nil {
		yield := float64(*growInfo.Yield) / 1_000
		data.Yield = &yield
	}
	data.YieldComment = growInfo.YieldComment
	return nil
}
