package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/atomic77/gopensearch/pkg/dsl"
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

		log.Println(q.sql)
		rows, err := s.db.Query(q.sql)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()
		if q.aggregation != nil {
			q.aggregation.SerializeResultset(rows)
			aggs = append(aggs, q.aggregation)

		} else {
			docs = s.execHitsSubquery(q)
		}
	}
	return docs, aggs, nil
}

func (m *BucketAggregation) SerializeResultset(rows *sql.Rows) {
	for rows.Next() {
		b := Bucket{}
		err := rows.Scan(&b.Key, &b.DocCount)
		if err != nil {
			panic(err)
		}
		m.Buckets = append(m.Buckets, b)
	}
}

func (m *MetricMultipleAggregation) SerializeResultset(rows *sql.Rows) {
	// TODO IMPLEMENT ME as with Bucket aggregation
}

func (m *MetricSingleAggregation) SerializeResultset(rows *sql.Rows) {
	for rows.Next() {
		err := rows.Scan(&m.Value)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Server) execHitsSubquery(q dbSubQuery) []Document {

	var docs []Document
	log.Println(q.sql)
	rows, err := s.db.Query(q.sql)
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
