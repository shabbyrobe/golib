package interval

import (
	"fmt"
	"strconv"
	"time"

	"github.com/shabbyrobe/golib/times"
)

type (
	// Span identifies the unit of the Interval, i.e. Seconds, Minutes, etc.
	Span uint8

	// Qty is the number of Spans in an Interval.
	Qty uint

	// Period identifies the number of Intervals that have passed since the
	// Unix Epoch.
	Period int64

	// Interval combines a Span and a Qty into a single value.
	Interval uint16
)

// These must increase numerically as the durations they represent increase in
// size. Unfortunately, intervals are not perfectly sortable as 24 months will
// still come before 1 day. The Less function has a red hot go, but it's not
// perfect either as it checks against a fixed date, but intervals can
// represent different spans of time at different dates (daylight savings, leap
// seconds, etc).
const (
	Second Span = 9
	Minute Span = 10
	Hour   Span = 11
	Day    Span = 12
	Week   Span = 13
	Month  Span = 14
	Year   Span = 15

	// This mistake happens so frequently there's no obvious reason not to
	// support plurals, but there may be a non-obvious one. Including for now,
	// will remove this comment if the plurals work without incident:
	Seconds Span = Second
	Minutes Span = Minute
	Hours   Span = Hour
	Days    Span = Day
	Weeks   Span = Week
	Months  Span = Month
	Years   Span = Year

	// These must not exceed 255.
	MaxSecond Qty = 60
	MaxMinute Qty = 90
	MaxHour   Qty = 48
	MaxDay    Qty = 120
	MaxWeek   Qty = 52
	MaxMonth  Qty = 24
	MaxYear   Qty = 255
)

// Spans contains all valid interval spans in guaranteed ascending order.
var Spans = []Span{
	Second, Minute, Hour, Day, Week, Month, Year,
}

// intervalRefTime is used for sorting. It is an imperfect mechanism to sort
// intervals that may have different Qtys, i.e. 25 hours should come after 1
// day. It can not account for leap-seconds or leap-years.
var intervalRefTime = time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)

var epochTime = time.Unix(0, 0)

// Of returns a valid interval from a Qty and a Span, or an error indicating
// why one could not be created.
func Of(qty Qty, span Span) (Interval, error) {
	err := Validate(span, qty)
	if err != nil {
		return 0, err
	}
	return Interval((uint(span) << 8) | uint(qty)), nil
}

// OfValid returns a guaranteed valid interval from a Qty and a Span, panicking
// if it is not valid.
func OfValid(qty Qty, span Span) Interval {
	i, err := Of(qty, span)
	if err != nil {
		panic(err)
	}
	return i
}

// Raw returns an unchecked Interval from a Qty and a Span, which may be
// invalid.
func Raw(qty Qty, span Span) Interval {
	return Interval((uint(span) << 8) | uint(qty))
}

// FormatIntervalPeriod is the complement to ParseIntervalPeriod.
func FormatIntervalPeriod(intvl Interval, period Period) string {
	return intvl.String() + ":" + strconv.FormatInt(int64(period), 10)
}

func (p Period) Format(intvl Interval) string {
	return FormatIntervalPeriod(intvl, p)
}

func (q Qty) IsZero() bool { return q == 0 }

func (i Interval) String() string {
	if is, ok := intervalStrings[i]; ok {
		return is
	}
	return fmt.Sprintf("%d%s", i.Qty(), i.Span().String())
}

func (i Interval) IsZero() bool { return i == 0 }

func (i Interval) Span() Span {
	return Span(i >> 8)
}

// Less returns a best-effort guess as to whether one interval is smaller than
// another. It is not 100% guaranteed to be accurate as it uses a reference
// time.
func (i Interval) Less(j Interval) bool {
	return i.LessAt(j, intervalRefTime)
}

// LessAt returns whether one interval is less than another at the supplied
// reference time.
func (i Interval) LessAt(j Interval, at time.Time) bool {
	iStart := i.Start(at)
	jStart := j.Start(at)

	iNext := i.Next(iStart)
	jNext := j.Next(jStart)

	return iNext.Sub(iStart) < jNext.Sub(jStart)
}

func (i Interval) Qty() Qty {
	return Qty(i & 0xFF)
}

func (i Interval) Valid() bool {
	return Validate(i.Span(), i.Qty()) == nil
}

