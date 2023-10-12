// Credits to https://github.com/akhenakh/gozim
package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	lru "github.com/hashicorp/golang-lru"
)

type ResponseType int8

const (
	RedirectResponse ResponseType = iota
	DataResponse
	NoResponse
)

// CachedResponse cache the answer to an URL in the zim
type CachedResponse struct {
	ResponseType ResponseType
	Data         []byte
	MimeType     string
}

var (
	port       = flag.Int("port", -1, "port to listen to, read HOST env if not specified, default to 8080 otherwise")
	fave       = flag.String("fave", "http://localhost:1234/v1", "FaVe API endpoint")
	collection = flag.String("collection", "", "Collection name to store content in FaVe")
	help       = flag.Bool("help", false, "print help")

	// Cache is filled with CachedResponse to avoid hitting the zim file for a zim URL
	cache *lru.ARCCache

	templates *template.Template

	//go:embed static
	staticFS embed.FS

	//go:embed templates/*
	templateFS embed.FS
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tpls, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	templates = tpls

	// static file handler
	fileServer := http.FileServer(http.FS(staticFS))
	http.Handle("/static/", fileServer)

	// compress wiki pages
	http.HandleFunc("/wiki/", makeGzipHandler(wikiHandler))

	// tpl
	http.HandleFunc("/search/", makeGzipHandler(searchHandler))
	http.HandleFunc("/", makeGzipHandler(browseHandler))

	// the need for a cache is absolute
	// a lot of the same urls will be called repeatedly, css, js ...
	// avoid to look for those one
	cache, _ = lru.NewARC(40)

	// default listening to port 8080
	listenPath := ":8080"

	if len(os.Getenv("PORT")) > 0 {
		listenPath = ":" + os.Getenv("PORT")
	}

	if port != nil && *port > 0 {
		listenPath = ":" + strconv.Itoa(*port)
	}

	log.Println("Listening on", listenPath)

	err = http.ListenAndServe(listenPath, nil)
	if err != nil {
		log.Fatal(err)
	}
}
