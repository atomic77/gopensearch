package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atomic77/gopensearch/pkg/date"
	"github.com/atomic77/gopensearch/pkg/dsl"
)

func GenSql(index string, q *dsl.Dsl) (string, error) {
	var sql string

	wh, _ := genQueryWherePredicates(q)

	if q.Aggs != nil {
		ai, _ := handleAggs(q.Aggs)

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
	if q.Query == nil {
		return " 1 = 1 ", nil
	}
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
			} else if v.Range != nil {
				sqlFragments = append(sqlFragments, handleRange(v.Range))
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
		key := cleanseKeyField(prop.Key)
		iVal, err := strconv.ParseInt(prop.Value, 10, 64)
		var s string
		if err == nil {
			// Interpret this as an integer
			s = fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') = %d `, key, iVal)
		} else {
			s = fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') = '%s' `, key, prop.Value)
		}
		preds = append(preds, s)
		// t := fmt.Sprintf(" %s MATCH '%s'", index, v)
	}

	return strings.Join(preds, " AND ")
}

func cleanseKeyField(f string) string {
	// Strip away .keyword since we don't distinguish it
	key := strings.Split(f, ".keyword")[0]
	return key
}

func handleRange(rng *dsl.Range) string {
	// Currently only working for date ranges
	var (
		preds []string
	)
	fmtFn := date.DateFormatFn(*rng.RangeOptions.Format)

	if rng.RangeOptions.Lte != nil {
		s := fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) <= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lte))
		preds = append(preds, s)
	} else if rng.RangeOptions.Lt != nil {
		s := fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) < '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lt))
		preds = append(preds, s)
	}

	if rng.RangeOptions.Gte != nil {
		s := fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) >= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gte))
		preds = append(preds, s)
	} else if rng.RangeOptions.Gt != nil {
		s := fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) > '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gt))
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
			fld := cleanseKeyField(a.Aggregation.Terms.Field)
			ai.selectExpr = fmt.Sprintf(
				` JSON_EXTRACT(content, '$.%s') as %s, COUNT(*) as %s`,
				fld, ai.aggrAlias, ai.groupAlias,
			)
			return ai, nil
		}
	}
	return nil, nil
}
