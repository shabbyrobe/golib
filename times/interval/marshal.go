package interval

type IntervalEncoded Interval

func (i IntervalEncoded) Interval() Interval {
	return Interval(i)
}

func (i IntervalEncoded) MarshalText() (text []byte, err error) {
	return []byte(Interval(i).String()), nil
}

func (i *IntervalEncoded) UnmarshalText(text []byte) (err error) {
	ip, err := Parse(string(text))
	*i = IntervalEncoded(ip)
	return err
}
