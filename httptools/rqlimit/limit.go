package rqlimit

import (
	"net/url"
	"time"
)

type LimitKey string

type Budget struct {
	// Key is global; make sure you namespace your key names yourself.
	Key LimitKey

	// This Budget allows request costs totalling this much to occur inside
	// of Interval before the request is rejected with a delay:
	Budget int

	Interval time.Duration
}

type Cost struct {
	Host string
	Path string

	// Key is global; make sure you namespace your key names yourself.
	Key LimitKey

	// How much of the Limit's Budget claimed by this Cost:
	Cost int

	// CostCalc, if set, allows you to dynamically add additional cost to the
	// constant cost declared in the Cost field.
	CostCalc func(verb string, url *url.URL) int
}

func (lc Cost) TotalCost(verb string, url *url.URL) int {
	c := lc.Cost
	if lc.CostCalc != nil {
		c += lc.CostCalc(verb, url)
	}
	return c
}
