package server

import (
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/dsl"
)

func TestBasic(t *testing.T) {

	q := &dsl.Dsl{}
	err := dsl.DslParser.ParseString("", `
	{
	  "query": {
		"term": {"foo": "bar"}
	  },
	  "size": 1
    }`, q)
	require.NoError(t, err)
	sql, err2 := GenSql("testindex", q)
	require.NoError(t, err2)
	repr.Println(sql)
}

func TestBool(t *testing.T) {

	q := &dsl.Dsl{}
	err := dsl.DslParser.ParseString("", `
	{
      "query":{
		"bool":{"must":[{"term":{"foo":"bar"}}]}
	  },
	  "size": 1
    }`, q)
	require.NoError(t, err)
	sql, err2 := GenSql("testindex", q)
	require.NoError(t, err2)
	repr.Println(sql)
}

func TestSort(t *testing.T) {
	q := &dsl.Dsl{}
	err := dsl.DslParser.ParseString("", `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  },
      "sort":[{"asdf":{"order":"desc"}}]
    }`, q)
	require.NoError(t, err)
	sql, err2 := GenSql("testindex", q)
	require.NoError(t, err2)
	repr.Println(sql)
}

func TestRange(t *testing.T) {
	q := &dsl.Dsl{}
	err := dsl.DslParser.ParseString("", `
	{
	  "query": {
		"range":{ 
			"fooTime": {
				"gte":"1654718054570",
				"lte":"1655322854570",
				"format":"epoch_millis"
			}
		}
	  }
    }`, q)
	require.NoError(t, err)
	sql, err2 := GenSql("testindex", q)
	require.NoError(t, err2)
	repr.Println(sql)
}

func TestAggTerms(t *testing.T) {
	q := &dsl.Dsl{}
	err := dsl.DslParser.ParseString("", `
	{
	    "aggs":{
	        "generalStatus":{
	            "terms":{"field":"foo"}
			}
	    },
	    "size":5,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `, q)
	require.NoError(t, err)
	sql, err2 := GenSql("testindex", q)
	require.NoError(t, err2)
	repr.Println(sql)
}
