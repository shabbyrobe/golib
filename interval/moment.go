package interval

import "time"

func NewMoment(intvl Interval, tm time.Time) Moment {
	return Moment{Interval: intvl, Period: intvl.Period(tm)}
}

func (m Moment) Time(loc *time.Location) time.Time {
	return m.Interval.Time(m.Period, loc)
}

func (m Moment) Next() Moment {
	m.Period++
	return m
}

func (m Moment) Prev() Moment {
	m.Period--
	return m
}

func (m Moment) String() string {
	return m.Time(nil).String()
}
