package dsl

// Custom json handling methods to deal with all the wacky ways ES allows users
// to submit queries

import "encoding/json"

func (jq *JQuery) UnmarshalJSON(b []byte) error {
	// ES accepts a shorthand version of the match structure, so use this custom unmarshaller
	// to transform what's in the "Raw" match field to the field we'll use internally
	// More info: https://www.elastic.co/guide/en/elasticsearch/reference/7.17/query-dsl-match-query.html#query-dsl-match-query-short-ex
	type JQuery_ JQuery
	var base JQuery_

	if err := json.Unmarshal(b, &base); err != nil {
		return err
	}

	jq.Term = base.Term
	jq.Bool = base.Bool
	jq.Range = base.Range

	if len(base.RawMatch) > 0 {

		jq.Match = make(map[string]JMatch, len(base.RawMatch))

		for k, rawVal := range base.RawMatch {
			if v, ok := rawVal.(string); ok {
				jm := JMatch{Query: v}
				jq.Match[k] = jm
			} else {
				// I can't find any better way to re-parse this map[string]interface{}
				// back to our struct than to redo the serialization from JSON.
				s, err := json.Marshal(rawVal)
				if err != nil {
					return err
				}
				jm := JMatch{}
				if err = json.Unmarshal([]byte(s), &jm); err != nil {
					return err
				}
				jq.Match[k] = jm
			}
		}
	}

	return nil
}

func (jd *JDsl) UnmarshalJSON(b []byte) error {
	type JDsl_ JDsl
	var base JDsl_

	if err := json.Unmarshal(b, &base); err != nil {
		return err
	}
	jd.Query = base.Query
	jd.Size = base.Size
	if len(base.RawAggregations) > 0 {
		jd.Aggs = base.RawAggregations
	} else if len(base.RawAggs) > 0 {
		jd.Aggs = base.RawAggs
	}

	return nil
}
