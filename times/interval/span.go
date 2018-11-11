package interval

import "fmt"

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

var firstSpan, lastSpan Span

func init() {
	for i, span := range Spans {
		if i == 0 {
			firstSpan = span
		}
		lastSpan = span
	}
}

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

func (p Span) MaxQty() Qty {
	switch p {
	case Second:
		return MaxSecond
	case Minute:
		return MaxMinute
	case Hour:
		return MaxHour
	case Day:
		return MaxDay
	case Week:
		return MaxWeek
	case Month:
		return MaxMonth
	case Year:
		return MaxYear
	default:
		return 0
	}
}

func (p Span) IsZero() bool { return p == 0 }