// CanDivide reports whether this interval can cleanly subdivide into the 'by'
// interval. For example, 4 hours can combine cleanly to 1 day, but 7 hours
// cannot. The "4 hours" part in this example is the "by" interval, and the "1 day"
// part is the receiver.
//
// This returns false if i == by.
func (i Interval) CanDivideBy(by Interval) bool {
	return by.CanCombineTo(i)
}

// CanCombine reports whether this interval represents a clean subdivision of
// the 'to' interval. For example, 4 hours can combine cleanly to 1 day, but 7
// hours cannot.
//
// The "4 hours" part in this example is the receiver, and the "1 day" part is
// the "to" interval.
//
// This returns false if i == to.
func (i Interval) CanCombineTo(to Interval) bool {
	if !i.Less(to) {
		return false
	}

	fromSpan := i.Span()
	toSpan := to.Span()

	if fromSpan == Week && toSpan > Week {
		return false
	} else if toSpan == Week && fromSpan > Week {
		return false
	}

	startOfPeriod := i.Time(0, nil)
	startOfToPeriod := to.Time(0, nil)

	// Some periods have a start time for the 0-period that isn't exactly
	// the epoch.
	var offset = startOfToPeriod.Sub(startOfPeriod)

	endOfToPeriod := to.Time(1, nil).Add(offset)
	for p := Period(0); ; p++ {
		inTime := i.Time(p, nil)
		if inTime.Equal(endOfToPeriod) {
			return true
		} else if inTime.After(endOfToPeriod) {
			return false
		}
	}
}

func (i Interval) Duration() time.Duration {
	return i.Time(1, nil).Sub(i.Time(0, nil))
}

func (i Interval) DurationAt(at time.Time) time.Duration {
	start, end := i.Range(at)
	return end.Sub(start)
}

func (i Interval) Period(t time.Time) Period {
	qty := int64(i.Qty())

	var out int64
	switch i.Span() {
	case Second:
		un := t.UnixNano()
		if un >= 0 {
			out = ((un - (un % (int64(time.Second) * qty))) / int64(time.Second)) / qty
		} else {
			out = un
			gap := un % (int64(time.Second) * qty)
			if gap != 0 {
				out -= (int64(time.Second) * qty) + gap
			}
			out = out / int64(time.Second) / qty
		}

	case Minute:
		un := t.UnixNano()
		if un >= 0 {
			out = ((un - (un % (int64(time.Minute) * qty))) / int64(time.Minute)) / qty
		} else {
			out = un
			gap := un % (int64(time.Minute) * qty)
			if gap != 0 {
				out -= (int64(time.Minute) * qty) + gap
			}
			out = out / int64(time.Minute) / qty
		}

	case Hour:
		un := t.UnixNano()
		if un >= 0 {
			out = ((un - (un % (int64(time.Hour) * qty))) / int64(time.Hour)) / qty
		} else {
			out = un
			gap := un % (int64(time.Hour) * qty)
			if gap != 0 {
				out -= (int64(time.Hour) * qty) + gap
			}
			out = out / int64(time.Hour) / qty
		}

	case Day:
		un := t.UnixNano()
		if un >= 0 {
			out = ((un - (un % (int64(time.Hour) * 24 * qty))) / int64(time.Hour)) / qty / 24
		} else {
			out = un
			gap := un % (int64(time.Hour) * 24 * qty)
			if gap != 0 {
				out -= (int64(time.Hour) * 24 * qty) + gap
			}
			out = out / int64(time.Hour) / 24 / qty
		}

	case Week:
		out = int64(times.PeriodWeeks(t, int(qty)))

	case Month:
		out = int64(times.PeriodMonths(t, int(qty)))

	case Year:
		y := int64(t.Year()) - 1970
		if y >= 0 {
			out = (y - (y % qty)) / qty
		} else {
			out = y
			gap := y % qty
			if gap != 0 {
				out -= qty + gap
			}
			out = out / qty
		}

	default:
		panic(fmt.Errorf("unknown span %d", i.Span()))
	}

	return Period(out)
}

