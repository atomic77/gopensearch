# Gopensearch

An experimental single-process reimplementation of \[elastic|open\]search. Built with golang and sqlite3 as the backend data store, using the `fts5` and `json1` extensions.  

Aims to provide a lightweight ES-compatible implementation for constrained environments like SBCs / raspberry pis, dev-test scenarios and test pipelines, or anyone that just wants ES to start up quickly and use less memory. 

The full API is massive and most of it is unlikely to be implemented, though I imagine that there are plenty of people out there using only a small subset of its capabilities.

## Tool Support

With an enormous API full of functionality, I'm using popular open-source tools that depend on Elasticsearch as a backend to drive what features should be added. Here is a list of those that I have tried so far: 

| Tool | Description | Support | 
|------|-------------|---------|
| Jaeger      |  Open-source tracing framework    | Wildcard and multi-index searching used by front-end still not fully supported ; a hack in place ignores previous days when calling something like `POST /idx1,idx2,idx3/_search`   |
| Grafana | Open-source visualization and dashboarding tool |  ES Datasource can be added, and the Explore mode sort of works. More work needed on date histogram compatibility. |

## Demo

A demo docker-compose file can be used to bring up [Jaeger all-in-one](https://github.com/jaegertracing/jaeger/), the "HotRod"
trace generator, and a test instance of Gopensearch running in place of Elasticsearch. You can start it with:

```bash
$ docker-compose -f docker-compose-demo.yml up
Starting gopensearch_gopensearch_1 ... done
Starting gopensearch_jaeger-all-in-one_1 ... done
Starting gopensearch_hotrod_1            ... done
Attaching to gopensearch_gopensearch_1, gopensearch_jaeger-all-in-one_1, gopensearch_hotrod_1
gopensearch_1        | 2022/12/29 22:38:15 server.Config{DbLocation: "/tmp/test.db", ListenAddr: "0.0.0.0", Port: 9200, }
gopensearch_1        | 2022/12/29 22:38:15 Starting server on 0.0.0.0:9200
...

```
A Grafana instance can also be brought up with the demo playbook that can be connected to an Elasticsearch Datasource.

The indices created automatically can be viewed with:

```bash
$ curl http://localhost:9200/_cat/indices
green	open	jaeger-service-2022-12-29
green	open	jaeger-span-2022-12-29
```

The HotRod GUI will be available on http://localhost:8080, and the Jaeger UI available at http://localhost:16686:

![jaeger-GUI](./doc/jaeger-ui.png)

## Roadmap

Basic support in place for:
* Index and document creation
* Bulk doc creation
* Term/match queries
* Templates
  * Support for mapping date fields using ES format types like `epoch_millis` 
* Bool must/should/filter compound queries 
* Multiple single-value aggregates
* Simple subaggregations
  * Limited to those that can be easily mapped to a single SQL statement (eg. single metric aggregate coupled with terms)

Near-term goals:
* Wild-card and explicit multiple index searching
* Improved date formatting
* Date histograms
* Documentation for what is supported and what isn't
* Improved integration tests 
* Artifact releases (docker image and binary build) direct to github

Future work:
* Indexing and storage optimizations in sqlite usage

Out of scope:
* Most things :) Clustering, sharding, painless lang, etc.


## Building

For Linux amd64:

```bash
make
```

For ARM chipsets (32 and 64-bit):

```bash
make arm
```

### Why make this?

I like running things on SBCs, often just to see if it can be done. And as I discovered, elasticsearch is not too happy running on a 1GB armv7 Orange Pi. I have also enjoyed the challenge of building something complex in golang from scratch, and I'm learning as I go (pardon the pun). 