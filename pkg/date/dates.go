package date

import (
	"strconv"
	"time"
)

// Formats supported by elasticsearch
// https://www.elastic.co/guide/en/elasticsearch/reference/7.17/mapping-date-format.html
// https://github.com/elastic/elasticsearch/tree/master/server/src/main/java/org/elasticsearch/common/time

func epochMillis(val string) string {
	m, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic("couldn't parse field value " + val + " to format epochMillis")
	}
	fmtField := time.UnixMilli(m).UTC().Format(time.RFC3339)
	return fmtField
}

func epochSecond(val string) string {
	m, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic("couldn't parse field value " + val + " to format epochSecond")
	}
	fmtField := time.Unix(m, 0).UTC().Format(time.RFC3339)
	return fmtField
}

var (
	dateFormats = map[string]func(string) string{
		"epoch_millis": epochMillis,
		"epoch_second": epochSecond,
		// TODO Lots more to implement...
	}
)

func DateFormatFn(f string) func(string) string {
	if fn, ok := dateFormats[f]; ok {
		return fn
	}
	// If not found, return no-op formatter
	return func(s string) string { return s }

}
