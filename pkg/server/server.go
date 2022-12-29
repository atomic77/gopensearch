package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

func openDb(loc string) *sqlx.DB {
	d, err := sqlx.Open("sqlite3", loc)
	if err != nil {
		panic(err)
	}
	if d == nil {
		panic("db nil")
	}
	return d
}

func (s *Server) registerRoutes() {
	r := mux.NewRouter()
	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+}", s.CreateIndexHandler).Methods("PUT")
	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_create", s.IndexDocumentHandler).Methods("POST")
	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_search", s.SearchDocumentHandler).Methods("POST")
	// FIXME Fix this when adding support for searching over multiple indices
	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+},{[a-zA-Z0-9\\-]+},{[a-zA-Z0-9\\-]+}/_search", s.SearchDocumentHandler).Methods("POST")
	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_bulk", s.BulkHandler).Methods("POST")
	r.HandleFunc("/_bulk", s.BulkHandler).Methods("POST")

	r.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_msearch", s.MSearchHandler).Methods("GET")
	r.HandleFunc("/_msearch", s.MSearchHandler).Methods("GET")

	// Administrative functions
	r.HandleFunc("/", s.HeadHandler).Methods("HEAD")
	r.HandleFunc("/", s.ClusterStatusHandler).Methods("GET")
	r.HandleFunc("/_cat/indices", s.CatalogIndicesHandler).Methods("GET")

	// Template-related
	r.HandleFunc("/_template/{index:[a-zA-Z0-9\\-]+}", s.CreateTemplateHandler).Methods("PUT")

	r.PathPrefix("/").HandlerFunc(s.DefaultHandler)
	if s.Cfg.Debug {
		r.Use(debugMiddleware)
	}

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	s.Router = loggedRouter
}

func (s *Server) Init() {
	s.db = openDb(s.Cfg.DbLocation)
	s.registerRoutes()
	s.createMetadata()
	s.loadTemplateMetadata()
}

func debugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		b, _ := io.ReadAll(r.Body)
		log.Println("Request", r.RequestURI)
		log.Println("Complete body: ", string(b))
		rdr := bytes.NewBuffer(b)
		r.Body = io.NopCloser(rdr)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) IndexDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["index"]

	// Check if we need to implicitly create this index
	idxMap, err := s.ListTables()
	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	if _, ok := idxMap[index]; !ok {
		err = s.CreateTable(index)
		if err != nil {
			handleErrorResponse(w, err)
			return
		}
	}

	b, _ := io.ReadAll(r.Body)
	err = s.IndexDocument(string(b), index)
	if err != nil {
		handleErrorResponse(w, err)
		return
	}
	resp := IndexDocumentResponse{
		Index: index,
		// TODO Check if we can easily get back the rowid after the insert
		Id:      0,
		Version: 1,
		Result:  "created",
	}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (s *Server) CreateIndexHandler(w http.ResponseWriter, r *http.Request) {
	// PUT /<index>  - creates a new index
	vars := mux.Vars(r)
	index, ok := vars["index"]
	if !ok {
		fmt.Fprintf(w, "index name is missing in parameters")
	}
	err := s.CreateTable(index)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failure "+err.Error())
		return
	}
	resp := CreateIndexResponse{
		Acknowledged:       true,
		ShardsAcknowledged: true,
		Index:              index,
	}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (m MetricSingleAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.MetricsSingle
}

func (m MetricMultipleAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.MetricsMultiple
}

func (m BucketAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.Bucket
}

func (bkt *Bucket) MarshalJSON() ([]byte, error) {
	// Found this approach here: https://stackoverflow.com/a/59923606
	// feels like there has to be a better way to optionally tack on some
	// extra stuff to the JSON we serialize by default...
	type Bucket_ Bucket
	b, err := json.Marshal(Bucket_(*bkt))
	if err != nil {
		return nil, err
	}
	if len(bkt.subaggregates) > 0 {
		s, err := json.Marshal(bkt.subaggregates)
		if err != nil {
			return nil, err
		}
		b[len(b)-1] = ','
		b = append(b, s[1:]...)
	}
	return b, nil
}

