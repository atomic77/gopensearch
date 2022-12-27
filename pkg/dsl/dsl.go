package dsl

import (
	"encoding/json"
)

// Another useful example with a json-like custom DSL
// https://github.com/alecthomas/participle/discussions/207

// https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html#search-search-api-example
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html

type Term struct {
	Fields map[string]string
}

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

	// TODO This also needs to be able to handle shorthand forms
	// that can provide additional properties like boost:
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-term-query.html
	Term  map[string]string `json:"term"`
	Bool  *Bool             `json:"bool"`
	Range map[string]Range  `json:"range"`
}

type Match struct {
	Query     string `json:"query"`
	Fuzziness string `json:"fuzziness"`
	Operator  string `json:"operator"`
}

type Bool struct {
	// RawMust json.RawMessage `json:"must"`
	RawMust json.RawMessage `json:"must"`
	Must    []Query
	// Should []*Should `| "should" ":" "["? @@* "]"? )`
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
