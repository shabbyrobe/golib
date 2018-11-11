package interval

import (
	"fmt"
	"math"
	"time"
)

// DivideNicely takes an interval and splits into an Interval that will fit at
// least n times into the original. The resulting interval will be the nearest
// 'human friendly' interval that will supply the desired number of parts.
//
// If the limit is set to a non-zero Interval, the resulting interval will
// never be less than this limit.
func DivideNicely(intvl Interval, n int, limit Interval) Interval {
	size := intvl.Time(1, nil).Sub(intvl.Time(0, nil))
	partSize := size / time.Duration(n)

	// Handle the upper limit gracefully:
	if intvl.Span() == Year && intvl.Qty() > Qty(n) {
		return Raw(intvl.Qty()/Qty(n), Year)
	}

	var limitDuration time.Duration
	if !limit.IsZero() {
		limitDuration = limit.Duration()
	}

	result := Seconds1

	var lastInterval Interval
	for _, niceSize := range niceIntervalSizes {
		if niceSize.duration < limitDuration {
			result = lastInterval
			break
		}
		if niceSize.duration <= partSize {
			result = niceSize.interval
			break
		}
		lastInterval = niceSize.interval
	}
	return result
}

// DivideNicelyFor takes an interval and splits into an Interval that will fit at
// least n times into the original, but which must be divisible by forIntvl.
// The resulting interval will be the nearest 'human friendly' interval that
// will supply the at least the desired number of parts.
//
func DivideNicelyFor(intvl Interval, n int, forIntvl Interval) (result Interval, ok bool) {
	size := intvl.Time(1, nil).Sub(intvl.Time(0, nil))
	partSize := size / time.Duration(n)

	// Handle the upper limit gracefully:
	if intvl.Span() == Year && intvl.Qty() > Qty(n) {
		return Raw(intvl.Qty()/Qty(n), Year), true
	}

	for _, niceSize := range niceIntervalSizes {
		if niceSize.duration <= partSize && (niceSize.interval == forIntvl || niceSize.interval.CanDivideBy(forIntvl)) {
			return niceSize.interval, true
		}
	}
	return forIntvl, false
}

// Find will find the smallest interval that encapsulates the duration, as
// observed at a reference time. This will not be accurate, and should only
// be used when the result is not 100% important to be correct.
//
// See FindAt for more information about caveats.
func Find(duration time.Duration) Interval {
	return FindAt(duration, intervalRefTime)
}

// FindAt will find the smallest interval that encapsulates the duration, as
// observed at the provided time.
//
// Currently, FindAt is rather naive. It will first search by Span, then work
// out how many of that span to use. This may change at some point to attempt
// several spans to find a better fit.
//
// For example, the current behaviour:
//	FindAt(86400 * time.Second) == Days1
//	FindAt(86401 * time.Second) == Days2
//
// Possible eventual behaviour (accounting for span size limits):
//	FindAt(86400 * time.Second) == Days1
//	FindAt(86401 * time.Second) == Raw(25, Hours)
//
func FindAt(duration time.Duration, at time.Time) Interval {
	if duration < 0 {
		duration = -duration
	}

	var foundSpan = Seconds
	var foundDuration time.Duration

	for span := Seconds; span <= Years; span++ {
		checkInterval := Raw(1, span)
		spanDuration := checkInterval.DurationAt(at)
		if spanDuration > duration {
			break
		}
		foundSpan, foundDuration = span, spanDuration
	}

	if foundDuration == 0 || duration <= foundDuration {
		return Raw(1, foundSpan)

	} else {
		// Integer division that 'truncates' up rather than down:
		qty := Qty((duration-1)/foundDuration + 1)

		return Raw(qty, foundSpan)
	}
}

// MUST be storted, otherwise panic!!
var niceIntervals = []Interval{
	Seconds1, Seconds2, Seconds5, Seconds10, Seconds15, Seconds30,
	Mins1, Mins2, Mins5, Mins10, Mins15, Mins30,
	Hours1, Hours2, Hours3, Hours4, Hours6, Hours8, Hours12,
	Days1, Days2, Days3, OfValid(4, Days), OfValid(5, Days), OfValid(6, Days),
	Weeks1, OfValid(2, Weeks), OfValid(3, Weeks), OfValid(4, Weeks),
	Months1, OfValid(2, Months), OfValid(3, Months), OfValid(4, Months), OfValid(6, Months),
	Years1,
}

type niceInterval struct {
	interval Interval
	duration time.Duration
}

var niceIntervalSizes []niceInterval

func init() {
	for i := len(niceIntervals) - 1; i >= 0; i-- {
		ni := niceIntervals[i]
		niceIntervalSizes = append(niceIntervalSizes, niceInterval{
			interval: ni,
			duration: ni.Duration(),
		})
	}

	last := time.Duration(math.MaxInt64)
	for _, nis := range niceIntervalSizes {
		if nis.duration < last {
			last = nis.duration
		} else {
			panic(fmt.Errorf("niceIntervals are not sorted at interval %s!", nis.interval))
		}
	}
}
