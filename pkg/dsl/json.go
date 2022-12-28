package dsl

// Custom json handling methods to deal with all the wacky ways ES allows users
// to submit queries

import "encoding/json"

func (jq *Query) UnmarshalJSON(b []byte) error {
	// ES accepts a shorthand version of the match structure, so use this custom unmarshaller
	// to transform what's in the "Raw" match field to the field we'll use internally
	// More info: https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html#query-dsl-match-query-short-ex
	type Query_ Query
	var base Query_

	if err := json.Unmarshal(b, &base); err != nil {
		return err
	}

	jq.Bool = base.Bool
	jq.Range = base.Range

	if len(base.RawMatch) > 0 {

		jq.Match = make(map[string]Match, len(base.RawMatch))

		for k, rawVal := range base.RawMatch {
			if v, ok := rawVal.(string); ok {
				jm := Match{Query: v}
				jq.Match[k] = jm
			} else {
				// I can't find any better way to re-parse this map[string]interface{}
				// back to our struct than to redo the serialization from JSON.
				s, err := json.Marshal(rawVal)
				if err != nil {
					return err
				}
				jm := Match{}
				if err = json.Unmarshal([]byte(s), &jm); err != nil {
					return err
				}
				jq.Match[k] = jm
			}
		}
	}

	if len(base.RawTerm) > 0 {

		jq.Term = make(map[string]Term, len(base.RawTerm))

		for k, rawVal := range base.RawTerm {
			if v, ok := rawVal.(string); ok {
				jm := Term{Value: v}
				jq.Term[k] = jm
			} else {
				s, err := json.Marshal(rawVal)
				if err != nil {
					return err
				}
				jm := Term{}
				if err = json.Unmarshal([]byte(s), &jm); err != nil {
					return err
				}
				jq.Term[k] = jm
			}
		}
	}

	return nil
}

func (jd *Dsl) UnmarshalJSON(b []byte) error {
	type JDsl_ Dsl
	var base JDsl_

	if err := json.Unmarshal(b, &base); err != nil {
		return err
	}
	jd.Query = base.Query
	jd.Size = base.Size
	jd.Sort = base.Sort
	if len(base.RawAggregations) > 0 {
		jd.Aggs = base.RawAggregations
	} else if len(base.RawAggs) > 0 {
		jd.Aggs = base.RawAggs
	}

	return nil
}

func (bl *Bool) UnmarshalJSON(b []byte) error {
	type Bool_ Bool
	var base Bool_
	var err error

	if err = json.Unmarshal(b, &base); err != nil {
		return err
	}
	m1 := Query{}
	m2 := make([]Query, 0)

	// Must can be provided with a single object or an array
	err = json.Unmarshal(base.RawMust, &m1)
	if err == nil {
		bl.Must = append(bl.Must, m1)
		return nil
	}

	err = json.Unmarshal(base.RawMust, &m2)
	if err == nil {
		bl.Must = m2
		return nil
	}
	// TODO This approach to handling these variable types needs to be
	// generalized somehow
	// Should can also be provided with a single object or an array
	m1 = Query{}
	m2 = make([]Query, 0)
	err = json.Unmarshal(base.RawShould, &m1)
	if err == nil {
		bl.Should = append(bl.Should, m1)
		return nil
	}

	err = json.Unmarshal(base.RawShould, &m2)
	if err == nil {
		bl.Should = m2
		return nil
	}

	return err
}
