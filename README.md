## Gopensearch

A lightweight single-process reimplementation of \[elastic|open\]search. Built with golang and sqlite3 as the backend data store, using the `fts5` and `json1` extensions.  

Aims to provide an elasticsearch-compatible implementation in constrained environments like SBCs / raspberry pis, dev-test scenarios and test pipelines. 

The full API is massive and most of it is unlikely to be implemented, though I imagine that there are plenty of people out there using only a small subset of its capabilities.

## Implementation Roadmap

Working:
* Index and doc creation
* Bulk doc creation
* Basic support for term/match queries against string fields
* Bool must/should compound queries crudely supported
* Multiple single-value aggregates

Near-term goals:
* Multi-valued and nested-aggregations
* Improved test case coverage
* Improved documentation for what is supported and what isn't, and helpful error messages when unsupported features are used

Future work:
* Indexing and storage optimizations in sqlite usage

Out of scope:
* Most things :) Clustering, sharding, painless lang, etc.

### Support Matrix

Until I get around to properly documenting what works and what doesn't, 
you can get an idea for sorts of queries are supported by looking at 
the test cases in the `pkg/dsl` folder.

### Why make this?

I like running things on SBCs, often just to see if it can be done. And as I discovered, elasticsearch is not too happy running on a 1GB armv7 Orange Pi.

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
