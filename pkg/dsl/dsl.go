package dsl

import (
	"encoding/json"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Another useful example with a json-like custom DSL
// https://github.com/alecthomas/participle/discussions/207

var (
	dslLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Number", `\d+`},
		{"String", `"[^"]*"`},
		{"Boolean", `\w+`},
		{"Whitespace", `\s+`},
		{"Punct", `[,.<>(){}=:\[\]]`},
		{"Comment", `//.*`},
	},
		lexer.MatchLongest())

	DslParser = participle.MustBuild(&Dsl{},
		participle.Lexer(dslLexer),
		participle.Unquote("String"),
		// Union type is not exported properly?? Disable this for now
		// and fix the type of each field
		// participle.Union[Value](String{}, Number{}),
		participle.Elide("Whitespace"),
	)
)

type Dsl struct {
	Query *Query       `"{" ( "query" ":" @@ ","?`
	Size  *int         `| "size" ":" @Number ","?`
	Aggs  []*Aggregate `| ( "aggs" | "aggregations" ) ":" "{" @@* "}" ","?`
	Sort  []*Sort      `| "sort" ":" "[" @@* "]" ","?)+ "}"`
}

// https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html#search-search-api-example
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html

type Query struct {
	Term  *Term  `"{" ( "term" ":" "{" @@ "}"`
	Match *Match `| "match" ":" "{" @@ "}"`
	Range *Range `| "range" ":" "{" @@ "}"`
	Bool  *Bool  `| "bool" ":" "{" @@ "}" ) "}"`
}

type Bool struct {
	Must   []*Must   `( "must" ":" "["? @@* "]"?`
	Should []*Should `| "should" ":" "["? @@* "]"? )`
}

type Must struct {
	Query *Query `@@ ","?`
}

type Should struct {
	Query *Query `@@ ","?`
}

type JTerm struct {
	Fields map[string]string
}

type Term struct {
	Properties []*Property `@@*`
}

type Match struct {
	Properties []*Property `@@*`
}

/* Take another try at using base json parsing */

type JDsl struct {
	Query JQuery `json:"query"`
	Size  int    `json:"size"`
	//
	RawAggs         map[string]JAggregate `json:"aggs"`
	RawAggregations map[string]JAggregate `json:"aggregations"`
	Aggs            map[string]JAggregate

	Sort []map[string]JSort `json:"sort"`
}

type JQuery struct {
	// Better to use json.RawMessage ?
	RawMatch map[string]interface{} `json:"match"`
	Match    map[string]JMatch

	// TODO This also needs to be able to handle shorthand forms
	// that can provide additional properties like boost:
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-term-query.html
	Term  map[string]string `json:"term"`
	Bool  *JBool            `json:"bool"`
	Range map[string]JRange `json:"range"`
}

type JMatch struct {
	Query     string `json:"query"`
	Fuzziness string `json:"fuzziness"`
	Operator  string `json:"operator"`
}

type JBool struct {
	// RawMust json.RawMessage `json:"must"`
	Must []JQuery `json:"must"`
	// Should []*Should `| "should" ":" "["? @@* "]"? )`
}

type JRange struct {
	Gt json.Number `json:"gt"`
	// TODO Post-process to copy from -> gt since they're equivalent
	// from == gt
	// to == lt
	From   json.Number `json:"from"`
	Gte    json.Number `json:"gte"`
	Lt     json.Number `json:"lt"`
	To     json.Number `json:"to"`
	Lte    json.Number `json:"lte"`
	Format string      `json:"format"`
	// These have been deprecated since version 0.9 (!) but some clients
	// in the wild still depend on them.
	// https://github.com/elastic/elasticsearch/issues/48538
	IncludeLower bool   `json:"include_lower"`
	IncludeUpper bool   `json:"include_upper"`
	Boost        string `json:"boost"`
}

type Range struct {
	Field        string       `@String ":"`
	RangeOptions RangeOptions `"{" @@ "}" `
}

type RangeOptions struct {
	Gt     *string `( ( "gt" | "from" ) ":" @( String | Number )","?`
	Gte    *string `| "gte" ":" @( String | Number )","?`
	Lt     *string `| ( "lt" | "to" ) ":" @( String | Number )","?`
	Lte    *string `| "lte" ":" @( String | Number )","?`
	Format *string `| "format" ":" @String ","?`
	// These have been deprecated since version 0.9 (!) but some clients
	// in the wild still depend on them.
	// https://github.com/elastic/elasticsearch/issues/48538
	IncludeLower *bool   `| "include_lower" ":" @Boolean ","?`
	IncludeUpper *bool   `| "include_upper" ":" @Boolean ","?`
	Boost        *string `| "boost" ":" @String )+`
}

type Sort struct {
	// Pos   lexer.Position
	Field     string    `"{" @String ":"`
	SortOrder SortOrder `"{" @@ "}" "}"`
}

// https://www.elastic.co/guide/en/elasticsearch/reference/current/sort-search-results.html
type JSort struct {
	Order string `json:"order"`
	Mode  string `json:"mode"`
}

type SortOrder struct {
	Order string `"order" ":" @String`
}

type Property struct {
	// Pos   lexer.Position
	Key   string `@String ":"`
	Value string `@( String | Number ) ","?`
}
