package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/codesoap/mycolog/store"
	"github.com/codesoap/mycolog/store/pics"
)

type picturesTmplData struct {
	ID       int64
	Spores   bool
	Myc      bool
	Spawn    bool
	Grow     bool
	Token    string
	Pictures []pics.PictureName
}

func servePictures(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	comp, err := db.GetComponent(id)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	pics := picStore.Pictures(id)
	data := picturesTmplData{
		ID:       comp.ID,
		Spores:   comp.Type == store.TypeSpores,
		Myc:      comp.Type == store.TypeMycelium,
		Spawn:    comp.Type == store.TypeSpawn,
		Grow:     comp.Type == store.TypeGrow,
		Token:    comp.Token,
		Pictures: pics,
	}
	if err := tmpls["pictures"].Execute(w, data); err != nil {
		log.Println(err.Error())
	}
}

func addPictures(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		showError(w, err, r.URL.Path)
		return
	}
	mr, err := r.MultipartReader()
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
			showError(w, err, r.URL.Path)
			return
		} else if part.FormName() == "file" {
			if _, err := picStore.Add(id, part); err != nil {
				io.Copy(io.Discard, r.Body)
				r.Body.Close()
				showError(w, err, r.URL.Path)
				return
			}
		}
	}
	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func deletePicture(w http.ResponseWriter, r *http.Request) {
	err := picStore.Delete(pics.PictureName(r.PathValue("name")))
	redirect := "/components-pictures/" + r.FormValue("component-id")
	if err != nil {
		showError(w, err, redirect)
		return
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
