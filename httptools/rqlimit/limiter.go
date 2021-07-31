package rqlimit

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const DefaultHeadroom float64 = 0.1

type Limiter struct {
	policies []Policy
	limits   map[LimitKey]map[time.Duration]*limitState

	headroom    float64
	headroomSet bool
	minWait     time.Duration

	// This is a cheap and nasty way to get this up and running. Need to
	// implement a simple tree once this exceeds about 100 routes, but this
	// saves a lot of time for now.
	tree *http.ServeMux
}

type LimiterOption func(l *Limiter) error

func WithMinWait(d time.Duration) LimiterOption {
	return func(l *Limiter) error { l.minWait = d; return nil }
}

// Headroom sets the percentage (0.0, 1.0) of allocation that will be left
// unused before the requests are delayed.
//
// A headroom of 0.1 will leave 10% of all budgets unused.
//
func WithHeadroom(h float64) LimiterOption {
	return func(l *Limiter) error { return l.SetHeadroom(h) }
}

func WithPolicies(ps ...Policy) LimiterOption {
	return func(l *Limiter) error { return l.AddPolicies(ps...) }
}

func NewLimiter(opts ...LimiterOption) (*Limiter, error) {
	tree := http.NewServeMux()
	lm := &Limiter{
		tree:     tree,
		headroom: DefaultHeadroom,
		limits:   map[LimitKey]map[time.Duration]*limitState{},
	}
	for _, o := range opts {
		if err := o(lm); err != nil {
			return nil, err
		}
	}
	if lm.minWait <= 0 {
		lm.minWait = 100 * time.Millisecond
	}
	return lm, nil
}

// SetHeadroom sets the percentage (0.0-1.0) of allocation that will be left
// unused before the requests are delayed.
//
// SetHeadroom(0.1) will leave 10% of all budgets unused.
//
func (l *Limiter) SetHeadroom(headroom float64) error {
	if headroom < 0 || headroom >= 1 {
		return fmt.Errorf("headroom must be 0.0 <= h < 1.0")
	}
	l.headroom = headroom
	l.headroomSet = true
	return nil
}

func (l *Limiter) AddPolicies(lps ...Policy) error {
	index := make(map[string]*policyWrapper)

	for _, lp := range lps {
		l.policies = append(l.policies, lp)

		for _, limit := range lp.Budgets() {
			if _, ok := l.limits[limit.Key]; !ok {
				l.limits[limit.Key] = map[time.Duration]*limitState{}
			}
			budget := limit.Budget
			if l.headroom > 0 {
				headroom := float64(budget) * l.headroom
				budget = int(float64(budget) - headroom)
			}

			if state, ok := l.limits[limit.Key][limit.Interval]; !ok {
				l.limits[limit.Key][limit.Interval] = &limitState{
					interval: limit.Interval,
					free:     budget,
				}
			} else {
				if budget < state.free {
					state.free = budget
				}
			}
		}

		for idx, cost := range lp.Costs() {
			if _, ok := l.limits[cost.Key]; !ok {
				return fmt.Errorf("unknown limit key %q in limit cost at index %d for policy %s", cost.Key, idx, lp.Name())
			}

			ptn := path.Join(cost.Host, cost.Path)

			if cost.Host != "" && (cost.Path == "" || cost.Path == "/") {
				// This hack works around an inconsistency in http.ServeMux. Once this
				// is rewritten to use a hand-rolled tree, this can go.
				ptn += "/"
			}

			if cost.Cost < 0 {
				return fmt.Errorf("httplimiter: cost must be >= 0, found %d", cost.Cost)
			} else if cost.Cost > 0 {
				for _, limit := range lp.Budgets() {
					budget := l.limits[limit.Key][limit.Interval] // must account for previous headroom calculation
					if limit.Key == cost.Key && cost.Cost > budget.free {
						return fmt.Errorf("httplimiter: cost %d higher than headroom adjusted budget %d for key %s; will never succeed", cost.Cost, budget.free, cost.Key)
					}
				}
			}

			if pw := index[ptn]; pw != nil {
				pw.costs = append(pw.costs, cost)

			} else {
				wrapped := &policyWrapper{costs: []Cost{cost}}

				// Workaround for Mux hack; empty path should be supported:
				if ptn == "" {
					ptn = "/"
				}

				l.tree.Handle(ptn, wrapped)
				index[ptn] = wrapped
			}
		}
	}

	return nil
}

func (l *Limiter) FindCosts(url *url.URL) []Cost {
	if url.Path == "" {
		// This hack works around an inconsistency in http.ServeMux. Once this
		// is rewritten to use a hand-rolled tree, this can go.
		url.Path = "/"
	}

	hr := http.Request{Host: url.Host, URL: url, Method: "GET"}
	handler, _ := l.tree.Handler(&hr)
	if pw, ok := handler.(*policyWrapper); ok {
		return pw.costs
	}
	return nil
}

func (l *Limiter) Use(verb string, url *url.URL, at time.Time) (ok bool, wait time.Duration) {
	var costs = l.FindCosts(url)
	if len(costs) == 0 {
		return true, 0
	}

	var allReady time.Time
	var budgetsAvail, budgetsTotal int

	for _, cost := range costs {
		currentCost := cost.TotalCost(verb, url)

		states := l.limits[cost.Key]
		for interval, state := range states {
			state.update(at)
			budgetsTotal++

			if state.free < currentCost {
				var saved int
				var ready time.Time

				for _, use := range state.uses {
					saved += use.cost
					if saved >= currentCost {
						ready = use.at
						break
					}
				}

				if !ready.IsZero() {
					ready = ready.Add(interval)
				} else {
					ready = at.Add(interval)
				}

				if ready.After(allReady) {
					allReady = ready
				}

			} else {
				budgetsAvail++
			}
		}
	}

	if budgetsTotal != budgetsAvail {
		wait := allReady.Sub(at)
		if wait < l.minWait {
			wait = l.minWait
		}
		return false, wait
	}

	for _, cost := range costs {
		currentCost := cost.TotalCost(verb, url)

		states := l.limits[cost.Key]
		for _, state := range states {
			state.uses = append(state.uses, limitUse{at: at, cost: currentCost})
			state.free -= cost.Cost
		}
	}

	return true, 0
}

type limitLookup struct {
	key      LimitKey
	interval time.Duration
}

type limitUse struct {
	at   time.Time
	cost int
}

type limitState struct {
	free     int
	interval time.Duration
	uses     []limitUse
}

func (ls *limitState) update(at time.Time) {
	chop := 0
	exp := at.Add(-ls.interval)
	for _, u := range ls.uses {
		if !exp.After(u.at) && !exp.Equal(u.at) {
			break
		}
		ls.free += u.cost
		chop++
	}

	if chop > 0 {
		ls.uses = ls.uses[chop:]
	}

	// we are constantly trimming from the front of the slice, which is
	// basically a memory leak. this re-creates the slice if the cap is too
	// much bigger than the len.
	//
	// FIXME: this was the 2 minute solution to this problem, probably worth
	// doing the 2 hour version at some point.
	if cap(ls.uses) > len(ls.uses)*10 {
		u := make([]limitUse, len(ls.uses))
		copy(u, ls.uses)
		ls.uses = u
	}
}

type policyWrapper struct {
	costs []Cost
}

func (pw *policyWrapper) ServeHTTP(http.ResponseWriter, *http.Request) { return }

func hostPath(host string) []string {
	hp := strings.Split(host, ".")
	for i, j := 0, len(hp)-1; i < j; i, j = i+1, j-1 {
		hp[i], hp[j] = hp[j], hp[i]
	}
	return hp
}
