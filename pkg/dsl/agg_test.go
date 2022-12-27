package dsl

import (
	"encoding/json"
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
)

func TestAggTerms(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	    "aggs":{
	        "generalStatus":{
	            "terms":{"field":"foo"}
			}
	    },
	    "size":0,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}`
	err := json.Unmarshal([]byte(q), &dsl)
	require.Equal(t, dsl.Aggs["generalStatus"].Terms.Field, "foo")
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestAggTermsWithLongName(t *testing.T) {
	dsl := &Dsl{}
	q := `
		{
			"aggregations":{
				"distinct_services":{
					"terms":{
						"field":"serviceName",
						"size":10000
					}
				}
			},
			"size":0
		}
    `
	err := json.Unmarshal([]byte(q), &dsl)
	require.Equal(t, dsl.Aggs["distinct_services"].Terms, &AggTerms{Field: "serviceName", Size: 10000})
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestAvg(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	    "aggs":{
	        "avgPrice":{
	            "avg":{"field":"monies"}
			}
	    },
	    "size":0,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `
	err := json.Unmarshal([]byte(q), &dsl)
	require.Equal(t, dsl.Aggs["avgPrice"].Avg, &AggField{Field: "monies"})
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestMultipleSingle(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	    "aggs":{
	        "avgPrice":{
	            "avg":{"field":"monies"}
			},
	        "maxPrice":{
	            "max":{"field":"monies"}
			}
	    },
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
	require.Equal(t, dsl.Aggs["avgPrice"].Avg, &AggField{Field: "monies"})
	require.Equal(t, dsl.Aggs["maxPrice"].Max, &AggField{Field: "monies"})
}

func TestDateHistogram(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	    "aggs":{
	        "datecounts":{
	            "date_histogram":{
					"field":"datefld",
					"fixed_interval": "3d"
				}
			}
	    },
	    "size":0
	}`
	err := json.Unmarshal([]byte(q), &dsl)
	repr.Println(dsl)
	require.NoError(t, err)
	require.Equal(t, dsl.Aggs["datecounts"].DateHistogram,
		&DateHistogram{Field: "datefld", FixedInterval: "3d"},
	)
}

func TestSubAggregate(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
		"size":0,
		"aggs":{
			"aggOuter":{
				"terms": { "field" : "groupField"},
				"aggregations" : { 
					"maxTime" : {
						"max":{"field":"Time"}
					}
				}
			}
		}
	}`

	err := json.Unmarshal([]byte(q), &dsl)
	repr.Println(dsl)
	require.NoError(t, err)
	require.Equal(t, dsl.Aggs["aggOuter"].Terms.Field, "groupField")
	require.Equal(t, dsl.Aggs["aggOuter"].Aggs["maxTime"].Max.Field, "Time")
}
