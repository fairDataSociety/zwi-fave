package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	swagger "github.com/onepeerlabs/w3kipedia/pkg/fave_api"
)

var (
	zwiPath    = flag.String("zwi", "", "directory that contains zwi files")
	fave       = flag.String("fave", "http://localhost:1234/v1", "FaVe API endpoint")
	collection = flag.String("collection", "", "Collection name to store content in FaVe")
	help       = flag.Bool("help", false, "print help")
)

type Metadata struct {
	Title   string `json:"Title"`
	Content struct {
		ArticleHTML     string `json:"article.html"`
		ArticleWikitext string `json:"article.wikitext"`
		ArticleTxt      string `json:"article.txt"`
	} `json:"Content"`
}

func main() {

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if zwiPath == nil || *zwiPath == "" {
		log.Fatal("please input location for zwi files")
	}

	if fave == nil || *fave == "" {
		log.Fatal("please input FaVe api endpoint")
	}

	if collection == nil || *collection == "" {
		log.Fatal("please input collection name")
	}
	fmt.Println(*zwiPath, *fave, *collection)
	// create FaVe client
	cfg := swagger.NewConfiguration()
	cfg.BasePath = *fave
	client := swagger.NewAPIClient(cfg)
	fmt.Println("client created")

	// create collection
	indexes := []swagger.Index{
		{FieldName: "title", FieldType: "string"},
	}
	msg, resp, err := client.DefaultApi.FaveCreateCollection(context.Background(), swagger.Collection{Name: *collection, Indexes: indexes})
	if err != nil {
		log.Fatal(err, resp, msg)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Fatal("failed to create collection")
	}
	fmt.Println(*collection, "collection created")

	var documents = make([]swagger.Document, 0)

	//get zwi file lists
	entries, err := os.ReadDir(*zwiPath)
	if err != nil {
		fmt.Println("Error opening zwi source file:", err)
		return
	}
	for _, entry := range entries {
		fmt.Println("File:", entry.Name())

		zipFile, err := zip.OpenReader(filepath.Join(*zwiPath, entry.Name()))
		if err != nil {
			fmt.Println("Error opening ZIP file:", err)
			continue
		}
		defer zipFile.Close()

		var props = make(swagger.Property)
		for _, file := range zipFile.File {

			if file.Name == "article.txt" || file.Name == "metadata.json" || file.Name == "article.html" {
				buffer, err := getContent(file)
				if err != nil {
					fmt.Println("Error reading file:", err)
					continue
				}
				switch file.Name {
				case "article.html":
					props["html"] = string(buffer)
				case "article.txt":
					props["rawText"] = string(buffer)
					re := regexp.MustCompile(`\|.*`)
					filteredText := re.ReplaceAllString(string(buffer), "")

					re2 := regexp.MustCompile(`(?m)^This editable Main Article.*$`)
					filteredText = re2.ReplaceAllString(filteredText, "")

					re3 := regexp.MustCompile(`(?m)^This article.*$`)
					filteredText = re3.ReplaceAllString(filteredText, "")

					props["article"] = filteredText
				case "metadata.json":
					metadata := &Metadata{}
					err = json.Unmarshal(buffer, metadata)
					if err != nil {
						fmt.Println("Error unmarshalling JSON:", err)
						continue
					}
					props["title"] = metadata.Title
				}
			}
		}

		if props["article"] == "" {
			log.Println("article.txt not found")
			continue
		}
		if props["title"] == "" {
			log.Println("metadata.json not found in zwi file", entry.Name())
			continue
		}
		if props["html"] == "" {
			log.Println("article.html not found in zwi file", entry.Name())
			continue
		}
		doc := swagger.Document{
			Id:         uuid.New().String(),
			Properties: &props,
		}
		documents = append(documents, doc)
	}

	// upload the documents on FaVe
	rqst := swagger.AddDocumentsRequest{
		Documents:             documents,
		Name:                  *collection,
		PropertiesToVectorize: []string{"article"},
	}
	okResp, resp, err := client.DefaultApi.FaveAddDocuments(context.Background(), rqst)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(okResp, resp.StatusCode)
}

func getContent(file *zip.File) ([]byte, error) {
	fileReader, err := file.Open()
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer fileReader.Close()
	buffer, err := io.ReadAll(fileReader)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}
	return buffer, nil
}
