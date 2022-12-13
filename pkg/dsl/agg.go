package dsl

type Aggregate struct {
	Name          string           `@String ":"`
	AggregateType []*AggregateType `"{" @@* "}" ","?`
}

type AggregationCategory int

const (
	MetricsSingle AggregationCategory = 1 << iota
	MetricsMultiple
	Bucket
	Pipeline
)

type AggregateType struct {
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-aggregations.html
	// Many to be implemented...
	Terms             *AggTerms          `( "terms" ":" "{" @@ "}" ","?`
	DateHistogram     *DateHistogram     `| "date_histogram" ":" "{" @@ "}" ","?`
	AutoDateHistogram *AutoDateHistogram `| "auto_date_histogram" ":" "{" @@ "}" ","?`
	Aggs              []*Aggregate       `| ( "aggs" | "aggregations" ) ":" "{" @@* "}" ","?`
	Avg               *AggField          `| "avg" ":" "{" @@ "}" ","?`
	Max               *AggField          `| "max" ":" "{" @@ "}" ","? )`
}

type AggField struct {
	Field   string `( "field" ":" @String ","?`
	Missing string `| "missing" ":" @(String | Number) ","? )+`
}

type AggTerms struct {
	Field string `( "field" ":" @String ","?`
	Size  int    `| "size" ":" @Number ","? `
	// FIXME This should support multiple
	Order *Property `| "order" ":" "[" "{" @@ "}" "]" ","? )+`
	// Term  *Term  `"{" ( "term" ":" "{" @@ "}"`
}

type Order struct {
	Property *Property `@@`
	// Properties []*Property `@@*`
}

type DateHistogram struct {
	Field            string `( "field" ":" @String ","?`
	Buckets          int    `| "buckets" ":" @Number ","?`
	FixedInterval    string `| "fixed_interval" ":" @String ","?`
	CalendarInterval string `| "calendar_interval" ":" @String ","? )+`
}

type AutoDateHistogram struct {
	Field           string `( "field" ":" @String ","?`
	Buckets         int    `| "buckets" ":" @Number ","?`
	MinimumInterval string `| "buckets" ":" @String ","? )+`
}

func (a *AggregateType) GenAggregationCategory() AggregationCategory {
	// There must be a better way to do this... ?

	if a.Avg != nil ||
		a.Max != nil {
		return MetricsSingle
	} else if a.Terms != nil {
		return Bucket
	}
	return 0

}
