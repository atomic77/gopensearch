package dsl

import (
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
)

func TestBasic(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
	{
	  "query": {
		"term": {"foo": "bar"}
	  },
	  "size": 1
    }`, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestBasicMatch(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
	{
	  "query": {
		"match": {"foo": "bar"}
	  },
	  "size": 1
    }`, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestMultipleTerms(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  }
    }`, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestNestedBoolArray(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
{
    "query":{"bool":{"must":[{"match":{"foo":"bar"}}]}},
    "size":1,
}
    `, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestNestedBoolArrayMultiple(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
{
    "query":{
		"bool":{
			"must":[
				{"match":{"foo":"bar"}},
				{"range":{ "fooTime": { "gte": 1654718054570, "lte": "1655322854570", "format":"epoch_millis" }}}
			]
		}
	}
    "size":1,
}
    `, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestNestedBoolSingle(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
{
    "query":{"bool":{"must":{"match":{"oof":"rab"}}}},
    "size":1,
}
    `, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestSort(t *testing.T) {
	dsl := &Dsl{}
	err := DslParser.ParseString("", `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  },
      "sort":[{"asdf":{"order":"desc"}}]
    }`, dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestRange(t *testing.T) {
	q := &Dsl{}
	// Not sure if this is compliant, but we'll change both of these gte/lte to strings internally
	err := DslParser.ParseString("", `
	{
	  "query": {
		"range":{ 
			"fooTime": {
				"gte": 1654718054570,
				"lte": "1655322854570",
				"format":"epoch_millis"
			}
		}
	  }
    }`, q)
	require.NoError(t, err)
	repr.Println(q)
}

func TestRangeWithBooleanParams(t *testing.T) {
	/* Test parsing deprecated boolean include lower/upper parameters */
	q := &Dsl{}
	err := DslParser.ParseString("", `
	{
	  "query": {
		"range":{ 
			"fooTime": {
				"gte": 1654718054570,
				"lte": "1655322854570",
                "include_lower": true,
                "include_upper": true,
				"format":"epoch_millis"
			}
		}
	  }
    }`, q)
	require.NoError(t, err)
	repr.Println(q)
}
