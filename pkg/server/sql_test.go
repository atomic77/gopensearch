package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/huandu/go-sqlbuilder"
)

var s Server

func loadFixtureData() {
	// This serves essentially as a mini-integration test of data loading since we'll
	// use the bulk loader API to get data into our in-memory version of sqlite
	// If there's a failure, we'll just crash because there's not much point
	// continuing the rest of the suite
	files, err := filepath.Glob("./testdata/*.ndjson")
	if err != nil {
		log.Fatal("failure loading fixtures " + err.Error())
	}

	if len(files) < 1 {
		log.Fatal("Failed to find test fixtures")
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal("failure loading fixture " + err.Error())
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/_bulk", strings.NewReader(string(b)))
		s.Router.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			log.Fatal("bulk load failed with response " + repr.String(rec))
		}
	}
}

func getResponse(t *testing.T, res *http.Response) *SearchResponse {
	// Utility function to check that the response was successful, and send back
	// a generic map of the JSON
	require.Equal(t, res.StatusCode, http.StatusOK)

	// d := make(map[string]interface{})
	d := SearchResponse{}

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = json.Unmarshal(b, &d)
	require.NoError(t, err)

	return &d
}

func TestMain(m *testing.M) {

	cfg := Config{
		DbLocation: "file:memdb?mode=memory&cache=shared",
		ListenAddr: "127.0.0.1",
		Port:       13337,
		Debug:      false,
	}
	s = Server{Cfg: cfg}
	s.Init()
	loadFixtureData()

	os.Exit(m.Run())
}

func TestBasic(t *testing.T) {

	q := `{
	  "query": {
		"term": {"serviceName": "frontend"}
	  },
	  "size": 1
    }`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jaeger-service-2022-11-11/_search", strings.NewReader(q))
	s.Router.ServeHTTP(rec, req)
	d := getResponse(t, rec.Result())
	require.Equal(t, len(d.Hits.Hits), 1)
}

func TestBool(t *testing.T) {

	q := `
	{
      "query":{
		"bool":{"must":[{"term":{"serviceName":"frontend"}}]}
	  },
	  "size": 1
    }`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jaeger-service-2022-11-11/_search", strings.NewReader(q))
	s.Router.ServeHTTP(rec, req)
	d := getResponse(t, rec.Result())
	require.Equal(t, len(d.Hits.Hits), 1)
}

//////////////////
// TODO Finish migrating the rest of these to be real tests of the functionality,
// rather than just ensuring that the query plan looks sane
//////////////////

func TestSort(t *testing.T) {
	q := `
	{
	  "query": {
		"term": { "foo": "bar", "oof": "rab" }
	  },
	  "sort":[{"asdf":{"order":"desc"}}]
    }`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jaeger-service-2022-11-11/_search", strings.NewReader(q))
	s.Router.ServeHTTP(rec, req)
	d := getResponse(t, rec.Result())
	require.Equal(t, len(d.Hits.Hits), 0)
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
	q := `
		{
			"aggs":{
				"dates": {
					"date_histogram":{"field":"startTimeMillis","buckets":200}
				},
				"generalStatus":{
					"terms":{"field":"foo"}
				}
			},
			"size":0
		}
    `
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jaeger-span-2022-11-11/_search", strings.NewReader(q))
	s.Router.ServeHTTP(rec, req)
	d := getResponse(t, rec.Result())

	require.Equal(t, len(d.Hits.Hits), 1)

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
