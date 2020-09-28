package times

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestFirstDayOfWeek(t *testing.T) {
	iter := 20000

	start := time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC)

	for i := 0; i < iter; i++ {
		f := start.Add(time.Duration(i) * 24 * time.Hour)
		n := FirstMondayOfWeek(f)
		if n.Weekday() != time.Monday {
			t.Fatal(f.String())
		}
	}

	for i := 0; i < iter; i++ {
		f := start.Add(time.Duration(i) * 24 * time.Hour)
		f = f.Add(time.Duration(rand.Intn(60)) * time.Minute)
		f = f.Add(time.Duration(rand.Intn(24)) * time.Hour)
		n := FirstMondayOfWeek(f)
		if n.Weekday() != time.Monday {
			t.Fatal(f.String())
		}
	}
}

func TestTruncateMonths(t *testing.T) {
	// qty of "0" should infer "1":
	tm := TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 0)
	exp := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	// Negative qty should be treated like positive
	tm = TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), -2)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 2, 3, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(-1, 12, 31, 23, 59, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(-1, 11, 01, 23, 59, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(-1, 11, 01, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 11, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(-1, 10, 31, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(-1, 9, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	// 3 months:

	tm = TruncateMonths(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 2, 3, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 3, 3, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1970, 4, 1, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 4, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1969, 11, 01, 23, 59, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1969, 11, 01, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1969, 10, 31, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateMonths(time.Date(1969, 9, 30, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 7, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}
}

func TestTruncateWeeks(t *testing.T) {
	tm := TruncateWeeks(time.Date(1970, 1, 14, 0, 0, 0, 0, time.UTC), 2)
	exp := time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 1, 21, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 1, 28, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(2017, 1, 28, 12, 30, 0, 0, time.UTC), 2)
	exp = time.Date(2017, 1, 23, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 16, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 14, 0, 0, 0, 0, time.UTC), 2)
	exp = time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	// 3 weeks
	tm = TruncateWeeks(time.Date(1970, 1, 21, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 1, 28, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 2, 5, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 1, 19, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1970, 2, 12, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1970, 2, 9, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 16, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

	tm = TruncateWeeks(time.Date(1969, 12, 7, 0, 0, 0, 0, time.UTC), 3)
	exp = time.Date(1969, 11, 17, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(exp) {
		t.Fatal(tm, "!=", exp)
	}

}

func TestAddMonths(t *testing.T) {
	for idx, tc := range []struct {
		tm     time.Time
		months int
		result time.Time
	}{
		{time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(2005, 1, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), -2,
			time.Date(2004, 9, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(-1, 11, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(0, 1, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(-2, 11, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(-1, 1, 1, 10, 9, 8, 7, time.UTC)},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			v := AddMonths(tc.tm, tc.months)
			if !v.Equal(tc.result) {
				t.Fatal(v, "!=", tc.result)
			}
		})
	}
}

func TestAddYears(t *testing.T) {
	for idx, tc := range []struct {
		tm     time.Time
		years  int
		result time.Time
	}{
		{time.Date(2004, 11, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(2006, 11, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(2006, 9, 1, 10, 9, 8, 7, time.UTC), -2,
			time.Date(2004, 9, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(-2, 1, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(0, 1, 1, 10, 9, 8, 7, time.UTC)},

		{time.Date(-3, 11, 1, 10, 9, 8, 7, time.UTC), 2,
			time.Date(-1, 11, 1, 10, 9, 8, 7, time.UTC)},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			v := AddYears(tc.tm, tc.years)
			if !v.Equal(tc.result) {
				t.Fatal(v, "!=", tc.result)
			}
		})
	}
}

func TestPeriodMonths(t *testing.T) {
	tm := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	if PeriodMonths(tm, 1) != 564 {
		t.Fatal()
	}
	if PeriodMonths(tm, 1) != PeriodMonth(tm) {
		t.Fatal()
	}
	if !tm.Equal(PeriodMonthsTime(564, 1, nil)) {
		t.Fatal()
	}

	tm = time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)
	if PeriodMonths(tm, 1) != -24 {
		t.Fatal()
	}
	if PeriodMonths(tm, 1) != PeriodMonth(tm) {
		t.Fatal()
	}
	if !tm.Equal(PeriodMonthsTime(-24, 1, nil)) {
		t.Fatal()
	}
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
