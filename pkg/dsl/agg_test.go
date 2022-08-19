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
	require.Equal(t, q.Aggs[0].AggregateType.DateHistogram.FixedInterval, "3d")
	repr.Println(q)
}
