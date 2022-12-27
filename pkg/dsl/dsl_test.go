package dsl

import (
	"encoding/json"
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
)

func TestBasicTerm(t *testing.T) {
	dsl := &Dsl{}
	q := `{
	  "query": {
		"term": {"foo": "bar"}
	  },
	  "size": 1
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	// require.Equal(t, dsl.Query.Term.Fields)
	require.Equal(t, dsl.Query.Term["foo"], "bar")
	repr.Println(dsl)
}

func TestBasicMatch(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	  "query": {
		"match": {"foo": "bar"}
	  },
	  "size": 1
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	require.Equal(t, dsl.Query.Match["foo"].Query, "bar")
	repr.Println(dsl)
}

func TestVerboseMatch(t *testing.T) {
	jd1 := Dsl{}
	jd2 := Dsl{}
	jShort := `
	{
	  "query": {
		"match": {
			"foo": "bar"
		}
	  },
	  "size": 1
    }`
	jVerb := `
	{
	  "query": {
		"match": {
			"foo": {
				"query": "bar",
				"operator": "OR"
			}
		}
	  },
	  "size": 1
    }`
	err := json.Unmarshal([]byte(jShort), &jd1)
	require.NoError(t, err)
	// repr.Println(jd)
	err = json.Unmarshal([]byte(jVerb), &jd2)
	require.NoError(t, err)
	// repr.Println(jd)
	require.Equal(t, jd1.Query.Match["foo"].Query, "bar")
	require.Equal(t, jd2.Query.Match["foo"].Query, "bar")
	require.Equal(t, jd2.Query.Match["foo"].Operator, "OR")
}

func TestMultipleTerms(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  }
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
	require.Equal(t, dsl.Query.Term["foo"], "bar")
	require.Equal(t, dsl.Query.Term["oof"], "rab")
}

func TestNestedBoolArray(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
		"query": {
			"bool": {
				"must":[
					{"match": { "foo" : "bar" } }
				]
			}
		},
		"size":1
	} `
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	require.Equal(t, dsl.Query.Bool.Must[0].Match["foo"].Query, "bar")
	repr.Println(dsl)
}

func TestNestedBoolArrayMultiple(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
		"query":{
			"bool":{
				"must":[
					{"match":{"foo":"bar"}},
					{"range":{ "fooTime": { "gte": 1654718054570, "lte": "1655322854570", "format":"epoch_millis" }}}
				]
			}
		},
		"size":1
	}`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
	require.Equal(t, dsl.Query.Bool.Must[0].Match["foo"].Query, "bar")
	require.Equal(t, dsl.Query.Bool.Must[1].Range["fooTime"].Gte.String(), "1654718054570")
}

func TestNestedBoolSingle(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
		"query":{"bool":{"must":{"match":{"oof":"rab"}}}},
		"size":1
	}`

	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestNestedBoolTermSingle(t *testing.T) {
	// Example of a query generated by Jaeger
	dsl := &Dsl{}
	q := `
	{
		"query": { "bool": { "must": { "term": { "traceID": "5aa29bf8d8454e24" } } } },
		"size": 10000,
		"sort": [ { "startTime": { "order": "asc" } } ]
	}
	`
	// TODO Add support for search_after / terminate_after
	// , "search_after": [ 1672152391201000 ],
	// "terminate_after": 10000
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}

func TestSort(t *testing.T) {
	dsl := &Dsl{}
	q := `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  },
      "sort":[{"asdf":{"order":"desc"}}]
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	require.Equal(t, dsl.Sort[0]["asdf"].Order, "desc")
	repr.Println(dsl)
}

func TestRange(t *testing.T) {
	dsl := &Dsl{}
	// Not sure if this is compliant, but we'll change both of these gte/lte to strings internally
	q := `{
	  "query": {
		"range":{ 
			"fooTime": {
				"gte": 1654718054570,
				"lte": "1655322854570",
				"format":"epoch_millis"
			}
		}
	  }
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
	require.Equal(t, dsl.Query.Range["fooTime"].Gte.String(), "1654718054570")
}

func TestRangeWithBooleanParams(t *testing.T) {
	/* Test parsing deprecated boolean include lower/upper parameters */
	dsl := &Dsl{}
	q := `
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
    }`
	err := json.Unmarshal([]byte(q), &dsl)
	require.NoError(t, err)
	repr.Println(dsl)
}
