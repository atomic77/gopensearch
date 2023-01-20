package dsl

import (
	"encoding/json"
)

// https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html#search-search-api-example
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html

type Dsl struct {
	Query *Query `json:"query"`
	Size  *int   `json:"size"`
	//
	RawAggs         map[string]Aggregate `json:"aggs"`
	RawAggregations map[string]Aggregate `json:"aggregations"`
	Aggs            map[string]Aggregate

	Sort []map[string]Sort `json:"sort"`
}

type Query struct {
	// Better to use json.RawMessage ?
	RawMatch map[string]interface{} `json:"match"`
	Match    map[string]Match

	RawTerm map[string]interface{} `json:"term"`
	Term    map[string]Term

	Bool  *Bool            `json:"bool"`
	Range map[string]Range `json:"range"`

	QueryString *QueryString `json:"query_string"`
}

type Term struct {
	Value           string   `json:"value"`
	Boost           *float64 `json:"fuzziness"`
	CaseInsensitive bool     `json:"operator"`
}
type Match struct {
	Query     string `json:"query"`
	Fuzziness string `json:"fuzziness"`
	Operator  string `json:"operator"`
}

type Bool struct {
	// RawMust json.RawMessage `json:"must"`
	RawMust   json.RawMessage `json:"must"`
	Must      []Query
	RawShould json.RawMessage `json:"should"`
	Should    []Query
	// Filter is similar to must, but not relevant to scoring, which we don't really use for
	// the moment anyway so it's basically the same.
	Filter []Query `json:"filter"`
}

type Range struct {
	Gt *json.Number `json:"gt"`
	// TODO Post-process to copy from -> gt since they're equivalent
	// from == gt
	// to == lt
	From   *json.Number `json:"from"`
	Gte    *json.Number `json:"gte"`
	Lt     *json.Number `json:"lt"`
	To     *json.Number `json:"to"`
	Lte    *json.Number `json:"lte"`
	Format *string      `json:"format"`
	// These have been deprecated since version 0.9 (!) but some clients
	// in the wild still depend on them.
	// https://github.com/elastic/elasticsearch/issues/48538
	IncludeLower bool   `json:"include_lower"`
	IncludeUpper bool   `json:"include_upper"`
	Boost        string `json:"boost"`
}

// https://www.elastic.co/guide/en/elasticsearch/reference/current/sort-search-results.html
type Sort struct {
	Order string `json:"order"`
	Mode  string `json:"mode"`
}

// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-query-string-query.htmlj
type QueryString struct {
	Query           string `json:"query"`
	AnalyzeWildcard bool   `json:"analyze_wildcard"`
	DefaultField    string `json:"default_field"`
}
