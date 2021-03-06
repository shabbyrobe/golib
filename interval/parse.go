package interval

import (
	"fmt"
	"strconv"
	"strings"
)

func MustParse(intvl string) Interval {
	p, err := Parse(intvl)
	if err != nil {
		panic(err)
	}
	return p
}

// Parse an interval from a string representation of the interval size
// as an integer followed by the unit as a string, for example:
//	"1min" == interval.Interval(1, interval.Minute)
//	"1mo"  == interval.Interval(1, interval.Month)
//
// See ParseUnit for the list of supported unit strings.
//
func Parse(intvl string) (Interval, error) {
	intvl = strings.TrimSpace(intvl)
	nidx := -1
	for idx, c := range intvl {
		if c < '0' || c > '9' {
			break
		}
		nidx = idx
	}

	var qty int64
	if nidx < 0 {
		qty = 1

	} else {
		var err error
		qty, err = strconv.ParseInt(intvl[:nidx+1], 10, 64)
		if err != nil {
			return 0, err
		}
	}

	unit, err := ParseUnit(intvl[nidx+1:])
	if err != nil {
		return 0, err
	}
	if err := Validate(unit, Qty(qty)); err != nil {
		return 0, err
	}
	return Raw(Qty(qty), unit), err
}

// ParseIntervalPeriod parses a string representing an interval combined with a
// period using a colon, in the format "<interval>:<period>".
//
// The values allowed for "<interval>" are defined by interval.Parse(). The
// value for "<period>" must be parseable by strconv.ParseInt().
//
// See FormatIntervalPeriod for the complement.
func ParseIntervalPeriod(v string) (intvl Interval, period Period, err error) {
	var pi int64

	i := strings.IndexByte(v, ':')
	if i < 0 {
		goto fail
	}
	intvl, err = Parse(v[:i])
	if err != nil {
		goto fail
	}
	pi, err = strconv.ParseInt(v[i+1:], 10, 64)
	if err != nil {
		goto fail
	}
	return intvl, Period(pi), nil

fail:
	return 0, 0, fmt.Errorf("interval: invalid interval/period %q; expected format '1min:1234'", v)
}

func Validate(unit Unit, qty Qty) error {
	switch unit {
	case Second:
		if qty > MaxSecond {
			return fmt.Errorf("interval: qty too large for seconds: expected <= %d, found %d", MaxSecond, qty)
		}
	case Minute:
		if qty > MaxMinute {
			return fmt.Errorf("interval: qty too large for minutes: expected <= %d, found %d", MaxMinute, qty)
		}
	case Hour:
		if qty > MaxHour {
			return fmt.Errorf("interval: qty too large for hours: expected <= %d, found %d", MaxHour, qty)
		}
	case Day:
		if qty > MaxDay {
			return fmt.Errorf("interval: qty too large for days: expected <= %d, found %d", MaxDay, qty)
		}
	case Week:
		if qty > MaxWeek {
			return fmt.Errorf("interval: qty too large for weeks: expected <= %d, found %d", MaxWeek, qty)
		}
	case Month:
		if qty > MaxMonth {
			return fmt.Errorf("interval: qty too large for months: expected <= %d, found %d", MaxMonth, qty)
		}
	case Year:
		if qty > MaxYear {
			return fmt.Errorf("interval: qty too large for years: expected <= %d, found %d", MaxYear, qty)
		}
	default:
		return fmt.Errorf("interval: unknown unit %s", unit)
	}
	return nil
}

// Parse a unit from a string.
//
// Supported unit strings are:
//
//	Second == "s", "sec", "secs", "second", "seconds"
//	Minute == "min", "mins", "minute", "minutes"
//	Hour   == "h", "hr", "hrs", "hour", "hours"
//	Day    == "d", "ds", "day", "days"
//	Week   == "w", "ws", "wk", "wks", "weeks"
//	Month  == "mo", "mos", "month", "months"
//	Year   == "y", "yr", "ys", "yrs", "year", "years"
//
func ParseUnit(sstr string) (unit Unit, err error) {
	ips := strings.ToLower(strings.TrimSpace(sstr))
	switch ips {
	case "s", "sec", "secs", "second", "seconds":
		unit = Second

	case "min", "mins", "minute", "minutes":
		unit = Minute

	case "h", "hr", "hrs", "hour", "hours":
		unit = Hour

	case "d", "ds", "day", "days":
		unit = Day

	case "w", "ws", "wk", "wks", "week", "weeks":
		unit = Week

	case "mo", "mos", "mnth", "mnths", "month", "months":
		unit = Month

	case "y", "ys", "yr", "yrs", "year", "years":
		unit = Year

	default:
		err = fmt.Errorf("interval: unknown unit %q", sstr)
	}
	return
}
