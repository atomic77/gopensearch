package date

import (
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestEpochMillisDirect(t *testing.T) {
	i := 1668173489840
	targ := "2022-11-11T13:31:29Z"
	d1, _ := DateFormat("epoch_millis", int64(i))
	require.Equal(t, *d1, targ)

	f := 1668173489840.0
	d2, _ := DateFormat("epoch_millis", f)
	require.Equal(t, *d2, targ)

	s := "1668173489840"
	d3, _ := DateFormat("epoch_millis", s)
	require.Equal(t, *d3, targ)
}
