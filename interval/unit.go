package interval

import "fmt"

// These must increase numerically as the durations they represent increase in
// size. Unfortunately, intervals are not perfectly sortable as 24 months will
// still come before 1 day. The Less function has a red hot go, but it's not
// perfect either as it checks against a fixed date, but intervals can
// represent different units of time at different dates (daylight savings, leap
// seconds, etc).
const (
	Second Unit = 9
	Minute Unit = 10
	Hour   Unit = 11
	Day    Unit = 12
	Week   Unit = 13
	Month  Unit = 14
	Year   Unit = 15

	// This mistake happens so frequently there's no obvious reason not to
	// support plurals, but there may be a non-obvious one. Including for now,
	// will remove this comment if the plurals work without incident:
	Seconds Unit = Second
	Minutes Unit = Minute
	Hours   Unit = Hour
	Days    Unit = Day
	Weeks   Unit = Week
	Months  Unit = Month
	Years   Unit = Year

	// These must not exceed 255.
	MaxSecond Qty = 60
	MaxMinute Qty = 90
	MaxHour   Qty = 48
	MaxDay    Qty = 120
	MaxWeek   Qty = 52
	MaxMonth  Qty = 24
	MaxYear   Qty = 255
)

// Units contains all valid interval units in guaranteed ascending order.
var Units = []Unit{
	Second, Minute, Hour, Day, Week, Month, Year,
}

var firstUnit, lastUnit Unit

func init() {
	for i, unit := range Units {
		if i == 0 {
			firstUnit = unit
		}
		lastUnit = unit
	}
}

func (p Unit) String() string {
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

func (p Unit) MaxQty() Qty {
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

func (p Unit) IsZero() bool { return p == 0 }
