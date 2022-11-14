package dsl

import (
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
)

func TestAggTerms(t *testing.T) {
	q := &Dsl{}
	err := DslParser.ParseString("", `
	{
	    "aggs":{
	        "generalStatus":{
	            "terms":{"field":"foo"}
			}
	    },
	    "size":0,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `, q)
	require.NoError(t, err)
	repr.Println(q)
}

func TestAggTermsWithLongName(t *testing.T) {
	q := &Dsl{}
	err := DslParser.ParseString("", `
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
    `, q)
	require.NoError(t, err)
	repr.Println(q)
}

func TestAvg(t *testing.T) {
	q := &Dsl{}
	err := DslParser.ParseString("", `
	{
	    "aggs":{
	        "avgPrice":{
	            "avg":{"field":"monies"}
			}
	    },
	    "size":0,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `, q)
	require.NoError(t, err)
	repr.Println(q)
}

func TestMultipleSingle(t *testing.T) {
	q := &Dsl{}
	err := DslParser.ParseString("", `
	{
	    "aggs":{
	        "avgPrice":{
	            "avg":{"field":"monies"}
			},
	        "maxPrice":{
	            "max":{"field":"monies"}
			}
	    },
	    "size":0,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `, q)
	require.NoError(t, err)
	repr.Println(q)
}

func TestDateHistogram(t *testing.T) {
	q := &Dsl{}
	DslParser.ParseString("", `
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
	}
    `, q)
	require.Equal(t, q.Aggs[0].AggregateType[0].DateHistogram.FixedInterval, "3d")
	repr.Println(q)
}

func TestSubAggregate(t *testing.T) {
	q := &Dsl{}
	err := DslParser.ParseString("", `
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
	}`, q)

	require.NoError(t, err)
	repr.Println(q)
}
