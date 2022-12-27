package dsl

type Aggregate struct {
	Terms         *AggTerms      `json:"terms"`
	DateHistogram *DateHistogram `json:"date_histogram"`
	// AutoDateHistogram *AutoDateHistogram `| "auto_date_histogram" ":" "{" @@ "}" ","?`
	Aggs map[string]Aggregate `json:"aggregations"`
	Avg  *AggField            `json:"avg"`
	Max  *AggField            `json:"max"`
}

type AggregationCategory int

const (
	MetricsSingle AggregationCategory = 1 << iota
	MetricsMultiple
	Bucket
	Pipeline
)

type AggField struct {
	Field   string `json:"field"`
	Missing string `json:"missing"`
}

type AggTerms struct {
	Field string `json:"field"`
	Size  int    `json:"size"`
	// FIXME This should support multiple
	// Order *Property `| "order" ":" "[" "{" @@ "}" "]" ","? )+`
	// Term  *Term  `"{" ( "term" ":" "{" @@ "}"`
}

type DateHistogram struct {
	Field            string `json:"field"`
	Buckets          int    `json:"buckets"`
	FixedInterval    string `json:"fixed_interval"`
	CalendarInterval string `json:"calendar_interval"`
}

/* TODO Implement
type AutoDateHistogram struct {
	Field           string `( "field" ":" @String ","?`
	Buckets         int    `| "buckets" ":" @Number ","?`
	MinimumInterval string `| "buckets" ":" @String ","? )+`
}
*/

func (a *Aggregate) GenAggregationCategory() AggregationCategory {
	// There must be a better way to do this... ?

	if a.Avg != nil ||
		a.Max != nil {
		return MetricsSingle
	} else if a.Terms != nil {
		return Bucket
	}
	return 0

}
