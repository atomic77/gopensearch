package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/jmoiron/sqlx"
)

func (s *Server) IndexDocument(doc string, index string) error {
	// Crude implementation of ES index API; write our
	// document into the doc TEXT blob with the given id
	// Insert into fts5 index; rowid will be created automatically
	sql := fmt.Sprintf(` INSERT INTO '%s' (content) VALUES (json(?)) `, index)

	// How to index specific json columns in sqlite:
	// https://dgl.cx/2020/06/sqlite-json-support
	// https://www.sqlite.org/gencol.html  <- as of 3.31
	// Need to figure out best way to provide mapping between indexing features of ES and how we'd add
	// columns on the fly to ensure fast performance without bloating the size of the resulting file

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(doc)
	return err2
}

func (s *Server) CreateTable(index string) {
	// Mimic the creation of an elasticsearch index with an FTS5 virtual table
	sql := fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS "%s" USING fts5(content);`,
		index,
	)
	_, err := s.db.Exec(sql, index)
	if err != nil {
		panic(err)
	}
}

/*
Example sqlite query to do FT-searching + json-based filtering:

select json(ftidx.content) from ftidx
where json_extract(ftidx.content, '$.a') = 123
and ftidx MATCH 'earth';

*/
func (s *Server) SearchItem(index string, q *dsl.Dsl) ([]Document, []Aggregation, error) {
	var (
		aggs []Aggregation
		docs []Document
	)

	subQueries, err := GenPlan(index, q)
	if err != nil {
		panic(err)
	}

	for _, q := range subQueries {
		log.Println(q.sb.String())
		rows, err := s.db.Queryx(q.sb.String())
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()
		if q.aggregation != nil {
			q.aggregation.SerializeResultset(rows, &q)
			aggs = append(aggs, q.aggregation)

		} else {
			docs = s.execHitsSubquery(q)
		}
	}
	return docs, aggs, nil
}

func (m *BucketAggregation) getSubAggregates() map[string]interface{} {
	// TODO Figure out a way of returning all of the sub-aggregates from the
	// underlying sql statement
	return nil
}

func (m *BucketAggregation) SerializeResultset(rows *sqlx.Rows, dbq *dbSubQuery) {
	for rows.Next() {
		b := makeBucket()
		dest := make(map[string]interface{})
		err := rows.MapScan(dest)
		if err != nil {
			panic(err)
		}
		// FIXME Assume the first group element is the key we want
		for k := range dbq.groupAliases {
			if v, ok := dest[k].(string); ok {
				b.Key = v
			} else {
				b.Key = ""
			}
			break
		}

		// TODO Another hack - just assume the first AggTerms instance we see is
		// the entry we want to put in the main Buckets{} struct
		// All of these type conversions need to be converted into appropriate
		// interface implementations to clean this up
		for k, v := range dbq.fnAliases {
			if _, ok := v.(*dsl.AggTerms); ok {
				b.DocCount = dest[k].(int64)
			}
			if af, ok := v.(*dsl.AggField); ok {
				// Extract this struct literal out
				var destVal string
				if val, ok := dest[k].(string); ok {
					destVal = val
				} else if val, ok := dest[k].(int64); ok {
					destVal = fmt.Sprintf("%d", val)
				}
				b.subaggregates[af.Field] = struct {
					Value         string `json:"value,omitempty"`
					ValueAsString string `json:"value_as_string,omitempty"`
				}{
					Value:         destVal,
					ValueAsString: "not_yet_implemented",
				}
			}
		}
		m.Buckets = append(m.Buckets, b)
	}
}

func (m *MetricMultipleAggregation) SerializeResultset(rows *sqlx.Rows, dbq *dbSubQuery) {
	// TODO IMPLEMENT ME as with Bucket aggregation
}

func (m *MetricSingleAggregation) SerializeResultset(rows *sqlx.Rows, dbq *dbSubQuery) {
	for rows.Next() {
		err := rows.Scan(&m.Value)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Server) execHitsSubquery(q dbSubQuery) []Document {

	var docs []Document
	log.Println(q.sb.String())
	rows, err := s.db.Query(q.sb.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		doc := Document{}
		var s string
		err2 := rows.Scan(&doc.Id, &s)
		json.Unmarshal([]byte(s), &doc.Content)
		if err2 != nil {
			panic(err2)
		}
		docs = append(docs, doc)
	}
	return docs
}
