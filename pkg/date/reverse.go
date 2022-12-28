package date

import (
	"time"
)

// Inverse functions for returning back to the desired format
// from internal representation used in sqlite storage (RFC3339)
func AsDateFormat(fmt, s string) (interface{}, error) {
	switch fmt {
	case "epoch_millis":
		return asEpochMillis(s)
	}
	return nil, nil
}

func asEpochMillis(s string) (int64, error) {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return -1, err
	}
	ms := tm.UnixMilli()

	return ms, nil
}
