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
	aggInfo      *aggregateInfo
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

// Not clear if we need both of these
type aggregateInfo struct {
	subAggregateInfo *aggregateInfo
}

func makeAggregateInfo() aggregateInfo {
	a := aggregateInfo{}
	return a
}
func GenPlan(index string, q *dsl.Dsl) ([]dbSubQuery, error) {

	plan := make([]dbSubQuery, 0)

	for _, a := range q.Aggs {
		aggQ := makeDbSubQuery()
		aggInfo := aggQ.genAggregateSelectExprs(a)

		aggQ.genSelectExpression()
		aggQ.sb.From(fmt.Sprintf(`"%s"`, index))
		aggQ.genQueryWherePredicates(q)
		aggQ.genAggGroupBy(aggInfo)
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
	fmtFn := date.DateFormatFn(*rng.RangeOptions.Format)

	if rng.RangeOptions.Lte != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) <= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lte)))
	} else if rng.RangeOptions.Lt != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) < '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Lt)))
	}

	if rng.RangeOptions.Gte != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) >= '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gte)))
	} else if rng.RangeOptions.Gt != nil {
		dbq.sb.Where(fmt.Sprintf(` DATETIME(JSON_EXTRACT(content, '$.%s')) > '%s' `, rng.Field, fmtFn(*rng.RangeOptions.Gt)))
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
func (dbq *dbSubQuery) genAggregateSelectExprs(root *dsl.Aggregate) aggregateInfo {

	ai := makeAggregateInfo()

	for _, agg := range root.AggregateType {
		grpIdx := fmt.Sprintf("g%d", len(dbq.groupAliases)+1)
		fnIdx := fmt.Sprintf("f%d", len(dbq.fnAliases)+1)

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
			// TODO Currently only support sub-aggregations that can be mapped
			// to a non-nested SQL statement, so we'll "absorb" the first subagg here
			subAi := dbq.genAggregateSelectExprs(agg.Aggs[0])
			ai.subAggregateInfo = &subAi
		}
	}
	return ai
}

func (dbq *dbSubQuery) genLimit(q *dsl.Dsl) {
	l := 10
	if q.Size != nil {
		l = *q.Size
	}
	dbq.sb.Limit(l)
}

func (dbq *dbSubQuery) genAggGroupBy(ai aggregateInfo) {
	for k := range dbq.groupAliases {
		dbq.sb.GroupBy(k)
	}
}
