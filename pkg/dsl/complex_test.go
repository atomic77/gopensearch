package dsl

import (
	"encoding/json"
	"testing"

	require "github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
)

// Dumping ground for various complex queries we get that

func TestGrafanaExplore(t *testing.T) {
	// Part of an _msearch query used by grafana in initial explore mode
	dsl := &Dsl{}
	q := `
	{
		"size": 0,
		"query": {
			"bool": {
				"filter": [
					{ "range": { "startTimeMillis": { "gte": 1673789792872, "lte": 1673793392872, "format": "epoch_millis" } } },
					{ "query_string": { "analyze_wildcard": true, "query": "*" } }
				]
			}
		},
		"aggs": {
			"2": {
				"date_histogram": {
					"interval": "1s",
					"field": "startTimeMillis",
					"min_doc_count": 0,
					"extended_bounds": {
						"min": 1673789792872,
						"max": 1673793392872
					},
					"format": "epoch_millis"
				},
				"aggs": {}
			}
		}
	}`
	err := json.Unmarshal([]byte(q), &dsl)
	// require.Equal(t, dsl.Aggs["generalStatus"].Terms.Field, "foo")
	require.NoError(t, err)
	repr.Println(dsl)

}
