package times

import "time"

type TimeDiff struct {
	Years       int
	Months      int
	Days        int
	Hours       int
	Minutes     int
	Seconds     int
	Nanoseconds int
	Inverted    bool
}

// Diff calculates the difference between two dates, grouped into units of time.
//
// This (or something like it) really does belong in the standard library.
//
// It is based on this stackoverflow answer:
// https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years
//
func Diff(t1, t2 time.Time) TimeDiff {
	if t1.Location() != t2.Location() {
		t2 = t2.In(t1.Location())
	}

	var rev bool
	if t1.After(t2) {
		t1, t2, rev = t2, t1, true
	}

	year1, month1, day1 := t1.Date()
	year2, month2, day2 := t2.Date()

	hour1, min1, sec1 := t1.Clock()
	hour2, min2, sec2 := t2.Clock()

	diff := TimeDiff{
		Inverted:    rev,
		Years:       year2 - year1,
		Months:      int(month2 - month1),
		Days:        day2 - day1,
		Hours:       hour2 - hour1,
		Minutes:     min2 - min1,
		Seconds:     sec2 - sec1,
		Nanoseconds: t2.Nanosecond() - t1.Nanosecond(),
	}

	if diff.Nanoseconds < 0 {
		diff.Nanoseconds += 1e9
		diff.Seconds--
	}
	if diff.Seconds < 0 {
		diff.Seconds += 60
		diff.Minutes--
	}
	if diff.Minutes < 0 {
		diff.Minutes += 60
		diff.Hours--
	}
	if diff.Hours < 0 {
		diff.Hours += 24
		diff.Days--
	}
	if diff.Days < 0 {
		diff.Days += DaysInMonth(year2, month2-1)
		diff.Months--
	}
	if diff.Months < 0 {
		diff.Months += 12
		diff.Years--
	}

	return diff
}
