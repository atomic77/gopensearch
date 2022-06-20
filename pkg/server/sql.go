package server

import (
	"fmt"
	"strings"

	"github.com/atomic77/gopensearch/pkg/dsl"
)

func GenSql(index string, q *dsl.Dsl) (string, error) {
	var sql string

	wh, _ := genQueryWherePredicates(q)

	if q.Aggs != nil {
		ai, _ := handleAggs(q.Aggs)

		// sql := " SELECT "
		sql = fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE %s  GROUP BY %s`,
			ai.selectExpr, index, wh, ai.aggrAlias,
		)

	} else {
		sql = fmt.Sprintf(`SELECT rowid, JSON(content) FROM "%s" WHERE `, index)
		sql += wh
	}

	if q.Sort != nil {
		sql += handleSort(q.Sort)
	}

	if q.Size != nil {
		sql += fmt.Sprintf(" LIMIT %d ", *q.Size)
	}
	return sql, nil
}

// Generate sql statement for a given Query DSL
func genQueryWherePredicates(q *dsl.Dsl) (string, error) {
	var sql string
	if q.Query.Bool != nil {
		sql += handleBool(q.Query.Bool)
	} else if q.Query.Term != nil {
		sql += handleTermOrMatch(q.Query.Term.Properties)
	} else if q.Query.Match != nil {
		sql += handleTermOrMatch(q.Query.Match.Properties)
	} else if q.Query.Range != nil {
		sql += handleRange(q.Query.Range)
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
		s := fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') = '%s' `, prop.Key, prop.Value)
		preds = append(preds, s)
		// t := fmt.Sprintf(" %s MATCH '%s'", index, v)
	}

	return strings.Join(preds, " AND ")
}

func handleRange(rng *dsl.Range) string {
	// TODO Generates something, but the format and casting of dates from the string input
	// coming in needs to be properly handled
	var preds []string
	if rng.RangeOptions.Lte != nil {
		s := fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') <= '%s' `, rng.Field, *rng.RangeOptions.Lte)
		preds = append(preds, s)
	} else if rng.RangeOptions.Lt != nil {
		s := fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') < '%s' `, rng.Field, *rng.RangeOptions.Lt)
		preds = append(preds, s)
	}

	if rng.RangeOptions.Gte != nil {
		s := fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') >= '%s' `, rng.Field, *rng.RangeOptions.Gte)
		preds = append(preds, s)
	} else if rng.RangeOptions.Gt != nil {
		s := fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') > '%s' `, rng.Field, *rng.RangeOptions.Gt)
		preds = append(preds, s)
	}

	return strings.Join(preds, " AND ")
}

func handleSort(sortFields []*dsl.Sort) string {

	sql := " ORDER BY "
	var frags []string

	for _, v := range sortFields {
		s := fmt.Sprintf(
			` JSON_EXTRACT(content, '$.%s') %s `,
			v.Field,
			strings.ToUpper(v.SortOrder.Order),
		)
		frags = append(frags, s)
	}
	sql += strings.Join(frags, " , ")
	return sql
}

type aggregateInfo struct {
	selectExpr string
	groupAlias string
	aggrAlias  string
}

// Returns the field expression for the select, and the alias for the group by
func handleAggs(aggs []*dsl.Aggregate) (*aggregateInfo, error) {
	// TODO This will only work with a single aggregation. We'll probably
	// need to generate two queries to sqlite for each

	for i, a := range aggs {
		if a.Aggregation.Terms != nil {
			ai := &aggregateInfo{}
			ai.aggrAlias = fmt.Sprintf("%s%d", "a", i)
			ai.groupAlias = fmt.Sprintf("%s%d", "g", i)
			ai.selectExpr = fmt.Sprintf(
				` JSON_EXTRACT(content, '$.%s') as %s, COUNT(*) as %s`,
				a.Aggregation.Terms.Field,
				ai.aggrAlias,
				ai.groupAlias,
			)
			return ai, nil
		}
	}
	return nil, nil
}
