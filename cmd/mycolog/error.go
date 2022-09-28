package main

import (
	"log"
	"net/http"
)

func showError(w http.ResponseWriter, err error, redirect string) {
	log.Println(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	data := make(map[string]string)
	data["ErrorMsg"] = err.Error()
	data["Redirect"] = redirect
	if err2 := tmpls["error"].Execute(w, data); err2 != nil {
		log.Println(err2.Error())
	}
}
