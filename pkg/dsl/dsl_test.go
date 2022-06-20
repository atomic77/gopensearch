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

func TestNestedBoolMultiple(t *testing.T) {
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
