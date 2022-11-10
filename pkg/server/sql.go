package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atomic77/gopensearch/pkg/date"
	"github.com/atomic77/gopensearch/pkg/dsl"
)

type dbSubQuery struct {
	sql         string
	aggregation Aggregation
}

func GenPlan(index string, q *dsl.Dsl) ([]dbSubQuery, error) {

	plan := make([]dbSubQuery, 0)
	wh, _ := genQueryWherePredicates(q)

	for _, a := range q.Aggs {
		aggSubQuery := dbSubQuery{}
		aggSubQuery.sql = "SELECT "
		aggInfo := genAggregateSelect(a)
		aggSubQuery.aggregation = aggInfo.aggregation
		aggSubQuery.sql += aggInfo.selectExpr
		aggSubQuery.sql += fmt.Sprintf(` FROM "%s" `, index)
		aggSubQuery.sql += " WHERE "
		aggSubQuery.sql += wh
		if len(aggInfo.groupAliases) > 0 {
			aggSubQuery.sql += " GROUP BY "
			aggSubQuery.sql += strings.Join(aggInfo.groupAliases, " , ")
		}
		plan = append(plan, aggSubQuery)
	}

	// Handle hits selection case
	recordQuery := dbSubQuery{}
	recordQuery.aggregation = nil
	recordQuery.sql = genHitsSelect(index, q)
	recordQuery.sql += wh
	recordQuery.sql += genSort(q.Sort)
	recordQuery.sql += genLimit(q)
	plan = append(plan, recordQuery)
	return plan, nil
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

func genSort(sortFields []*dsl.Sort) string {

	if len(sortFields) == 0 {
		return ""
	}
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
	selectExpr   string
	groupAliases []string
	fnAliases    []string
	aggregation  Aggregation
}

func genHitsSelect(index string, _q *dsl.Dsl) string {

	// TODO Add _source selection capability
	sql := fmt.Sprintf(`SELECT rowid, JSON(content) FROM "%s" WHERE `, index)
	return sql
}

// Returns the field expression for the select, and the alias for the group by
func genAggregateSelect(agg *dsl.Aggregate) aggregateInfo {
	// TODO This will only work with a single aggregation. We'll probably
	// need to generate two queries to sqlite for each

	ai := aggregateInfo{}

	if agg.AggregateType.Terms != nil {
		ai.groupAliases = []string{"g0"}
		ai.fnAliases = []string{"a0"}
		fld := cleanseKeyField(agg.AggregateType.Terms.Field)
		ai.selectExpr = fmt.Sprintf(
			` JSON_EXTRACT(content, '$.%s') as %s, COUNT(*) as %s `,
			fld, ai.groupAliases[0], ai.fnAliases[0],
		)

		ai.aggregation = &BucketAggregation{}
	} else if agg.AggregateType.DateHistogram != nil {
		ai.groupAliases = []string{"g0"}
		ai.fnAliases = []string{"a0"}
		fld := cleanseKeyField(agg.AggregateType.DateHistogram.Field)
		// TODO Can cast dates to an epoch, then divide by the number of seconds the
		// interval corresponds to, eg:
		// SELECT strftime("%s", JSON_EXTRACT(content, '$.Time')) / 1234 as a0  FROM "test-202206" LIMIT 5;
		ai.selectExpr = fmt.Sprintf(
			` JSON_EXTRACT(content, '$.%s') as %s, COUNT(*) as %s `,
			fld, ai.groupAliases[0], ai.fnAliases[0],
		)

		ai.aggregation = &BucketAggregation{}
	} else if agg.AggregateType.Avg != nil {
		ai.fnAliases = []string{"a0"}
		fld := cleanseKeyField(agg.AggregateType.Avg.Field)
		ai.selectExpr = fmt.Sprintf(
			` AVG(JSON_EXTRACT(content, '$.%s')) as %s `,
			fld, ai.fnAliases[0],
		)
		ai.aggregation = &MetricSingleAggregation{}
	} else if agg.AggregateType.Max != nil {
		ai.fnAliases = []string{"a0"}
		fld := cleanseKeyField(agg.AggregateType.Max.Field)
		ai.selectExpr = fmt.Sprintf(
			` MAX(JSON_EXTRACT(content, '$.%s')) as %s `,
			fld, ai.fnAliases[0],
		)
		ai.aggregation = &MetricSingleAggregation{}
	}
	return ai
}

func genLimit(q *dsl.Dsl) string {
	l := 10
	if q.Size != nil {
		l = *q.Size
	}
	return fmt.Sprintf(` LIMIT %d `, l)
}
