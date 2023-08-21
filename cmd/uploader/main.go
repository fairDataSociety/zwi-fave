package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	zim "github.com/akhenakh/gozim"
	"github.com/google/uuid"
	"github.com/jaytaylor/html2text"
	"github.com/microcosm-cc/bluemonday"
	swagger "github.com/onepeerlabs/w3kipedia/pkg/go-client"
)

var (
	zimPath    = flag.String("zim", "", "zim file location")
	fave       = flag.String("fave", "http://localhost:1234/v1", "FaVe API endpoint")
	collection = flag.String("collection", "", "Collection name to store content in FaVe")
	help       = flag.Bool("help", false, "print help")
)

func main() {

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if zimPath == nil || *zimPath == "" {
		log.Fatal("please input zim location")
	}

	if fave == nil || *fave == "" {
		log.Fatal("please input FaVe api endpoint")
	}

	if collection == nil || *collection == "" {
		log.Fatal("please input collection name")
	}

	// create FaVe client
	cfg := swagger.NewConfiguration()
	cfg.BasePath = *fave
	client := swagger.NewAPIClient(cfg)
	fmt.Println("client created")

	// create collection
	indexes := make(map[string]interface{})
	indexes["title"] = "string"
	indexes["fullURL"] = "string"
	msg, resp, err := client.DefaultApi.FaveCreateCollection(context.Background(), swagger.Collection{Name: *collection, Indexes: indexes})
	if err != nil {
		log.Fatal(err, resp.StatusCode, msg)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Fatal("failed to create collection")
	}
	fmt.Println(*collection, "collection created")

	var documents = make([]swagger.Document, 0)

	// open zim
	z, err := zim.NewReader(*zimPath, false)
	if err != nil {
		log.Fatal(err)
	}

	// read the zim file
	z.ListTitlesPtrIterator(func(idx uint32) {
		a, err := z.ArticleAtURLIdx(idx)
		if err != nil || a.EntryType == zim.DeletedEntry {
			return
		}

		redirectURL := ""
		data, err := a.Data()
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(data) == 0 {
			return
		}
		title := a.Title
		if title == "" {
			title = filepath.Base(a.FullURL())
		}
		var props = make(swagger.PropertySchema, 0)
		props["title"] = title
		props["namespace"] = string(a.Namespace)
		props["fullURL"] = a.FullURL()
		props["mimeType"] = a.MimeType()
		props["entryType"] = fmt.Sprintf("%d", a.EntryType)

		if a.MimeType() == "text/html" {
			p := bluemonday.StripTagsPolicy()
			html := p.Sanitize(string(data))
			props["content"] = data

			// Tokenize the article content
			text, err := html2text.FromString(html, html2text.Options{TextOnly: true})
			if err != nil {
				log.Fatal(err.Error())
			}
			props["rawText"] = text
		} else {
			// process other files
			props["content"] = data
		}

		if a.EntryType == zim.RedirectEntry {
			ridx, err := a.RedirectIndex()
			if err != nil {
				return
			}
			ra, err := z.ArticleAtURLIdx(ridx)
			if err != nil {
				return
			}
			redirectURL = ra.FullURL()
		}
		props["redirectURL"] = redirectURL

		// Process the documents to be stored on FaVe
		doc := swagger.Document{
			Id:         uuid.New().String(),
			Properties: props,
		}
		documents = append(documents, doc)
	})

	// upload the documents on FaVe
	rqst := swagger.AddDocumentsRequest{
		Documents:         documents,
		Name:              *collection,
		PropertiesToIndex: []string{"rawText"},
	}
	okResp, resp, err := client.DefaultApi.FaveAddDocuments(context.Background(), rqst)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(okResp, resp.StatusCode)
}
