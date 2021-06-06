package times

import (
	"strings"
	"time"
)

// Converts a time to a string representation which is lexicographically
// comparable to another time formatted by this function.
//
// "2020-01-01T12:00:00.000000001Z" >= "2020-01-01T12:00:00.000000000Z"
//
func TimeToComparableRFC3339(tm time.Time) string {
	const trail = ".000000000Z"

	if tm.Location() != time.UTC {
		tm = tm.In(time.UTC)
	}

	tstr := tm.Format("2006-01-02T15:04:05.999999999")
	idx := strings.IndexByte(tstr, '.')
	if idx < 0 {
		return tstr + trail
	}
	places := (len(tstr) - idx)
	return tstr + trail[places:]
}
