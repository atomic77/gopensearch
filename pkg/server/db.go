package server

import (
	"encoding/json"
	"fmt"

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
func (s *Server) SearchItem(index string, q *dsl.Dsl) []Document {
	/* Implements a crude version of terms filter, using json_extract
	to do direct match against fields */
	sql, err := GenSql(index, q)
	if err != nil {
		panic(err)
	}
	println(sql)
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []Document
	for rows.Next() {
		doc := Document{}
		var s string
		err2 := rows.Scan(&doc.Id, &s)
		json.Unmarshal([]byte(s), &doc.Content)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, doc)
	}
	return result
}
