## Gopensearch

A basic and lightweight single-process replacement for open/elasticsearch. Built with go and sqlite as the backend data store, using the `fts5` and `json1` extensions.  

Aims to provide an elasticsearch-compatible implementation in constrained environments like SBCs / raspberry pis, and test pipelines. 

The elasticsearch API is massive and most of it is unlikely to be implemented, though I imagine that there are plenty of people out there using only a small subset of its capabilities.

## Implementation Roadmap

Working:
* Basic index and documentation creation 
* Bulk indexing
* Basic support for term/match queries against string fields
* Bool must/should compound queries crudely supported

Near-term goals:
* Improving parsing coverage of Query DSL
* Basic aggregation 
* Improved test case coverage

Future work:
* Indexing and storage optimizations in sqlite usage (lots of fts5 and json1 capabilities are unused)

Out of scope:
* Most things :) Clustering, sharding, etc.

### Why make this?

I like running things on SBCs, often just to see if it can be done. And as I discovered, elasticsearch is not too happy running on a 1GB Orange Pi.

### Example use

```bash
$ curl -s -X "PUT" http://localhost:8080/newindex | jq
{
  "acknowledged": true,
  "shards_acknowledged": true,
  "index": "newindex"
}
```


```bash
$ curl -s -X "POST" http://localhost:8080/newindex/_create -d '{"hello": "world"}' | jq
{
  "_index": "newindex",
  "_id": 2,
  "_version": 1,
  "result": "created"
}

```

```bash
$ curl -s -X "POST" http://localhost:8080/newindex/_search -d '{"query": { "term": {"hello": "world"} }' | jq
[
  {
    "id": 1,
    "_source": {
      "hello": "world"
    }
  }
]
```

`examples` has some more example use cases.
