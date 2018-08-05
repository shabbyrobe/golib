package incrementer

import (
	"errors"
)

const growSize = 2

func Must(v string) *Inc {
	inc, err := New(v)
	if err != nil {
		panic(err)
	}
	return inc
}

func New(v string) (*Inc, error) {
	inc := &Inc{}
	if err := inc.Set(v); err != nil {
		return nil, err
	}
	return inc, nil
}

type Inc struct {
	buf []byte
	len int
	cap int
}

func (inc *Inc) Set(v string) error {
	trim := 0
	trimmed := false
	for _, i := range v {
		if i != '0' {
			trimmed = true
		} else if !trimmed {
			trim++
		}
		if i < '0' || i > '9' {
			return errors.New("input must not contain non-numeric characters")
		}
	}
	inc.buf = []byte(v[trim:])
	inc.len = len(inc.buf)
	inc.cap = inc.len
	return nil
}

func (inc *Inc) Current() string {
	if inc.len == 0 {
		return "0"
	}
	return string(inc.buf[:inc.len])
}

func (inc *Inc) Next() string {
	carried := inc.len - 1

	// find the digit that will be incremented. all digits to the right are 9s
	// and will be zeroed.
	for ; carried >= 0 && inc.buf[carried] == '9'; carried-- {
		inc.buf[carried] = '0'
	}

	if carried < 0 {
		// Everything's a 9!
		if inc.len == inc.cap {
			inc.cap += growSize
			inc.buf = make([]byte, inc.cap)
			for i := 0; i < inc.cap; i++ {
				inc.buf[i] = '0'
			}
		}
		inc.buf[0] = '1'
		inc.len++

	} else {
		inc.buf[carried]++
	}

	return string(inc.buf[:inc.len])
}
