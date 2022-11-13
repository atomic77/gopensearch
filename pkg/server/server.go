package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DbLocation string
}

type Server struct {
	db     *sql.DB
	Router http.Handler
	Cfg    Config
}

type Document struct {
	Id      int                    `json:"id"`
	Content map[string]interface{} `json:"_source"`
}

func openDb(loc string) *sql.DB {
	d, err := sql.Open("sqlite3", loc)
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

	// Administrative functions
	r.HandleFunc("/", s.HeadHandler).Methods("HEAD")
	r.HandleFunc("/", s.ClusterStatusHandler).Methods("GET")

	// Template-related
	r.HandleFunc("/_template/{index:[a-zA-Z0-9\\-]+}", s.CreateTemplateHandler).Methods("PUT")

	r.PathPrefix("/").HandlerFunc(s.DefaultHandler)

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	s.Router = loggedRouter
}

func (s *Server) Init() {
	s.db = openDb(s.Cfg.DbLocation)
	s.registerRoutes()
}

type IndexDocumentResponse struct {
	// TODO Add shards:
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html#create-document-ids-automatically
	Index   string `json:"_index"`
	Id      int    `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

func (s *Server) IndexDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["index"]

	b, _ := io.ReadAll(r.Body)
	err := s.IndexDocument(string(b), index)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failure "+err.Error())
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

type CreateIndexResponse struct {
	Acknowledged       bool   `json:"acknowledged"`
	ShardsAcknowledged bool   `json:"shards_acknowledged"`
	Index              string `json:"index"`
}

func (s *Server) CreateIndexHandler(w http.ResponseWriter, r *http.Request) {
	// PUT /<index>  - creates a new index
	vars := mux.Vars(r)
	index, ok := vars["index"]
	if !ok {
		fmt.Fprintf(w, "index name is missing in parameters")
	}
	s.CreateTable(index)
	resp := CreateIndexResponse{
		Acknowledged:       true,
		ShardsAcknowledged: true,
		Index:              index,
	}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

type SearchResponse struct {
	Took         int                    `json:"took"`
	TimedOut     bool                   `json:"timed_out"`
	Shards       ShardsInfo             `json:"_shards"`
	Hits         *Hits                  `json:"hits"`
	Aggregations map[string]Aggregation `json:"aggregations,omitempty"`
}

type Aggregation interface {
	GetAggregateCategory() dsl.AggregationCategory
	SerializeResultset(rows *sql.Rows)
}

type Hits struct {
	Total int        `json:"total"`
	Hits  []Document `json:"hits"`
}
type MetricSingleAggregation struct {
	Value float64 `json:"value"`
}

func (m MetricSingleAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.MetricsSingle
}

type MetricMultipleAggregation struct {
	Values []float64 `json:"values"`
}

func (m MetricMultipleAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.MetricsMultiple
}

type BucketAggregation struct {
	DocCountErrorUpperBound int      `json:"doc_count_error_upper_bound"`
	Buckets                 []Bucket `json:"buckets"`
}

func (m BucketAggregation) GetAggregateCategory() dsl.AggregationCategory {
	return dsl.Bucket
}

type Bucket struct {
	KeyAsString string `json:"key_as_string,omitempty"`
	Key         string `json:"key"`
	DocCount    int    `json:"doc_count"`
}

func (s *Server) SearchDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["index"]

	buf := new(strings.Builder)
	_, err := io.Copy(buf, r.Body)

	q := &dsl.Dsl{}
	dsl.DslParser.ParseString("", buf.String(), q)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failure trying to parse "+err.Error())
		return
	}

	docs, aggs, err := s.SearchItem(index, q)
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
	for i, a := range q.Aggs {
		sr.Aggregations[a.Name] = aggs[i]
	}
	j, _ := json.Marshal(sr)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

type ShardsInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

func MakeShardsInfo() ShardsInfo {
	s := ShardsInfo{Total: 1, Successful: 1, Failed: 1}
	return s
}

// Not fully implemented
type BulkResponseItem struct {
	Index       string     `json:"_index"`
	Id          string     `json:"_id"`
	Type        string     `json:"_type"`
	Version     int        `json:"_version"`
	Result      string     `json:"result"`
	SeqNo       int        `json:"_seq_no"`
	Status      int        `json:"status"`
	PrimaryTerm int        `json:"_primary_term"`
	Shards      ShardsInfo `json:"_shards"`
	// Error       map[string]string `json:"error"`
}

type BulkResponse struct {
	Took   int                           `json:"took"`
	Errors bool                          `json:"errors"`
	Items  []map[string]BulkResponseItem `json:"items"`
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

				s.IndexDocument(string(b), index)
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
