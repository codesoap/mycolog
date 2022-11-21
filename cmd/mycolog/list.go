package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codesoap/mycolog/store"
)

type listTmplData struct {
	IntroSelected  bool
	SporesSelected bool
	MycSelected    bool
	SpawnSelected  bool
	GrowSelected   bool
	Components     []store.Component
	KnownSpecies   []string
	WantedSpecies  string
	ShowGone       bool
	ShowOld        bool
}

func serveComponentList(w http.ResponseWriter, r *http.Request) {
	componentType, err := componentTypeFromListPath(r.URL.Path)
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	componentFilter := getComponentFilter(componentType, r)
	components, err := db.FindComponents(componentFilter)
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	knownSpecies, err := db.GetAllSpecies()
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	w.Header().Add("Content-Type", "text/html")
	data := listTmplData{
		SporesSelected: componentType == store.TypeSpores,
		MycSelected:    componentType == store.TypeMycelium,
		SpawnSelected:  componentType == store.TypeSpawn,
		GrowSelected:   componentType == store.TypeGrow,
		Components:     components,
		KnownSpecies:   knownSpecies,
		WantedSpecies:  r.FormValue("species"),
		ShowGone:       len(r.FormValue("gone")) > 0,
		ShowOld:        len(r.FormValue("old")) > 0,
	}
	if err := tmpls["list"].Execute(w, data); err != nil {
		log.Println(err.Error())
	}
}

func getComponentFilter(componentType store.ComponentType, r *http.Request) store.ComponentFilter {
	filter := store.ComponentFilter{
		Types: []store.ComponentType{componentType},
	}
	wantedSpecies := r.FormValue("species")
	if len(wantedSpecies) > 0 && wantedSpecies != "any" {
		filter.Species = []string{wantedSpecies}
	}
	if len(r.FormValue("gone")) == 0 {
		showGone := false
		filter.Gone = &showGone
	}
	if len(r.FormValue("old")) == 0 {
		since := time.Now().Add(-30 * time.Hour * 24)
		filter.Since = &since
	}
	return filter
}

func componentTypeFromListPath(path string) (store.ComponentType, error) {
	switch path {
	case "/spores":
		return store.TypeSpores, nil
	case "/mycelium":
		return store.TypeMycelium, nil
	case "/spawn":
		return store.TypeSpawn, nil
	case "/grows":
		return store.TypeGrow, nil
	}
	return "", fmt.Errorf("invalid component list URL")
}
