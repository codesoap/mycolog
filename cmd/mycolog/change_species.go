package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/codesoap/mycolog/graph"
)

type changeSpeciesTmplData struct {
	ID           int64
	KnownSpecies []string
}

func serveChangeSpecies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	knownSpecies, err := db.GetAllSpecies()
	if err != nil {
		showError(w, err, fmt.Sprint("/component/", id))
		return
	}
	w.Header().Add("Content-Type", "text/html")
	data := changeSpeciesTmplData{
		ID:           id,
		KnownSpecies: knownSpecies,
	}
	if err := tmpls["changeSpecies"].Execute(w, data); err != nil {
		log.Println(err.Error())
	}
}

func handleChangeSpecies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	species := r.FormValue("species")
	if len(species) == 0 {
		showError(w, fmt.Errorf("species is empty"), r.URL.Path)
		return
	}
	relatives, err := graph.GetAllRelatives(db, id)
	if err != nil {
		showError(w, err, fmt.Sprint("/component/", id))
		return
	}
	err = db.UpdateSpecies(relatives, species)
	if err != nil {
		showError(w, err, fmt.Sprint("/component/", id))
		return
	}
	http.Redirect(w, r, fmt.Sprint("/component/", id, "?updated=true"), http.StatusSeeOther)
}
