package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/codesoap/mycolog/store"
	"github.com/codesoap/mycolog/store/pics"
	"github.com/pkg/browser"
)

//go:embed assets
var assets embed.FS

var db store.DB
var picStore pics.PictureStore

func init() {
	dbFilename, err := getDBFilename()
	if err != nil {
		panic(err)
	}
	migrateDBFileTo(dbFilename)
	db, err = store.GetDB(dbFilename)
	if err != nil {
		panic(err)
	}

	dataDir, err := getDataDir()
	if err != nil {
		panic(err)
	}
	picStore = pics.PictureStore{Path: filepath.Join(dataDir, "pics")}
	if err := os.MkdirAll(picStore.Path, 0755); err != nil {
		panic(err)
	}
}

func main() {
	headless := flag.Bool("headless", false, "run mycolog without opening a browser")
	flag.Parse()

	picsServer := http.StripPrefix("/pics/", http.FileServer(http.Dir(picStore.Path)))
	http.HandleFunc("GET /", redirectToDefaultPage)
	http.Handle("GET /assets/", http.FileServer(http.FS(assets)))
	http.Handle("GET /pics/", picsServer)
	http.HandleFunc("GET /intro", serveIntro)
	http.HandleFunc("GET /spores", serveComponentList)
	http.HandleFunc("GET /mycelium", serveComponentList)
	http.HandleFunc("GET /spawn", serveComponentList)
	http.HandleFunc("GET /grows", serveComponentList)
	http.HandleFunc("GET /add-spores", serveAddComponent)
	http.HandleFunc("GET /add-mycelium", serveAddComponent)
	http.HandleFunc("GET /add-spawn", serveAddComponent)
	http.HandleFunc("GET /add-grow", serveAddComponent)
	http.HandleFunc("POST /add-spores", handleAddComponent)
	http.HandleFunc("POST /add-mycelium", handleAddComponent)
	http.HandleFunc("POST /add-spawn", handleAddComponent)
	http.HandleFunc("POST /add-grow", handleAddComponent)
	http.HandleFunc("GET /component/{id}", serveComponent)
	http.HandleFunc("POST /component/{id}", handleComponentUpdate)
	http.HandleFunc("GET /components-pictures/{id}", servePictures)
	http.HandleFunc("POST /components-pictures/{id}", addPictures)
	http.HandleFunc("GET /delete-picture/{name}", deletePicture)
	http.HandleFunc("GET /delete-component-dialog/{id}", serveDeleteComponentDialog)
	http.HandleFunc("GET /delete-component/{id}", handleDeleteComponent)
	http.HandleFunc("GET /change-species/{id}", serveChangeSpecies)
	http.HandleFunc("POST /change-species/{id}", handleChangeSpecies)

	log.Print("Serving mycolog v0.4.0 from port 8080.")
	if !*headless {
		go openInBrowserWhenServing("http://localhost:8080/")
	}
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
