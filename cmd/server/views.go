package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"

	zim "github.com/akhenakh/gozim"
	"github.com/blevesearch/bleve"
)

type article struct {
	Title     string
	Namespace string
	Content   string
	FullURL   string
	MimeType  string
	EntryType int
	Address   []byte
}

type ArticleIndex struct {
	Title    string
	FullURL  string
	Address  string
	MimeType string
}

const (
	ArticlesPerPage = 16
)

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
		log.Printf("200 %s\n", r.URL.Path)
		w.Header().Set("Content-Type", cr.MimeType)
		// 15 days
		w.Header().Set("Cache-control", "public, max-age=1350000")
		w.Write(cr.Data)
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
		d, err := index.Document(url)
		if err != nil {
			cache.Add(url, CachedResponse{ResponseType: NoResponse})
			return
		}

		mime := ""
		entryType := ""
		redirect := ""
		for _, v := range d.Fields {
			if v.Name() == "Address" {
				url = string(v.Value())
			} else if v.Name() == "MimeType" {
				mime = string(v.Value())
			} else if v.Name() == "EntryType" {
				entryType = string(v.Value())
			} else if v.Name() == "RedirectURL" {
				redirect = string(v.Value())
			}
		}

		if entryType == fmt.Sprintf("%d", zim.RedirectEntry) && redirect != "" {
			cache.Add(url, CachedResponse{
				ResponseType: RedirectResponse,
				Data:         []byte(redirect),
			})
			if cr, iscached := cacheLookup(url); iscached {
				handleCachedResponse(cr, w, r)
			}
			return
		}
		ref, err := hex.DecodeString(url)
		if err != nil {
			cache.Add(url, CachedResponse{ResponseType: NoResponse})
		} else {
			data, _, err := B.DownloadBlob(ref)
			if err != nil {
				cache.Add(url, CachedResponse{ResponseType: NoResponse})
				return
			}
			if err != nil {
				cache.Add(url, CachedResponse{ResponseType: NoResponse})
			} else {
				cache.Add(url, CachedResponse{
					ResponseType: DataResponse,
					Data:         data,
					MimeType:     mime,
				})
			}
		}
		// look again in the cache for the same entry
		if cr, iscached := cacheLookup(url); iscached {
			handleCachedResponse(cr, w, r)
		}
	}

}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	pageString := r.FormValue("page")
	pageNumber, _ := strconv.Atoi(pageString)
	previousPage := pageNumber - 1
	if pageNumber == 0 {
		previousPage = 0
	}
	nextPage := pageNumber + 1
	q := r.FormValue("search_data")
	d := map[string]interface{}{
		"Query":        q,
		"Path":         "",
		"Page":         pageNumber,
		"PreviousPage": previousPage,
		"NextPage":     nextPage,
	}

	if q == "" {
		if err := templates.ExecuteTemplate(w, "search.html", d); err != nil {
			http.Error(w, err.Error(), 500)
		}

		return
	}

	itemCount := 20
	from := itemCount * pageNumber
	query := bleve.NewQueryStringQuery(q)
	search := bleve.NewSearchRequestOptions(query, itemCount, from, false)
	search.Fields = []string{"Title", "FullURL"}

	sr, err := index.Search(search)
	if err != nil {
		http.Error(w, err.Error(), 500)

		return
	}

	if sr.Total > 0 {
		d["Info"] = fmt.Sprintf("%d matches for query [%s], took %s", sr.Total, q, sr.Took)

		// Constructs a list of Hits
		var l []map[string]string

		for _, h := range sr.Hits {
			a := &ArticleIndex{}
			for otherFieldName, otherFieldValue := range h.Fields {
				if otherFieldName == "Title" {
					a.Title = fmt.Sprintf("%v", otherFieldValue)
				} else if otherFieldName == "FullURL" {
					if a.Title == "" {
						a.Title = fmt.Sprintf("%v", otherFieldValue)
					}
					a.FullURL = fmt.Sprintf("%v", otherFieldValue)
				} else if otherFieldName == "Address" {
					a.Address = fmt.Sprintf("%v", otherFieldValue)
				} else if otherFieldName == "MimeType" {
					a.MimeType = fmt.Sprintf("%v", otherFieldValue)
				}
			}
			l = append(l, map[string]string{
				"Score": strconv.FormatFloat(h.Score, 'f', 1, 64),
				"Title": a.Title,
				"URL":   "/wiki/" + a.FullURL,
			})

		}
		d["Hits"] = l

	} else {
		d["Info"] = fmt.Sprintf("No match for [%s], took %s", q, sr.Took)
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
	query := bleve.NewPhraseQuery([]string{"html"}, "MimeType")
	search := bleve.NewSearchRequestOptions(query, ArticlesPerPage, page*ArticlesPerPage, false)
	search.Fields = []string{"Title", "FullURL", "MimeType"}

	sr, err := index.Search(search)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range sr.Hits {
		a := &ArticleIndex{}
		for otherFieldName, otherFieldValue := range v.Fields {
			if otherFieldName == "Title" {
				a.Title = fmt.Sprintf("%v", otherFieldValue)
			} else if otherFieldName == "FullURL" {
				a.FullURL = fmt.Sprintf("%v", otherFieldValue)
			} else if otherFieldName == "Address" {
				a.Address = fmt.Sprintf("%v", otherFieldValue)
			} else if otherFieldName == "MimeType" {
				a.MimeType = fmt.Sprintf("%v", otherFieldValue)
			}
		}
		if a.Title == "" {
			a.Title = a.FullURL
		}
		Articles = append(Articles, a)
	}
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

func robotHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
}