func (s *Server) SearchDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["index"]

	buf, _ := io.ReadAll(r.Body)
	q := &dsl.Dsl{}
	// err = dsl.DslParser.ParseString("", buf.String(), q)
	err := json.Unmarshal(buf, &q)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if r.Header.Get("X-Gopensearch-Dsl-Dump") != "" {
		log.Println(repr.String(q))
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failure trying to parse "+err.Error())
		return
	}

	sr, err := s.getSearchResponse(index, q)
	if err != nil {
		eresp := &GenericErrorResponse{
			Reason:       err.Error(),
			Index:        index,
			ResourceId:   index,
			ResourceType: "index_or_alias",
		}
		j, _ := json.Marshal(eresp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(j)
		return
	}
	j, _ := json.Marshal(sr)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (s *Server) getSearchResponse(index string, q *dsl.Dsl) (*SearchResponse, error) {
	docs, aggs, err := s.SearchItem(index, q)
	if err != nil {
		return nil, err
	}
	sr := &SearchResponse{
		Took:     123,
		TimedOut: false,
		Shards:   MakeShardsInfo(),
		Hits: &Hits{
			Total: len(docs),
			Hits:  docs,
		},
	}
	sr.Aggregations = make(map[string]Aggregation)
	for label, agg := range aggs {
		sr.Aggregations[label] = agg
	}
	return sr, nil
}

/*

Bulk requests can be one of four types :
- index
- create
- update
- delete

https://www.elastic.co/guide/en/elasticsearch/reference/master/docs-bulk.html

The response is somewhat complex and is used by the python bulk helper client:
https://www.elastic.co/guide/en/elasticsearch/reference/master/docs-bulk.html#docs-bulk-api-example

*/

func (s *Server) BulkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["index"]

	// Current table map to see if we need to create any on the fly
	idxMap, err := s.ListTables()
	if err != nil {
		handleErrorResponse(w, err)
		return
	}
	var bulkReq map[string]interface{}
	bulkResp := BulkResponse{
		Took:  123,
		Items: make([]map[string]BulkResponseItem, 0),
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	for {
		err := decoder.Decode(&bulkReq)

		if err == io.EOF {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failure trying to parse "+err.Error())
			return
		}

		keys := make([]string, 0, len(bulkReq))
		for k := range bulkReq {
			keys = append(keys, k)
		}

		switch keys[0] {
		case "index":
			var doc map[string]interface{}
			err = decoder.Decode(&doc)
			if err != nil {
				resp := BulkResponseItem{
					Index:   index,
					Id:      "123",
					Version: 1,
					// Error:   nil,
					Status: 500,
				}
				var respWrapped = map[string]BulkResponseItem{"index": resp}
				bulkResp.Items = append(bulkResp.Items, respWrapped)
			} else {
				b, _ := json.Marshal(doc)

				keys2 := make([]string, 0, len(bulkReq))
				for k := range bulkReq {
					keys2 = append(keys2, k)
				}
				// FIXME Properly parse this structure
				idxType := bulkReq["index"].(map[string]interface{})
				index = idxType["_index"].(string)

				// Check if we need to create a table implicitly
				if _, ok := idxMap[index]; !ok {
					err = s.CreateTable(index)
					if err != nil {
						handleErrorResponse(w, err)
						return
					}
					idxMap[index] = 42
				}
				err = s.IndexDocument(string(b), index)
				if err == nil {

					resp := BulkResponseItem{
						Index:       index,
						Id:          "123",
						Type:        "_doc",
						Version:     1,
						SeqNo:       3,
						PrimaryTerm: 1,
						Result:      "created",
						Shards:      MakeShardsInfo(),
						// Error:   map[string]string{},
						Status: 201,
					}
					var respWrapped = map[string]BulkResponseItem{"index": resp}
					bulkResp.Items = append(bulkResp.Items, respWrapped)
				} else {

					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						fmt.Fprintf(w, "failure during bulk indexing: "+err.Error())
						return
					}

				}
			}

		case "update", "create", "delete":
			// Not implemented
		default:

		}

	}
	bulkResp.Errors = false
	j, _ := json.Marshal(bulkResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

// Similar to bulk handler, but for querying
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-multi-search.html
func (s *Server) MSearchHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	index := vars["index"]

	msearchHeader := MSearchHeader{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	responses := make([]*SearchResponse, 0)

	for {

		// MSearch requests come in the form of a "header" request, a new line,
		// and the standard search query
		err := decoder.Decode(&msearchHeader)

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintf(w, "failure trying to parse "+err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		qDsl := &dsl.Dsl{}
		err = decoder.Decode(&qDsl)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "failure parsing DSL ", err.Error())
			return
		}

		var sr *SearchResponse
		if msearchHeader.Index != nil {
			sr, err = s.getSearchResponse(*msearchHeader.Index, qDsl)
		} else if msearchHeader.Indices != nil {
			// TODO Search only the first for now until we can enable search
			// support against multiple indices seamlessly
			sr, err = s.getSearchResponse(*msearchHeader.Indices[0], qDsl)

		} else {
			sr, err = s.getSearchResponse(index, qDsl)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "failure when searching ", err.Error())
			return
		}

		responses = append(responses, sr)

	}
	msr := MSearchResponse{
		Took:      123,
		Responses: responses,
	}
	j, _ := json.Marshal(msr)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
