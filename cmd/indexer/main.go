package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	zim "github.com/akhenakh/gozim"
	"github.com/blevesearch/bleve"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/onepeerlabs/w3kipedia/pkg/bee"
	"github.com/onepeerlabs/w3kipedia/pkg/limiter"
	"github.com/sirupsen/logrus"
)

type article struct {
	Title       string
	Namespace   string
	Content     string
	FullURL     string
	MimeType    string
	EntryType   string
	Address     string
	RedirectURL string
}

var (
	indexPath = flag.String("index", "", "path for the index file")
	beeHost   = flag.String("bee", "", "Bee API endpoint")
	batch     = flag.String("batch", "", "Bee Postage Stamp ID")
	zimPath   = flag.String("zim", "", "zim file location")
)

func main() {
	flag.Parse()
	if indexPath == nil || *indexPath == "" {
		log.Fatal("index not found")
	}

	if beeHost == nil || *beeHost == "" {
		log.Fatal("please input bee endpoint")
	}

	if batch == nil || *batch == "" {
		log.Fatal("please input batch-id")
	}

	if zimPath == nil || *zimPath == "" {
		log.Fatal("please input zim location")
	}

	logger := logging.New(os.Stdout, logrus.ErrorLevel)
	b := bee.NewBeeClient(
		*beeHost,
		"",
		*batch,
		logger,
	)
	if !b.CheckConnection() {
		log.Fatal("connection unavailable")
	}
	bleve.Config.DefaultKVStore = "boltdb"
	mapping := bleve.NewIndexMapping()
	mapping.DefaultType = "Article"

	articleMapping := bleve.NewDocumentMapping()
	mapping.AddDocumentMapping("Article", articleMapping)

	indexMapping := bleve.NewTextFieldMapping()
	indexMapping.Store = true
	indexMapping.Index = true
	indexMapping.Analyzer = "standard"

	nonIndexMapping := bleve.NewTextFieldMapping()
	nonIndexMapping.Store = true
	nonIndexMapping.Index = false
	nonIndexMapping.Analyzer = "standard"

	articleMapping.AddFieldMappingsAt("Title", indexMapping)
	articleMapping.AddFieldMappingsAt("FullURL", nonIndexMapping)
	articleMapping.AddFieldMappingsAt("MimeType", indexMapping)
	articleMapping.AddFieldMappingsAt("EntryType", nonIndexMapping)
	articleMapping.AddFieldMappingsAt("Address", nonIndexMapping)
	articleMapping.AddFieldMappingsAt("RedirectURL", nonIndexMapping)

	index, err := bleve.New(*indexPath, mapping)
	if err != nil {
		fmt.Println("asdasd")
		log.Fatal(err)
	}
	defer index.Close()

	// open zim
	z, err := zim.NewReader(*zimPath, false)
	if err != nil {
		log.Fatal(err)
	}
	/*  read zim
	upload
	*/
	z.ListArticles()
	limit := limiter.NewConcurrencyLimiter(10)

	z.ListTitlesPtrIterator(func(idx uint32) {
		limit.Execute(func() {
			a, err := z.ArticleAtURLIdx(idx)
			if err != nil || a.EntryType == zim.DeletedEntry {
				return
			}
			redirectURL := ""
			if a.EntryType == zim.RedirectEntry {
				ridx, err := a.RedirectIndex()
				fmt.Println(ridx)
				if err != nil {
					return
				}
				ra, err := z.ArticleAtURLIdx(ridx)
				if err != nil {
					return
				}
				redirectURL = ra.FullURL()
			}
			data, err := a.Data()
			if err != nil {
				log.Fatal(err.Error())
			}
			address, err := b.UploadBlob(data, true, true)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(len(data), a.EntryType, a.Title, a.FullURL(), hex.EncodeToString(address))
			idoc := article{
				Title:       a.Title,
				Namespace:   string(a.Namespace),
				Content:     string(data),
				FullURL:     a.FullURL(),
				MimeType:    a.MimeType(),
				EntryType:   fmt.Sprintf("%d", a.EntryType),
				Address:     hex.EncodeToString(address),
				RedirectURL: redirectURL,
			}
			err = index.Index(a.FullURL(), idoc)
			if err != nil {
				log.Fatal(err.Error())
			}
		})

	})
	limit.Wait()
	fmt.Println("w3kipedia")
}
