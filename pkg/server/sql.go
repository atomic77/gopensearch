package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atomic77/gopensearch/pkg/date"
	"github.com/atomic77/gopensearch/pkg/dsl"
	"github.com/huandu/go-sqlbuilder"
)

type dbSubQuery struct {
	aggregation  Aggregation
	sb           *sqlbuilder.SelectBuilder
	selectExprs  []string
	groupAliases map[string]interface{}
	fnAliases    map[string]interface{}
}

func makeDbSubQuery() dbSubQuery {
	dbq := dbSubQuery{}
	dbq.sb = sqlbuilder.SQLite.NewSelectBuilder()
	dbq.selectExprs = make([]string, 0)
	dbq.groupAliases = make(map[string]interface{})
	dbq.fnAliases = make(map[string]interface{})
	return dbq
}

func (dbq *dbSubQuery) isAggregation() bool {
	return dbq.aggregation == nil
}

func GenPlan(index string, q *dsl.Dsl) ([]dbSubQuery, error) {

	plan := make([]dbSubQuery, 0)

	for _, a := range q.Aggs {
		aggQ := makeDbSubQuery()
		aggQ.genAggregateSelectExprs(a)

		aggQ.genSelectExpression()
		aggQ.sb.From(fmt.Sprintf(`"%s"`, index))
		aggQ.genQueryWherePredicates(q)
		aggQ.genAggGroupBy()
		plan = append(plan, aggQ)
	}

	// Handle hits selection case
	hitsQ := makeDbSubQuery()
	hitsQ.genHitsSelect(index, q)
	hitsQ.genQueryWherePredicates(q)
	hitsQ.genSort(q.Sort)
	hitsQ.genLimit(q)
	hitsQ.aggregation = nil
	plan = append(plan, hitsQ)
	return plan, nil
}

// Generate sql statement for a given Query DSL
func (dbq *dbSubQuery) genQueryWherePredicates(q *dsl.Dsl) error {
	// var sql string
	if q.Query == nil {
		dbq.sb.Where("1 = 1")
		return nil
	}
	if q.Query.Bool != nil {
		dbq.handleBool(q.Query.Bool)
	} else if q.Query.Term != nil {
		dbq.handleTermOrMatch(q.Query.Term.Properties)
	} else if q.Query.Match != nil {
		dbq.handleTermOrMatch(q.Query.Match.Properties)
	} else if q.Query.Range != nil {
		dbq.handleRange(q.Query.Range)
	}

	return nil
}

func (dbq *dbSubQuery) handleBool(b *dsl.Bool) error {
	if b.Must != nil {
		for _, v := range b.Must {
			if v.Query.Match != nil {
				dbq.handleTermOrMatch(v.Query.Match.Properties)
			} else if v.Query.Term != nil {
				dbq.handleTermOrMatch(v.Query.Term.Properties)
			} else if v.Query.Range != nil {
				dbq.handleRange(v.Query.Range)
			}
		}
	}
	//  else if b.Should != nil {
	return nil
}

// Treat Term and Match as interchangeable for now
func (dbq *dbSubQuery) handleTermOrMatch(props []*dsl.Property) error {
	for _, prop := range props {
		key := cleanseKeyField(prop.Key)
		iVal, err := strconv.ParseInt(prop.Value, 10, 64)
		if err == nil {
			// Interpret this as an integer
			dbq.sb.Where(fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') = %d `, key, iVal))
		} else {
			dbq.sb.Where(fmt.Sprintf(` JSON_EXTRACT(content, '$.%s') = '%s' `, key, prop.Value))
		}
	}

	return nil
}

func cleanseKeyField(f string) string {
	// Strip away .keyword since we don't distinguish it
	key := strings.Split(f, ".keyword")[0]
	return key
}

func (dbq *dbSubQuery) handleRange(rng *dsl.Range) error {
	// Currently only working for date ranges
	fmtStr := "epoch_millis"
	if rng.RangeOptions.Format != nil {
		fmtStr = *rng.RangeOptions.Format
	}
	fmtFn := date.DateFormatFn(fmtStr)

	if rng.RangeOptions.Lte != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s'), 'auto') <= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lte)))
	} else if rng.RangeOptions.Lt != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s'), 'auto') < '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lt)))
	}

	if rng.RangeOptions.Gte != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s'), 'auto') >= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gte)))
	} else if rng.RangeOptions.Gt != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s'), 'auto') > '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gt)))
	}
	return nil
}

func (dbq *dbSubQuery) genSort(sortFields []*dsl.Sort) {

	if len(sortFields) == 0 {
		return
	}

	for _, v := range sortFields {
		dbq.sb.OrderBy(fmt.Sprintf(
			` JSON_EXTRACT(content, '$.%s') %s `,
			v.Field,
			strings.ToUpper(v.SortOrder.Order),
		))
	}
}

func (dbq *dbSubQuery) genSelectExpression() {
	//
	dbq.sb.Select(dbq.selectExprs...)
}

