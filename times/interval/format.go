package interval

import (
	"fmt"
	"strconv"
	"time"
)

func (i Interval) Format(p Period) string      { return i.FormatIn(p, time.UTC) }
func (i Interval) FormatShort(p Period) string { return i.FormatShortIn(p, time.UTC) }

// FormatAfter returns the shortest string representation of the date that
// expresses all of the date fields that have changed between current
// and prev that are relevant to the span of the interval.
//
func (i Interval) FormatAfter(current Period, prev Period) string {
	return i.FormatAfterIn(current, prev, time.UTC)
}

func (i Interval) FormatIn(p Period, in *time.Location) string {
	switch i.Span() {
	case Second:
		return i.Time(p, in).Format(time.RFC3339)
	case Minute:
		return i.Time(p, in).Format("2006-01-02T15:04Z07:00")
	case Hour:
		return i.Time(p, in).Format("2006-01-02T15:00Z07:00")
	case Day:
		return i.Time(p, in).Format("2006-01-02Z07:00")
	case Week:
		return i.Time(p, in).Format("2006-01-02Z07:00")
	case Month:
		return i.Time(p, in).Format("2006-01-02Z07:00")
	case Year:
		return i.Time(p, in).Format("2006Z07:00")
	default:
		return fmt.Sprintf("unknown span %d for period %d", i.Span(), p)
	}
}

func (i Interval) FormatShortIn(p Period, in *time.Location) string {
	var tm = i.Time(p, in)

	if tm.Second() != 0 || tm.Minute() != 0 || tm.Hour() != 0 {
		if i.Span() == Second {
			return tm.Format("15:04:05")
		} else {
			return tm.Format("15:04")
		}

	} else if i.Span() == Week {
		return tm.Format("2006-01-02")

	} else {
		var firstDay = tm.Day() == 1
		if tm.Month() == 1 && firstDay {
			return strconv.FormatInt(int64(tm.Year()), 10)

		} else if firstDay {
			return tm.Format("2006-01")

		} else {
			return tm.Format("2006-01-02")
		}
	}
}

func (i Interval) FormatAfterIn(current Period, prev Period, in *time.Location) string {
	var curTime = i.Time(current, in)
	var prevTime = i.Time(prev, in)

	if !curTime.After(prevTime) {
		return curTime.Format("2006-01-02T15:04:05Z")
	}

	switch i.Span() {
	case Year, Month, Day:
		yrEq := curTime.Year() == prevTime.Year()
		moEq := curTime.Month() == prevTime.Month()
		dyEq := curTime.Day() == prevTime.Day()

		if !yrEq {
			if moEq && dyEq {
				return strconv.FormatInt(int64(curTime.Year()), 10)
			} else if dyEq {
				return curTime.Format("2006-01")
			} else {
				return curTime.Format("2006-01-02")
			}
		} else {
			return curTime.Format("02-Jan")
		}

	case Week:
		return curTime.Format("2006-01-02")

	case Hour, Minute, Second:
		var dfmt, tfmt, tjoin string
		var showTime bool

		yrEq := curTime.Year() == prevTime.Year()
		moEq := curTime.Month() == prevTime.Month()
		dyEq := curTime.Day() == prevTime.Day()

		if !yrEq || !moEq || !dyEq {
			if yrEq {
				dfmt, tjoin = "02-Jan", " "
			} else {
				dfmt, tjoin = "2006-01-01", "T"
			}
		}

		hrEq := curTime.Hour() == prevTime.Hour()
		mnEq := curTime.Minute() == prevTime.Minute()
		scEq := curTime.Second() == prevTime.Second()

		if i.Span() == Second && curTime.Second() != 0 {
			tfmt = "15:04:05"
			showTime = !hrEq || !mnEq || !scEq
		} else if !hrEq || !mnEq {
			tfmt = "15:04"
			showTime = !hrEq || !mnEq
		}

		if dfmt == "" && !showTime {
			return curTime.Format(tfmt)
		} else if !showTime {
			return curTime.Format(dfmt)
		} else {
			return curTime.Format(dfmt + tjoin + tfmt)
		}

	default:
		return curTime.Format("2016-01-02T15:04:05.999999999")
	}
}
