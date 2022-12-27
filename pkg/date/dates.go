package date

import (
	"encoding/json"
	"strconv"
	"time"
)

// Formats supported by elasticsearch
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/mapping-date-format.html
// https://github.com/elastic/elasticsearch/tree/master/server/src/main/java/org/elasticsearch/common/time

type DateLike interface {
	int64 | float64 | string
}

func epochMillisString(val string) (*string, error) {
	m, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic("couldn't parse field value " + string(val) + " to format epochMillis")
	}
	return epochMillisInt(m)
}

func epochMillisFloat(val float64) (*string, error) {
	i := int64(val)
	s := time.UnixMilli(i).UTC().Format(time.RFC3339)
	return &s, nil
}

func epochMillisInt(val int64) (*string, error) {
	s := time.UnixMilli(val).UTC().Format(time.RFC3339)
	return &s, nil
}

func epochSecondString(val string) (*string, error) {
	m, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic("couldn't parse field value " + val + " to format epochSecond")
	}
	return epochSecondInt(m)
}
func epochSecondInt(val int64) (*string, error) {
	s := time.Unix(val, 0).UTC().Format(time.RFC3339)
	return &s, nil
}

// Wrapper function to DateFormatGen that will handle the type assertions on interface{}
// since we'll usually not know what we received; generics haven't been as helpful
// as expected for this purpose, so this will basically be a series of switch statements
func DateFormat(fmt string, v interface{}) (*string, error) {
	switch d := v.(type) {
	case int64:
		return dateFormatInt(fmt, d)
	case float64:
		return dateFormatFloat(fmt, d)
	case string:
		return dateFormatString(fmt, d)
	case json.Number:
		i, err := d.Int64()
		if err != nil {
			return nil, err
		}
		return dateFormatInt(fmt, i)
	default:
		return nil, nil
	}

}

func dateFormatInt(fmt string, v int64) (*string, error) {
	switch fmt {
	case "epoch_millis":
		return epochMillisInt(v)
	case "epoch_second":
		return epochSecondInt(v)
	}
	return nil, nil
}

func dateFormatFloat(fmt string, v float64) (*string, error) {
	switch fmt {
	case "epoch_millis":
		return epochMillisFloat(v)
	}
	return nil, nil
}

func dateFormatString(fmt string, v string) (*string, error) {
	switch fmt {
	case "epoch_millis":
		return epochMillisString(v)
	case "epoch_second":
		return epochSecondString(v)
	}
	return nil, nil
}