func (i Interval) Time(p Period, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	qty := int64(i.Qty())

	var out time.Time
	switch i.Span() {
	case Second:
		out = time.Unix(int64(p)*qty, 0).In(loc)
	case Minute:
		out = time.Unix(int64(p)*60*qty, 0).In(loc)
	case Hour:
		out = time.Unix(int64(p)*3600*qty, 0).In(loc)
	case Day:
		out = time.Unix(int64(p)*86400*qty, 0).In(loc)
	case Week:
		out = times.PeriodWeeksTime(int(p), int(qty), loc)
	case Month:
		out = times.PeriodMonthsTime(int(p), int(qty), loc)
	case Year:
		out = time.Date(int(int64(p)*qty)+1970, 1, 1, 0, 0, 0, 0, loc)
	default:
		panic(fmt.Errorf("unknown span %d", i.Span()))
	}

	return out
}

// Start returns the time that represents the inclusive start of the Period
// that contains t.
func (i Interval) Start(t time.Time) time.Time {
	un := t.UnixNano()

	qty := int64(i.Qty())

	var out time.Time
	switch i.Span() {
	case Second:
		out = time.Unix(0, un-(un%(int64(time.Second)*qty)))
	case Minute:
		out = time.Unix(0, un-(un%(int64(time.Minute)*qty)))
	case Hour:
		out = time.Unix(0, un-(un%(int64(time.Hour)*qty)))
	case Day:
		out = time.Unix(0, un-(un%(int64(time.Hour)*24*qty)))
	case Week:
		out = times.TruncateWeeks(t, int(qty))
	case Month:
		out = times.TruncateMonths(t, int(qty))
	case Year:
		out = time.Date(t.Year()-(t.Year()%int(qty)), 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		panic(fmt.Errorf("unknown span %d", i.Span()))
	}

	return out.In(t.Location())
}

// End returns the time that represents the exclusive end of the Period
// that contains t, such that End(t) == Next(t).
func (i Interval) End(t time.Time) time.Time {
	// FIXME: is this functionally identical to Next()?

	qty := int64(i.Qty())

	var out time.Time
	switch i.Span() {
	case Second:
		out = i.Start(t).Add(time.Duration(qty) * time.Second)
	case Minute:
		out = i.Start(t).Add(time.Duration(qty) * time.Minute)
	case Hour:
		out = i.Start(t).Add(time.Duration(qty) * time.Hour)
	case Day:
		out = i.Start(t).Add(time.Duration(qty) * 24 * time.Hour)
	case Week:
		out = i.Start(t).Add(time.Duration(qty) * 7 * 24 * time.Hour)
	case Month:
		out = times.AddMonths(t, int(qty))
	case Year:
		out = times.AddYears(t, int(qty))
	default:
		panic(fmt.Errorf("unknown span %d", i.Span()))
	}

	return out.In(t.Location())
}

// Range returns the start and end time for the period represented by the passed-in time.
//
// For example, interval.Hours1.Range(time.Time(1970, 1, 1, 0, 30, 0, 0, time.UTC)) will
// return 1970-01-01T00:00Z and 1970-01-01T01:00Z.
//
func (i Interval) Range(t time.Time) (start, end time.Time) {
	start, end = i.Start(t), i.End(t)
	return
}

// Next returns the time at the beginning of the period that starts after the
// period that encapsulates the passed-in time.
func (i Interval) Next(t time.Time) time.Time {
	return i.Time(i.Period(t)+1, t.Location())
}

// Prev will return the Period prior to the Period that t is contained within.
//
// For a daily interval:
//
//	t == 2018-01-03T00:0Z0, Prev(t) == 2018-01-02T00:00Z
//	t == 2018-01-03T10:30Z, Prev(t) == 2018-01-02T00:00Z
//	t == 2018-01-03T23:59Z, Prev(t) == 2018-01-02T00:00Z
//	t == 2018-01-04T00:00Z, Prev(t) == 2018-01-03T00:00Z
//
func (i Interval) Prev(t time.Time) time.Time {
	return i.Time(i.Period(t)-1, t.Location())
}

/*
// FIXME: Breaks existing serialised representations; probably needs
// to handle integers too if it can.
func (i Interval) MarshalText() (text []byte, err error) {
	return []byte(i.String()), nil
}

func (i *Interval) UnmarshalText(text []byte) (err error) {
	*i, err = Parse(string(text))
	return err
}
*/

func (p Span) String() string {
	switch p {
	case Second:
		return "sec"
	case Minute:
		return "min"
	case Hour:
		return "hr"
	case Day:
		return "d"
	case Week:
		return "wk"
	case Month:
		return "mo"
	case Year:
		return "yr"
	default:
		return fmt.Sprintf("Unknown(%d)", p)
	}
}

func (p Span) IsZero() bool { return p == 0 }
