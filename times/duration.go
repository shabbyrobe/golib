package times

import (
	"time"
)

var (
	firstMondayOfEpochWeek = FirstMondayOfWeek(time.Unix(0, 0))
	epoch                  = time.Unix(0, 0)
)

const (
	week = 24 * 7 * time.Hour
)

func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func FirstMondayOfISOWeek(t time.Time) (tm time.Time, year, week int) {
	s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if s.Weekday() != time.Monday {
		diff := int(s.Weekday() - time.Monday)
		if diff < 0 {
			diff = 6
		}
		s = s.Add(-(24 * time.Duration(diff) * time.Hour))
	}
	y, w := s.ISOWeek()
	return s, y, w
}

func FirstMondayOfWeek(t time.Time) time.Time {
	s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if s.Weekday() != time.Monday {
		diff := int(s.Weekday() - time.Monday)
		if diff < 0 {
			diff = 6
		}
		s = s.Add(-(24 * time.Duration(diff) * time.Hour))
	}
	return s
}

func TruncateWeeks(t time.Time, n int) time.Time {
	p := PeriodWeeks(t, n)
	return PeriodWeeksTime(p, n, t.Location())
}

// PeriodWeeks returns a monotonically increasing/decreasing integer that
// represents a period of n weeks since the Unix epoch.
func PeriodWeeks(t time.Time, n int) int {
	ts := FirstMondayOfWeek(t)
	diff := ts.Sub(firstMondayOfEpochWeek)
	weeks := int(diff / week)

	var gap int
	if diff >= 0 {
		gap = weeks - (weeks % n)
	} else {
		gap = weeks - n - (weeks % n)
	}
	return gap / n
}

// PeriodWeeksTime returns a time for the integer identifying the period of
// n weeks since the Unix epoch.
func PeriodWeeksTime(p int, n int, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	p *= n
	out := firstMondayOfEpochWeek.Add(time.Duration(p) * week)
	out = time.Date(out.Year(), out.Month(), out.Day(), 0, 0, 0, 0, loc)
	return out
}

func TruncateMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// TruncateMonths will round the time down to the start of the period of
// n months, for example:
//	TruncateMonths("2018-08-22", 3) == "2018-07-01".
//
// If n == (0, 1), this call is equivalent to TruncateMonth(t)
//
func TruncateMonths(t time.Time, n int) time.Time {
	if n == 0 || n == 1 {
		// If TruncateMonths is called with n == 0, it seems the most natural
		// thing to do is presume 1 because the caller has explicitly asked for
		// truncation to months.
		return TruncateMonth(t)

	} else {
		inMnth := (t.Year() * 12) + (int(t.Month()) - 1)

		var oy, om int
		if inMnth >= 0 {
			ms := inMnth - (inMnth % n)
			oy = (ms / 12)
			om = (ms % 12) + 1
		} else {
			ms := inMnth + (inMnth % n)
			oy = (ms / 12) - 1
			om = 12 - (-ms % 12) + 1
		}

		return time.Date(oy, time.Month(om), 1, 0, 0, 0, 0, t.Location())
	}
}

// PeriodMonth returns the period value in month quantities for the given date.
func PeriodMonth(t time.Time) int {
	return ((t.Year() - 1970) * 12) + (int(t.Month()) - 1)
}

// PeriodMonths will return the number of n month periods that have fully
// elapsed since the unix epoch, for example:
//	PeriodMonths("1970-07-01", 3) == 2
//	PeriodMonths("1970-08-22", 3) == 2
//	PeriodMonths("1970-10-01", 3) == 3
//
// If n == (0, 1), this call is equivalent to PeriodMonth(t)
//
func PeriodMonths(t time.Time, n int) int {
	if n == 0 {
		n = 1
	}

	inMnth := ((t.Year() - 1970) * 12) + (int(t.Month()) - 1)
	if inMnth >= 0 {
		return (inMnth - (inMnth % n)) / n
	} else {
		out := inMnth
		gap := inMnth % n
		if gap != 0 {
			out -= n + gap
		}
		return out / n
	}
}

func PeriodMonthsTime(p int, n int, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	p *= n
	ms := p

	var oy, om int
	if p >= 0 {
		oy = (ms / 12)
		om = (ms % 12) + 1
	} else {
		oy = (ms / 12) - 1
		om = 12 - (-ms % 12) + 1
	}

	oy += 1970

	return time.Date(oy, time.Month(om), 1, 0, 0, 0, 0, loc)
}

func AddMonths(t time.Time, n int) time.Time {
	inMnth := (t.Year() * 12) + (int(t.Month()) - 1)
	trg := inMnth + n

	var oy, om int
	oy = (trg / 12)
	om = (trg % 12) + 1

	return time.Date(oy, time.Month(om), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func AddYears(t time.Time, n int) time.Time {
	return time.Date(t.Year()+n, t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
