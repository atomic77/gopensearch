package server

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/repr"
)

func TestCustomJsonSer(t *testing.T) {
	sr := SearchResponse{
		Took: 1,
	}
	sr.Aggregations = make(map[string]Aggregation)
	b := makeBucket()
	b.Key = "1234"
	b.KeyAsString = "5678"
	b.DocCount = 42
	br := BucketAggregation{}
	br.Buckets = make([]Bucket, 0)
	br.Buckets = append(br.Buckets, b)
	sr.Aggregations["agg1"] = &br

	j, _ := json.Marshal(sr)

	repr.Println(j)
}
