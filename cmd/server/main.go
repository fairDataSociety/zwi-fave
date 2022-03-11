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

	"github.com/blevesearch/bleve"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	lru "github.com/hashicorp/golang-lru"
	"github.com/onepeerlabs/w3kipedia/pkg/bee"
	"github.com/sirupsen/logrus"
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
	port      = flag.Int("port", -1, "port to listen to, read HOST env if not specified, default to 8080 otherwise")
	indexPath = flag.String("index", "", "path for the index file")
	beeHost   = flag.String("bee", "", "Bee API endpoint")
	beeIsProxy   = flag.Bool("proxy", false, "If Bee endpoint is gateway proxy")
	batch     = flag.String("batch", "", "Bee Postage Stamp ID")

	B *bee.BeeClient
	// Cache is filled with CachedResponse to avoid hitting the zim file for a zim URL
	cache *lru.ARCCache
	index bleve.Index

	templates *template.Template

	//go:embed static
	staticFS embed.FS

	//go:embed templates/*
	templateFS embed.FS
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	// Do we have an index ?
	if indexPath == nil || *indexPath == "" {
		log.Fatal("index not found")
	}

	if beeHost == nil || *beeHost == "" {
		log.Fatal("please input bee endpoint")
	}

	if batch == nil || *batch == "" {
		log.Fatal("please input batch-id")
	}

	if _, err := os.Stat(*indexPath); err != nil {
		log.Fatal(err)
	}

	// open the db
	var err error
	index, err = bleve.Open(*indexPath)
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.New(os.Stdout, logrus.ErrorLevel)
	B = bee.NewBeeClient(
		*beeHost,
		"",
		*batch,
		logger,
	)
	if !B.CheckConnection(*beeIsProxy) {
		log.Fatal("connection unavailable")
	}

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

	// Opening large indexes could takes minutes on raspberry
	log.Println("Listening on", listenPath)

	err = http.ListenAndServe(listenPath, nil)
	if err != nil {
		log.Fatal(err)
	}
}
