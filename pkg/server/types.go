package server

import (
	"net/http"

	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DbLocation string
	ListenAddr string
	Port       int
}

type Server struct {
	db               *sqlx.DB
	Router           http.Handler
	Cfg              Config
	TemplateMappings []TemplateMapping
}

type Document struct {
	Id      int                    `json:"id"`
	Content map[string]interface{} `json:"_source"`
}
type Bucket struct {
	KeyAsString   string `json:"key_as_string,omitempty"`
	Key           string `json:"key"`
	DocCount      int64  `json:"doc_count"`
	subaggregates map[string]interface{}
}

func makeBucket() Bucket {
	b := Bucket{}
	b.subaggregates = make(map[string]interface{})
	return b
}

type BucketAggregation struct {
	DocCountErrorUpperBound int      `json:"doc_count_error_upper_bound"`
	Buckets                 []Bucket `json:"buckets"`
}
type MetricMultipleAggregation struct {
	Values []float64 `json:"values"`
}

type IndexDocumentResponse struct {
	// TODO Add shards:
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html#create-document-ids-automatically
	Index   string `json:"_index"`
	Id      int    `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}
type CreateIndexResponse struct {
	Acknowledged       bool   `json:"acknowledged"`
	ShardsAcknowledged bool   `json:"shards_acknowledged"`
	Index              string `json:"index"`
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
	SerializeResultset(rows *sqlx.Rows, dbq *dbSubQuery)
}

type Hits struct {
	Total int        `json:"total"`
	Hits  []Document `json:"hits"`
}
type MetricSingleAggregation struct {
	Value float64 `json:"value"`
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

type BulkResponse struct {
	Took   int                           `json:"took"`
	Errors bool                          `json:"errors"`
	Items  []map[string]BulkResponseItem `json:"items"`
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

type MSearchHeader struct {
	IgnoreUnavailable *bool   `json:"ignore_unavailable"`
	Index             *string `json:"index"`
	// Can't find any documentation about this, but appears to be supported by ES
	// in use in the wild
	Indices []*string `json:"indices"`
	// TODO More options to implement
}

type MSearchResponse struct {
	Took      int               `json:"took"`
	Responses []*SearchResponse `json:"responses"`
	// TODO More options to implement
}
