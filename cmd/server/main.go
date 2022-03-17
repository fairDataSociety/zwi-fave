// Credits to https://github.com/akhenakh/gozim
package main

import (
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

type ResponseType int8

const (
	RedirectResponse ResponseType = iota
	DataResponse
	NoResponse

	storeRef = "7d76c8373ce8bcf635cb62e746f39de7278dfc189e887a904a69b19ad6fe616884a4e7da1357e9f597bb053c39ea6da4ffb8f9661e328aa4a6dfc15cb6ecb609"
	metaRef  = "3a4758cafe8c5fc07c043ac7ec8b9e86238d3f87d44f659696ba9879315cc6727e96966e2fc91604194f063cbf96ddfd249e6a86540a4e16b9afafa6e44c3f3a"
)

// CachedResponse cache the answer to an URL in the zim
type CachedResponse struct {
	ResponseType ResponseType
	Data         []byte
	MimeType     string
}

var (
	port       = flag.Int("port", -1, "port to listen to, read HOST env if not specified, default to 8080 otherwise")
	indexPath  = flag.String("index", "", "path for the index file")
	beeHost    = flag.String("bee", "", "bee API endpoint")
	beeIsProxy = flag.Bool("proxy", false, "if Bee endpoint is gateway proxy")
	offline    = flag.Bool("offline", false, "run server offline for listing only")
	help       = flag.Bool("help", false, "print help")

	b blockstore.Client
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
	if *help == true {
		flag.Usage()
		return
	}
	if indexPath == nil || *indexPath == "" {
		log.Fatal("index not found")
	}

	if beeHost == nil || *beeHost == "" {
		log.Fatal("please input bee endpoint")
	}
	if *offline {
		b = mock.NewMockBeeClient()
	} else {
		logger := logging.New(os.Stdout, logrus.ErrorLevel)
		b = bee.NewBeeClient(
			*beeHost,
			"",
			logger,
		)
	}

	if !b.CheckConnection(*beeIsProxy) {
		log.Fatal("connection unavailable")
	}
	// open the db
	_, err := os.Lstat(*indexPath)
	if os.IsNotExist(err) {
		fmt.Println("Hold tight. Downloading index....")
		err = os.MkdirAll(*indexPath, 0777)
		if err != nil {
			log.Fatal(err)
		}
		// downloadIndex
		metaHex, err := hex.DecodeString(metaRef)
		if err != nil {
			log.Fatal(err)
		}
		metaData, _, err := b.DownloadBlob(metaHex)
		if err != nil {
			log.Fatal(err)
		}
		metaFile, err := os.Create(filepath.Join(*indexPath, "index_meta.json"))
		if err != nil {
			log.Fatal(err)
		}
		defer metaFile.Close()
		_, err = metaFile.Write(metaData)
		if err != nil {
			log.Fatal(err)
		}

		storeHex, err := hex.DecodeString(storeRef)
		if err != nil {
			log.Fatal(err)
		}
		storeData, _, err := b.DownloadBlob(storeHex)
		if err != nil {
			log.Fatal(err)
		}
		storeFile, err := os.Create(filepath.Join(*indexPath, "store"))
		if err != nil {
			log.Fatal(err)
		}
		defer storeFile.Close()
		_, err = storeFile.Write(storeData)
		if err != nil {
			log.Fatal(err)
		}
	}

	index, err = bleve.Open(*indexPath)
	if err != nil {
		log.Fatal(err)
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

	log.Println("Listening on", listenPath)

	err = http.ListenAndServe(listenPath, nil)
	if err != nil {
		log.Fatal(err)
	}
}
