package times

import (
	"errors"
	"fmt"
	"time"
)

var atoiError = errors.New("date: invalid number")

func EpochDate() Date {
	return Date{1970, 0, 0}
}

// Date without time or timezone
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func New(y int, m time.Month, d int) Date {
	return Date{y, m, d}
}

func DateFromTime(t time.Time) (d Date) {
	d.Year, d.Month, d.Day = t.Date()
	return d
}

func MustParseDate(s string) (date Date) {
	d, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func ParseDate(s string) (date Date, err error) {
	const sy, sm, sd = 0, 1, 2
	var state int
	var y, m, d int
	var yn, mn, dn int

	for idx := 0; idx < len(s); idx++ {
		c := s[idx]
		if c == '-' {
			state++
			continue
		}

		n := int(c) - '0'
		if n < 0 || n > 9 {
			return date, atoiError
		}
		switch state {
		case sy:
			y = y*10 + n
			yn++
		case sm:
			m = m*10 + n
			mn++
		case sd:
			d = d*10 + n
			dn++
		default:
			return date, fmt.Errorf("date: invalid date format")
		}
	}

	date = Date{Year: y, Month: time.Month(m), Day: d}
	if yn != 4 || mn != 2 || dn != 2 {
		return Date{}, fmt.Errorf("date: invalid date")
	}
	if date.IsZero() {
		return date, nil
	}
	if !date.IsValid() {
		return Date{}, fmt.Errorf("date: invalid date")
	}

	return date, nil
}

func (d Date) Today() Date {
	now := time.Now()
	return Date{Year: now.Year(), Month: now.Month(), Day: now.Day()}
}

func (d Date) DaysSince(s Date) (days int) {
	// Clues taken from 'civil' package:
	deltaUnix := d.In(time.UTC).Unix() - s.In(time.UTC).Unix()
	return int(deltaUnix / 86400)
}

func (d Date) IsZero() bool {
	return d == Date{}
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) IsValid() bool {
	if d.Year < 0 || d.Year > 9999 || d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 {
		return false
	}
	t := d.In(time.UTC)
	return t.Year() == d.Year && t.Month() == d.Month && t.Day() == d.Day
}

func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

func (d Date) Before(other Date) bool {
	if d.Year != other.Year {
		return d.Year < other.Year
	}
	if d.Month != other.Month {
		return d.Month < other.Month
	}
	return d.Day < other.Day
}

func (d Date) After(other Date) bool {
	return other.Before(d)
}

func (d Date) Equal(other Date) bool {
	return d.Year == other.Year && d.Month == other.Month && d.Day == other.Day
}

func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, nil
	}
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	var err error
	if len(data) == 0 {
		*d = Date{}
		return nil
	}
	*d, err = ParseDate(string(data))
	return err
}

type DateFlag struct {
	Date
	IsSet bool
}

func (df DateFlag) String() string {
	if df.IsSet {
		return df.Date.String()
	}
	return ""
}

func (df *DateFlag) Set(s string) (err error) {
	df.Date, err = ParseDate(s)
	return err
}
