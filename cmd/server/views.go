package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"

	faveApi "github.com/onepeerlabs/w3kipedia/pkg/fave_api"
)

var (
	client *faveApi.APIClient
)

func init() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if fave == nil || *fave == "" {
		log.Fatal("please input FaVe api endpoint")
	}

	if collection == nil || *collection == "" {
		log.Fatal("please input collection name")
	}

	cfg := faveApi.NewConfiguration()
	cfg.BasePath = *fave
	client = faveApi.NewAPIClient(cfg)
}

type ArticleIndex struct {
	Title    string
	FullURL  string
	MimeType string
}

func cacheLookup(url string) (*CachedResponse, bool) {
	if v, ok := cache.Get(url); ok {
		c := v.(CachedResponse)
		return &c, ok
	}
	return nil, false
}

// dealing with cached response, responding directly
func handleCachedResponse(cr *CachedResponse, w http.ResponseWriter, r *http.Request) {
	if cr.ResponseType == RedirectResponse {
		log.Printf("302 from %s to %s\n", r.URL.Path, "wiki/"+string(cr.Data))
		http.Redirect(w, r, "/wiki/"+string(cr.Data), http.StatusMovedPermanently)
	} else if cr.ResponseType == NoResponse {
		log.Printf("404 %s\n", r.URL.Path)
		http.NotFound(w, r)
	} else if cr.ResponseType == DataResponse {
		w.Header().Set("Content-Type", "text/markdown; charset=UTF-8")
		// 15 days
		w.Header().Set("Cache-control", "public, max-age=1350000")
		_, err := w.Write(cr.Data)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// the handler receiving http request
func wikiHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[6:]
	// lookup in the cache for a cached response
	if cr, iscached := cacheLookup(url); iscached {
		handleCachedResponse(cr, w, r)
		return
	} else {
		d, resp, err := client.DefaultApi.FaveGetDocuments(context.Background(), "id", url, *collection)
		if err != nil {
			cache.Add(url, CachedResponse{ResponseType: NoResponse})
			return
		}
		if resp.StatusCode != 200 {
			cache.Add(url, CachedResponse{ResponseType: NoResponse})
			return
		}
		props := *d.Properties
		props["title"] = url
		//mime := fmt.Sprintf("%v", d.Properties["mimeType"])
		//entryType := d.Properties["entryType"]
		//redirect := fmt.Sprintf("%v", d.Properties["redirect"])
		//
		//if entryType == fmt.Sprintf("%d", RedirectEntry) && redirect != "" {
		//	cache.Add(url, CachedResponse{
		//		ResponseType: RedirectResponse,
		//		Data:         []byte(redirect),
		//	})
		//	if cr, iscached := cacheLookup(url); iscached {
		//		handleCachedResponse(cr, w, r)
		//	}
		//	return
		//}
		cache.Add(url, CachedResponse{
			ResponseType: DataResponse,
			Data:         []byte(props["rawText"].(string)),
		})
		// look again in the cache for the same entry
		if cr, iscached := cacheLookup(url); iscached {
			handleCachedResponse(cr, w, r)
		}
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("search_data")
	nr := faveApi.NearestDocumentsRequest{
		Text:     q,
		Name:     *collection,
		Distance: 1,
		Limit:    4,
	}
	nResp, _, err := client.DefaultApi.FaveGetNearestDocuments(context.Background(), nr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	d := map[string]interface{}{}
	if len(nResp.Documents) > 0 {

		// Constructs a list of Hits
		var l []map[string]string

		for _, h := range nResp.Documents {
			a := &ArticleIndex{}
			props := *h.Properties
			a.Title = props["title"].(string)

			l = append(l, map[string]string{
				"Score": strconv.FormatFloat(props["distance"].(float64), 'f', 1, 64),
				"Title": a.Title,
				"URL":   "/wiki/" + h.Id,
			})

		}
		d["Hits"] = l

	} else {
		d["Hits"] = 0
	}

	if err := templates.ExecuteTemplate(w, "searchResult.html", d); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// browseHandler is browsing the wikis page per page
func browseHandler(w http.ResponseWriter, r *http.Request) {
	var page, previousPage, nextPage int

	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	Articles := []*ArticleIndex{}

	if page == 0 {
		previousPage = 0
	} else {
		previousPage = page - 1
	}

	nextPage = page + 1

	d := map[string]interface{}{
		"Page":         page,
		"PreviousPage": previousPage,
		"NextPage":     nextPage,
		"Articles":     Articles,
	}
	if err := templates.ExecuteTemplate(w, "index.html", d); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
