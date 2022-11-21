package main

import (
	"embed"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"github.com/codesoap/mycolog/store"
)

//go:embed assets
var assets embed.FS

var db store.DB

func init() {
	dbFilename, err := getDBFilename()
	if err != nil {
		panic(err)
	}
	db, err = store.GetDB(dbFilename)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", redirectToDefaultPage)
	http.Handle("/assets/", http.FileServer(http.FS(assets)))
	http.HandleFunc("/intro", serveIntro)
	http.HandleFunc("/spores", serveComponentList)
	http.HandleFunc("/mycelium", serveComponentList)
	http.HandleFunc("/spawn", serveComponentList)
	http.HandleFunc("/grows", serveComponentList)
	http.HandleFunc("/add-spores", handleAddComponent)
	http.HandleFunc("/add-mycelium", handleAddComponent)
	http.HandleFunc("/add-spawn", handleAddComponent)
	http.HandleFunc("/add-grow", handleAddComponent)
	http.HandleFunc("/component/", handleComponent)
	http.HandleFunc("/delete-component-dialog/", serveDeleteComponentDialog)
	http.HandleFunc("/delete-component/", handleDeleteComponent)
	http.HandleFunc("/change-species/", handleChangeSpecies)

	log.Print("Serving from port 8080.")
	go openInBrowserWhenServing("http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func openInBrowserWhenServing(url string) {
	for i := 0; i < 10; i++ {
		time.Sleep(200 * time.Millisecond)
		if _, err := http.Get(url); err == nil {
			if err := browser.OpenURL(url); err != nil {
				break
			}
			return
		}
	}
	log.Println("Could not open browser, please go to", url, "manually.")
}

func redirectToDefaultPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/intro", http.StatusTemporaryRedirect)
}

func serveIntro(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	data := overviewTmplData{IntroSelected: true}
	if err := tmpls["intro"].Execute(w, data); err != nil {
		log.Println(err.Error())
	}
}
