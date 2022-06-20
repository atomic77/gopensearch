package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DbLocation string
}

type Server struct {
	db     *sql.DB
	Router *mux.Router
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
	s.Router = mux.NewRouter()
	s.Router.HandleFunc("/{index:[a-zA-Z0-9\\-]+}", s.CreateIndexHandler).Methods("PUT")
	s.Router.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_create", s.IndexDocumentHandler).Methods("POST")
	s.Router.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_search", s.SearchDocumentHandler).Methods("POST")
	s.Router.HandleFunc("/{index:[a-zA-Z0-9\\-]+}/_bulk", s.BulkHandler).Methods("POST")
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

	docs := s.SearchItem(index, q)
	j, _ := json.Marshal(docs)
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
