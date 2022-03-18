package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	zim "github.com/akhenakh/gozim"
	"github.com/blevesearch/bleve"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/jdkato/prose/v2"
	"github.com/microcosm-cc/bluemonday"
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

type kv struct {
	Key   string
	Value int
}

var (
	indexPath     = flag.String("index", "", "path for the index file")
	beeHost       = flag.String("bee", "", "bee API endpoint")
	beeIsProxy    = flag.Bool("proxy", false, "if Bee endpoint is gateway proxy")
	batch         = flag.String("batch", "", "bee Postage Stamp ID")
	zimPath       = flag.String("zim", "", "zim file location")
	indexContent  = flag.Bool("content", false, "whether to generate tags  from content for indexing (indexing process will be faster if false)")
	offline       = flag.Bool("offline", false, "run server offline for listing only")
	shouldEncrypt = flag.Bool("encrypt", false, "encrypt content while uploading into swarm")
	help          = flag.Bool("help", false, "print help")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

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
	var b blockstore.Client
	if *offline {
		b = mock.NewMockBeeClient()
	} else {
		logger := logging.New(os.Stdout, logrus.ErrorLevel)
		b = bee.NewBeeClient(
			*beeHost,
			*batch,
			logger,
		)
	}
	if !b.CheckConnection(*beeIsProxy) {
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

	var index bleve.Index
	_, err := os.Lstat(*indexPath)
	if os.IsNotExist(err) {
		index, err = bleve.New(*indexPath, mapping)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		index, err = bleve.Open(*indexPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer index.Close()

	// open zim
	z, err := zim.NewReader(*zimPath, false)
	if err != nil {
		log.Fatal(err)
	}

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
		tags := ""
		if a.MimeType() == "text/html" && *indexContent {
			p := bluemonday.StripTagsPolicy()
			html := p.SanitizeBytes(data)
			doc, err := prose.NewDocument(string(html))
			if err != nil {
				log.Fatal(err)
			}

			// Iterate over the doc's tokens:
			tags = strings.Join(mostWords(doc.Tokens(), 10), " ")
		}
		if len(data) == 0 {
			return
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
		title := a.Title
		if title == "" {
			title = filepath.Base(a.FullURL())
		}
		address, err := b.UploadBlob(data, true, *shouldEncrypt)
		if err != nil {
			fmt.Printf("Failed to upload %s : %s\n", a.FullURL(), err.Error())
			return
		}
		fmt.Println(a.FullURL(), hex.EncodeToString(address), len(data))
		idoc := article{
			Title:       title,
			Namespace:   string(a.Namespace),
			FullURL:     a.FullURL(),
			MimeType:    a.MimeType(),
			EntryType:   fmt.Sprintf("%d", a.EntryType),
			Address:     hex.EncodeToString(address),
			RedirectURL: redirectURL,
			Content:     tags,
		}
		err = index.Index(a.FullURL(), idoc)
		if err != nil {
			log.Fatal(err.Error())
		}
	})
	indexMeta, err := os.ReadFile(filepath.Join(*indexPath, "index_meta.json"))
	if err != nil {
		log.Fatal(err)
	}
	indexMetaAddress, err := b.UploadBlob(indexMeta, true, *shouldEncrypt)
	if err != nil {
		fmt.Printf("Failed to upload index meta : %s\n", err.Error())
		return
	}
	fmt.Println("index meta hash : ", hex.EncodeToString(indexMetaAddress))

	indexStore, err := os.ReadFile(filepath.Join(*indexPath, "store"))
	if err != nil {
		log.Fatal(err)
	}
	indexStoreAddress, err := b.UploadBlob(indexStore, true, *shouldEncrypt)
	if err != nil {
		fmt.Printf("Failed to upload index store : %s\n", err.Error())
		return
	}
	fmt.Println("index store hash :", hex.EncodeToString(indexStoreAddress))
}

func mostWords(input []prose.Token, count int) (top []string) {
	top = make([]string, count)
	var ss []kv
	wc := make(map[string]int)
	for _, tok := range input {
		if tok.Tag == "NNP" && len(tok.Text) > 2 {
			_, matched := wc[tok.Text]
			if matched {
				wc[tok.Text] += 1
			} else {
				wc[tok.Text] = 1
			}
		}
	}
	for k, v := range wc {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	limit := count
	if len(ss) < count {
		limit = len(ss)
	}
	for i := 0; i < limit; i++ {
		top[i] = ss[i].Key
	}
	return top
}
