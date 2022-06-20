package server

import (
	"fmt"
	"strings"

	"github.com/atomic77/gopensearch/pkg/dsl"
)

// Generate sql statement for a given Query DSL
func GenSql(index string, q *dsl.Dsl) (string, error) {
	sql := fmt.Sprintf(
		`SELECT rowid, json(content) FROM "%s" WHERE `, index)

	if q.Query.Bool != nil {
		sql += handleBool(q.Query.Bool)
	} else if q.Query.Term != nil {
		sql += handleTermOrMatch(q.Query.Term.Properties)
	} else if q.Query.Match != nil {
		sql += handleTermOrMatch(q.Query.Match.Properties)
	}

	if q.Size != nil {
		sql += fmt.Sprintf(" LIMIT %d ", *q.Size)
	}
	return sql, nil
}

func handleBool(b *dsl.Bool) string {
	sqlWhere := " ( "

	var sqlFragments []string
	if b.Must != nil {
		for _, v := range b.Must.Queries {
			if v.Match != nil {
				sqlFragments = append(sqlFragments, handleTermOrMatch(v.Match.Properties))
			} else if v.Term != nil {
				sqlFragments = append(sqlFragments, handleTermOrMatch(v.Term.Properties))
			}
		}
		sqlWhere += strings.Join(sqlFragments, " AND ")
	}
	//  else if b.Should != nil {

	// }

	sqlWhere += " ) "
	return sqlWhere
}

// Treat Term and Match as interchangeable for now
func handleTermOrMatch(props []*dsl.Property) string {
	var preds []string
	for _, prop := range props {
		// TODO This will only work with string types
		s := fmt.Sprintf(` json_extract(content, '$.%s') = '%s' `, prop.Key, prop.Value)
		preds = append(preds, s)
		// t := fmt.Sprintf(" %s MATCH '%s'", index, v)
	}

	return strings.Join(preds, " AND ")
}
