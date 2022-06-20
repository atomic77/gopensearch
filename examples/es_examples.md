## Actual ES db calls:

Creating a simple doc:

```bash
$ curl -s -X "POST" http://localhost:9200/newindex/_create/1 --json '{"hello": "world"}'
{"_index":"newindex","_type":"_doc","_id":"1","_version":1,"result":"created","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":0,"_primary_term":1}

```

Searching for the same record:

```bash
$ curl -s -X "POST" http://localhost:9200/newindex/_search --json '{"query": {"term": {"hello": "world"}}}' | jq
{
  "took": 0,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 1,
      "relation": "eq"
    },
    "max_score": 0.2876821,
    "hits": [
      {
        "_index": "newindex",
        "_type": "_doc",
        "_id": "1",
        "_score": 0.2876821,
        "_source": {
          "hello": "world"
        }
      }
    ]
  }
}
# Invalid search:
$ curl -s -X "POST" http://localhost:9200/newindex/_search --json '{"query": {"term": {"asdfhello": "world"}}}' | jq
{
  "took": 0,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 0,
      "relation": "eq"
    },
    "max_score": null,
    "hits": []
  }
}


```