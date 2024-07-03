package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/codesoap/mycolog/graph"
)

type changeSpeciesTmplData struct {
	ID           int64
	KnownSpecies []string
}

func handleChangeSpecies(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.Split(r.URL.Path, "/")
	if len(pathSplit) != 3 {
		showError(w, fmt.Errorf("invalid URL"), r.URL.Path)
		return
	}
	id, err := strconv.ParseInt(pathSplit[len(pathSplit)-1], 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	if r.Method == http.MethodPost {
		handleUpdateSpecies(w, r, id)
		return
	}
	serveChangeSpecies(w, id)
}

func serveChangeSpecies(w http.ResponseWriter, id int64) {
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

func handleUpdateSpecies(w http.ResponseWriter, r *http.Request, id int64) {
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
