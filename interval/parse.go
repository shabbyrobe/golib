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
// as an integer followed by the span as a string, for example:
//	"1min" == interval.Interval(1, interval.Minute)
//	"1mo"  == interval.Interval(1, interval.Month)
//
// See ParseSpan for the list of supported span strings.
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
	if nidx < 0 {
		return 0, fmt.Errorf("interval: invalid input %q", intvl)
	}

	qty, err := strconv.ParseInt(intvl[:nidx+1], 10, 64)
	if err != nil {
		return 0, err
	}

	span, err := ParseSpan(intvl[nidx+1:])
	if err != nil {
		return 0, err
	}
	if err := Validate(span, Qty(qty)); err != nil {
		return 0, err
	}
	return Raw(Qty(qty), span), err
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

func Validate(span Span, qty Qty) error {
	switch span {
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
		return fmt.Errorf("interval: unknown span %s", span)
	}
	return nil
}

// Parse a span from a string.
//
// Supported span strings are:
//
//	Second == "s", "sec", "secs", "second", "seconds"
//	Minute == "min", "mins", "minute", "minutes"
//	Hour   == "h", "hr", "hrs", "hour", "hours"
//	Day    == "d", "ds", "day", "days"
//	Week   == "w", "ws", "wk", "wks", "weeks"
//	Month  == "mo", "mos", "month", "months"
//	Year   == "y", "yr", "ys", "yrs", "year", "years"
//
func ParseSpan(sstr string) (span Span, err error) {
	ips := strings.ToLower(strings.TrimSpace(sstr))
	switch ips {
	case "s":
		span = Second
	case "sec":
		span = Second
	case "secs":
		span = Second
	case "second":
		span = Second
	case "seconds":
		span = Second
	case "min":
		span = Minute
	case "mins":
		span = Minute
	case "minute":
		span = Minute
	case "minutes":
		span = Minute
	case "h":
		span = Hour
	case "hr":
		span = Hour
	case "hrs":
		span = Hour
	case "hour":
		span = Hour
	case "hours":
		span = Hour
	case "d":
		span = Day
	case "ds":
		span = Day
	case "day":
		span = Day
	case "days":
		span = Day
	case "w":
		span = Week
	case "ws":
		span = Week
	case "wk":
		span = Week
	case "wks":
		span = Week
	case "week":
		span = Week
	case "weeks":
		span = Week
	case "mo":
		span = Month
	case "mos":
		span = Month
	case "month":
		span = Month
	case "months":
		span = Month
	case "y":
		span = Year
	case "ys":
		span = Year
	case "yr":
		span = Year
	case "yrs":
		span = Year
	case "year":
		span = Year
	case "years":
		span = Year
	default:
		err = fmt.Errorf("interval: unknown span %q", sstr)
	}
	return
}
