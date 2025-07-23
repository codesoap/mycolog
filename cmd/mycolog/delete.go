package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/codesoap/mycolog/store"
)

func serveDeleteComponentDialog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	w.Header().Add("Content-Type", "text/html")
	if err := tmpls["delete"].Execute(w, id); err != nil {
		log.Println(err.Error())
	}
}

func handleDeleteComponent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	comp, err := db.GetComponent(id)
	if err != nil {
		showError(w, err, "/intro")
		return
	}
	if err = db.DeleteComponent(id); err != nil {
		showError(w, err, fmt.Sprint("/component/", id))
		return
	}
	switch comp.Type {
	case store.TypeSpores:
		http.Redirect(w, r, "/spores", http.StatusSeeOther)
	case store.TypeMycelium:
		http.Redirect(w, r, "/mycelium", http.StatusSeeOther)
	case store.TypeSpawn:
		http.Redirect(w, r, "/spawn", http.StatusSeeOther)
	case store.TypeGrow:
		http.Redirect(w, r, "/grows", http.StatusSeeOther)
	default:
		showError(w, fmt.Errorf("component had unknown type"), "/intro")
	}
}
