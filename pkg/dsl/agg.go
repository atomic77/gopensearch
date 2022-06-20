package dsl

type Aggregate struct {
	Name        string        `@String ":"`
	Aggregation AggregateType `"{" @@ "}"`
}

type AggregateType struct {
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-aggregations.html
	// Many to be implemented...
	Terms *AggTerms `"terms" ":" "{" @@ "}" ","?`
}

type AggTerms struct {
	Field string `( "field" ":" @String ","?`
	Size  int    `| "size" ":" @Number ","? )+`
}