func (dbq *dbSubQuery) genHitsSelect(index string, _q *dsl.Dsl) {
	dbq.sb.
		Select("rowid", "JSON(content)").
		// Looks like the Sqlite dialect doesn't properly escape tables with odd characters
		From(fmt.Sprintf(`"%s"`, index))
}

// Returns the field expression for the select, and the alias for the group by
func (dbq *dbSubQuery) genAggregateSelectExprs(root *dsl.Aggregate) {

	for _, agg := range root.AggregateType {
		grpIdx := dbq.getNextGrpAlias()
		fnIdx := dbq.getNextFnAlias()

		if agg.Terms != nil {
			dbq.groupAliases[grpIdx] = agg.Terms
			dbq.fnAliases[fnIdx] = agg.Terms
			fld := cleanseKeyField(agg.Terms.Field)
			dbq.selectExprs = append(dbq.selectExprs,
				dbq.sb.As(fmt.Sprintf(` JSON_EXTRACT(content, '$.%s')`, fld), grpIdx),
				dbq.sb.As("COUNT(*)", fnIdx),
			)

			dbq.aggregation = &BucketAggregation{}
		} else if agg.DateHistogram != nil {
			dbq.groupAliases[grpIdx] = agg.DateHistogram
			dbq.fnAliases[fnIdx] = agg.DateHistogram
			fld := cleanseKeyField(agg.DateHistogram.Field)
			// TODO Can cast dates to an epoch, then divide by the number of seconds the
			// interval corresponds to, eg:
			// SELECT strftime("%s", JSON_EXTRACT(content, '$.Time')) / 1234 as a0  FROM "test-202206" LIMIT 5;
			dbq.selectExprs = append(dbq.selectExprs,
				dbq.sb.As(fmt.Sprintf(` JSON_EXTRACT(content, '$.%s')`, fld), grpIdx),
				dbq.sb.As("COUNT(*)", fnIdx),
			)

			dbq.aggregation = &BucketAggregation{}
		} else if agg.Avg != nil {
			dbq.fnAliases[fnIdx] = agg.Avg
			fld := cleanseKeyField(agg.Avg.Field)
			dbq.selectExprs = append(dbq.selectExprs,
				dbq.sb.As(fmt.Sprintf(` AVG(JSON_EXTRACT(content, '$.%s')`, fld), fnIdx),
			)
			dbq.aggregation = &MetricSingleAggregation{}
		} else if agg.Max != nil {
			dbq.fnAliases[fnIdx] = agg.Max
			fld := cleanseKeyField(agg.Max.Field)
			dbq.selectExprs = append(dbq.selectExprs,
				dbq.sb.As(fmt.Sprintf(` MAX(JSON_EXTRACT(content, '$.%s'))`, fld), fnIdx),
			)
			dbq.aggregation = &MetricSingleAggregation{}
		} else if agg.Aggs != nil {
			// Experiment with embedding (SELECT) clauses right into the SQL. Sqlite
			// seems to handle this well
			for _, subAgg := range agg.Aggs {
				subQry := makeDbSubQuery()
				subQry.genAggregateSelectExprs(subAgg)
				subQry.genSelectExpression()
				// sqlbuilder always adds to any select statement "FROM" even though
				// we haven't added any tables, so we have to wrap this in parenthesis manually
				subSql := subQry.sb.String()
				subSql = strings.TrimSuffix(subSql, "FROM ")
				dbq.appendSubQuery(&subQry)
				dbq.selectExprs = append(dbq.selectExprs,
					dbq.sb.As(fmt.Sprintf(" (%s) ", subSql), fnIdx),
				)
				fnIdx = dbq.getNextFnAlias()
			}
		}
	}
}

func (dbq *dbSubQuery) getNextGrpAlias() string {
	return fmt.Sprintf("g%d", len(dbq.groupAliases)+1)
}

func (dbq *dbSubQuery) getNextFnAlias() string {
	return fmt.Sprintf("f%d", len(dbq.fnAliases)+1)
}

func (dbq *dbSubQuery) appendSubQuery(subQry *dbSubQuery) {
	// Add all the child aliases from a subquery to a parent query
	fnIdx := len(dbq.fnAliases) + 1
	grpIdx := len(dbq.groupAliases) + 1

	for _, v := range subQry.fnAliases {
		k := fmt.Sprintf("f%d", fnIdx)
		dbq.fnAliases[k] = v
		fnIdx++
	}

	for _, v := range subQry.groupAliases {
		k := fmt.Sprintf("g%d", fnIdx)
		dbq.groupAliases[k] = v
		grpIdx++
	}
}

func (dbq *dbSubQuery) genLimit(q *dsl.Dsl) {
	l := 10
	if q.Size != nil {
		l = *q.Size
	}
	dbq.sb.Limit(l)
}

func (dbq *dbSubQuery) genAggGroupBy() {
	for k := range dbq.groupAliases {
		dbq.sb.GroupBy(k)
	}
}
