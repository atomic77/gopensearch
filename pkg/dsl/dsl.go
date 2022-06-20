package dsl

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Another useful example with a json-like custom DSL
// https://github.com/alecthomas/participle/discussions/207

var (
	dslLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Number", `\d+`},
		{"String", `"[^"]*"`},
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
	Query *Query `"{" ( "query" ":" @@ ","?`
	Size  *int   `| "size" ":" @Number ","?`
	Sort  *Sort  `| "sort" ":" "{" @@ "}" ","?)+ "}"`
}

// https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html#search-search-api-example
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html

type Query struct {
	Term  *Term  `"{" ( "term" ":" "{" @@ "}"`
	Match *Match `| "match" ":" "{" @@ "}"`
	Bool  *Bool  `| "bool" ":" "{" @@ "}" ) "}"`
}

type Bool struct {
	Must   *Must   `( "must" ":" "["? @@ "]"?`
	Should *Should `| "should" ":" "["? @@ "]"? )`
}

type Must struct {
	Queries []*Query `@@* ","?`
}

type Should struct {
	Queries []*Query `@@* ","?`
}

type Term struct {
	Properties []*Property `@@*`
}

type Match struct {
	Properties []*Property `@@*`
}

type Sort struct {
	// Pos   lexer.Position
	Value string `@String`
}
type Property struct {
	// Pos   lexer.Position
	Key   string `@String ":"`
	Value string `@String ","?`
}
