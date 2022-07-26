package dsl

type Aggregate struct {
	Name          string        `@String ":"`
	AggregateType AggregateType `"{" @@ "}" ","?`
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
	Terms *AggTerms `( "terms" ":" "{" @@ "}" ","?`
	Avg   *AggField `| "avg" ":" "{" @@ "}" ","?`
	Max   *AggField `| "max" ":" "{" @@ "}" ","? )`
}

type AggField struct {
	Field   string `( "field" ":" @String ","?`
	Missing string `| "missing" ":" @(String | Number) ","? )+`
}

type AggTerms struct {
	Field string `( "field" ":" @String ","?`
	Size  int    `| "size" ":" @Number ","? )+`
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
