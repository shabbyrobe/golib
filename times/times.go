package times

import (
	"math"
	"time"
)

func ToFloat64Secs(t time.Time) float64 {
	return float64(t.Unix()) + (float64(t.Nanosecond()) / 1000000000.0)
}

func FromFloat64Secs(f float64) time.Time {
	return time.Unix(int64(f), int64(math.Mod(f, 1)*1000000000))
}

func FromFloat64SecsLocation(f float64, l *time.Location) time.Time {
	t := time.Unix(int64(f), int64(math.Mod(f, 1)*1000000000))
	return t.In(l)
}

func DurationToFloat64(d time.Duration) float64 {
	return float64(d) / float64(time.Second)
}

func DurationFromFloat64(f float64) time.Duration {
	return time.Duration(f * float64(time.Second))
}
