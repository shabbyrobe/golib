package times

import (
	"math/rand"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func TestFirstDayOfWeek(t *testing.T) {
	tt := assert.WrapTB(t)

	iter := 20000

	start := time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC)

	for i := 0; i < iter; i++ {
		f := start.Add(time.Duration(i) * 24 * time.Hour)
		n := FirstMondayOfWeek(f)
		tt.MustEqual(time.Monday, n.Weekday(), f.String())
	}

	for i := 0; i < iter; i++ {
		f := start.Add(time.Duration(i) * 24 * time.Hour)
		f = f.Add(time.Duration(rand.Intn(60)) * time.Minute)
		f = f.Add(time.Duration(rand.Intn(24)) * time.Hour)
		n := FirstMondayOfWeek(f)
		tt.MustEqual(time.Monday, n.Weekday(), f.String())
	}
}

func TestTruncateMonths(t *testing.T) {
	tt := assert.WrapTB(t)

	tm := TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 2)
	exp := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1970, 2, 3, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(-1, 12, 31, 23, 59, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(-1, 11, 01, 23, 59, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(-1, 11, 01, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(-1, 10, 31, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 9, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	// 3 months:

	tm = TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1970, 2, 3, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1970, 3, 3, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1970, 4, 1, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 4, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1969, 11, 01, 23, 59, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1969, 11, 01, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1969, 10, 31, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateMonths(time.Date(1969, 9, 30, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 7, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)
}

func TestTruncateWeeks(t *testing.T) {
	tt := assert.WrapTB(t)

	tm := TruncateWeeks(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 2)
	exp := time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 1, 21, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 1, 28, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(2017, 1, 28, 12, 30, 0, 0, time.UTC), 2)
	exp = time.Date(2017, 1, 23, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 16, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 14, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	// 3 weeks
	tm = TruncateWeeks(time.Date(1970, 1, 21, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 1, 28, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 2, 5, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1970, 2, 12, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 2, 9, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 16, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)

	tm = TruncateWeeks(time.Date(1969, 12, 7, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 11, 17, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(exp, tm)
}

func TestAddMonths(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual(time.Date(2005, 1, 1, 10, 9, 8, 7, time.UTC),
		AddMonths(time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), 2))

	tt.MustEqual(time.Date(2004, 9, 1, 10, 9, 8, 7, time.UTC),
		AddMonths(time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), -2))

	tt.MustEqual(time.Date(0, 1, 1, 10, 9, 8, 7, time.UTC),
		AddMonths(time.Date(-1, 11, 1, 10, 9, 8, 7, time.UTC), 2))

	tt.MustEqual(time.Date(-1, 1, 1, 10, 9, 8, 7, time.UTC),
		AddMonths(time.Date(-2, 11, 1, 10, 9, 8, 7, time.UTC), 2))
}

func TestAddYears(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual(time.Date(2006, 11, 1, 10, 9, 8, 7, time.UTC),
		AddYears(time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), 2))

	tt.MustEqual(time.Date(2004, 9, 1, 10, 9, 8, 7, time.UTC),
		AddYears(time.Date(2006, 9, 1, 10, 9, 8, 7, time.UTC), -2))

	tt.MustEqual(time.Date(0, 1, 1, 10, 9, 8, 7, time.UTC),
		AddYears(time.Date(-2, 1, 1, 10, 9, 8, 7, time.UTC), 2))

	tt.MustEqual(time.Date(-1, 11, 1, 10, 9, 8, 7, time.UTC),
		AddYears(time.Date(-3, 11, 1, 10, 9, 8, 7, time.UTC), 2))
}

func TestPeriodMonths(t *testing.T) {
	tt := assert.WrapTB(t)
	tm := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(PeriodMonths(tm, 1), 564)
	tt.MustEqual(PeriodMonths(tm, 1), PeriodMonth(tm))
	tt.MustEqual(tm, PeriodMonthsTime(564, 1, nil))

	tm = time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.MustEqual(PeriodMonths(tm, 1), -24)
	tt.MustEqual(PeriodMonths(tm, 1), PeriodMonth(tm))
	tt.MustEqual(tm, PeriodMonthsTime(-24, 1, nil))
}

var benchPeriod int

func BenchmarkPeriodMonth(b *testing.B) {
	tm := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		benchPeriod = PeriodMonth(tm)
	}
}

func BenchmarkPeriodMonths1(b *testing.B) {
	tm := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		benchPeriod = PeriodMonths(tm, 1)
	}
}

func BenchmarkPeriodMonths2(b *testing.B) {
	tm := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		benchPeriod = PeriodMonths(tm, 2)
	}
}
