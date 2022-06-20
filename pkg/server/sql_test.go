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
