package server

import (
	"encoding/json"
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/huandu/go-sqlbuilder"
)

func TestBasic(t *testing.T) {

	d := &dsl.Dsl{}
	q := `{
	  "query": {
		"term": {"foo": "bar"}
	  },
	  "size": 1
    }`
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	if len(plan) != 1 {
		t.Error("Expected only one query in plan")
	}
	require.NoError(t, err2)
	repr.Println(plan)
}

func TestBool(t *testing.T) {

	d := &dsl.Dsl{}
	q := `
	{
      "query":{
		"bool":{"must":[{"term":{"foo":"bar"}}]}
	  },
	  "size": 1
    }`
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	require.NoError(t, err2)
	if len(plan) != 1 {
		t.Error("Expected only one query in plan")
	}
	repr.Println(plan)
	repr.Println(plan[0].sb.String())
}

func TestSort(t *testing.T) {
	d := &dsl.Dsl{}
	q := `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  },
	  "sort":[{"asdf":{"order":"desc"}}]
    }`
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	if len(plan) != 1 {
		t.Error("Expected only one query in plan")
	}
	require.NoError(t, err2)
	repr.Println(plan)
	repr.Println(plan[0].sb.String())
}

func TestRange(t *testing.T) {
	d := &dsl.Dsl{}
	q := `
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
    }`
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	if len(plan) != 1 {
		t.Error("Expected only one query in plan")
	}
	require.NoError(t, err2)
	repr.Println(plan)
	repr.Println(plan[0].sb.String())
}

func TestAggTerms(t *testing.T) {
	d := &dsl.Dsl{}
	q := `
	{
	    "aggs":{
	        "generalStatus":{
	            "terms":{"field":"foo"}
			}
	    },
	    "size":5,
		"query": { "term": { "foo": "bar", "oof": "rab" } }
	}
    `
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	if len(plan) != 2 {
		t.Error("Expected two queries in plan")
	}

	require.NoError(t, err2)
	repr.Println(plan)
	repr.Println(plan[0].sb.String())
	repr.Println(plan[1].sb.String())
}

func TestDateHistogram(t *testing.T) {
	d := &dsl.Dsl{}
	q := `
	{
	    "aggs":{
			"dates": {
				"date_histogram":{"field":"Time","buckets":200},
			}
	        "generalStatus":{
	            "terms":{"field":"foo"}
			}
	    },
	    "size":0
	}
    `
	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)
	if len(plan) != 3 {
		t.Error("Expected 3 queries in plan")
	}
	if plan[0].aggregation == nil && plan[0].aggregation.GetAggregateCategory() == dsl.MetricsSingle {
		t.Error("Expected an aggregation query first")
	}
	require.NoError(t, err2)
	repr.Println(plan)
}

func TestSubAggregate(t *testing.T) {
	d := &dsl.Dsl{}
	q := `
	{
		"aggregations": {
			"traceIDs": {
				"aggregations": {
					"startTime": {
						"max": {
							"field": "startTime"
						}
					}
				},
				"terms": {
					"field": "traceID",
					"size": 20
				}
			}
		},
		"size": 0
	}
    `

	err := json.Unmarshal([]byte(q), &d)
	require.NoError(t, err)
	plan, err2 := GenPlan("testindex", d)

	// if !strings.Contains(plan[1].sb.String(), "f1") {
	// 	t.Error("Did not find a second function statement")
	// }
	require.NoError(t, err2)
	repr.Println(plan[0].sb.String())
	repr.Println(plan[1].sb.String())
}

func TestSelectBuild(t *testing.T) {

	sb := sqlbuilder.NewSelectBuilder()

	// Looks like successive calls to Select destroy the previous
	// expressions, unlike with sb.Where() :-/
	sb.Select(
		sb.As("col1", "a1"),
		sb.As("col2", "a2"),
	)
	sb.Select(sb.As("col3", "a3"))
	repr.Println(sb.String())
}

func TestWhereBuild(t *testing.T) {

	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		sb.As("col1", "a1"),
		sb.As("col2", "a2"),
	)
	sb.From("mytable")
	// 	Where("1 =1 ")

	// sb.Where("2 = 2")
	sb.Select("col3")
	repr.Println(sb.String())
}

func TestStructBuild(t *testing.T) {

	b := Bucket{
		KeyAsString: "qwerqwer",
		DocCount:    123,
		Key:         "key1",
	}
	stc := sqlbuilder.NewStruct(b)
	sb := stc.SelectFrom("footab")
	repr.Println(stc)
	repr.Println(sb.String())
	repr.Println(stc.Addr(&b))
}

func TestSubSelectBuild(t *testing.T) {
	// Looks like we can't avoid the trailing "... FROM" being added
	//

	subSelect := sqlbuilder.NewSelectBuilder()
	sb := sqlbuilder.NewSelectBuilder()
	subSelect.Select("max(c3)")
	sb.Select("c1", "c2", sb.BuilderAs(subSelect, "f3"))
	repr.Println(sb.String())
}
