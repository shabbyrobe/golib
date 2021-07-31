# Request rate limiter for net/http

## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

Having said that, there are some useful things in here, some of which are
reasonably well tested, and you may get some use out of them.

I recommend you copy-paste pieces as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.


## Basic explanation

A Policy can have one or more Budgets. A Budget allows a certain number
of points to be spent within a rolling Interval. A Cost indicates a
Budget to which a request applies, an optional Host and Path that will
be matched, and the Cost in points deducted from the budget.

If there are not enough points in the Budget, the Limiter will deny the
request and provide a wait time.


## Limitations

Note that this is a very naive solution to the problem that was thrown together
quickly; there is no queue, you get bumped by the limiter and you may get
bumped again. If the limiter is too contested, requests may be unlucky and get
bumped several times in a row.

This library was cobbled together to support simple scripts with either a single
loop or a small pool of goroutines with loops trying to play nice with a small
number of external APIs.


## Example

Simple policy, all requests cost the same, one budget:

```go
	policy := (&SimplePolicy{}).
		WithBudgets(Budget{Budget: 1, Interval: 60 * time.Second}).
		WithCosts(Cost{Cost: 1})

	limiter, err := NewLimiter(WithPolicies(policy))
```

Using with an http.Client and a 5 minute total wait limit:

```go
client := http.Client{
    Timeout: 60*time.Second,
	Transport: NewRoundTripper(limiter, nil, 5*time.Minute, nil),
}
rs, err := client.Get("http://example.com")
```

More complex policies:

```go
// Two costs sharing a single budget:
policy := (&rqlimit.SimplePolicy{}).
    WithName("example").
    WithBudgets(
        rqlimit.Budget{Budget: 10, Interval: 60 * time.Second},
    ).
    WithCosts(
        rqlimit.Cost{Host: "www.fleeb.com", Path: "/path", Cost: 1},
        rqlimit.Cost{Host: "www.derb.com", Path: "/other", Cost: 2},
    )
```

Using a fancy cost calculation function:

```go
// Fancy cost calculator:
policy := (&rqlimit.SimplePolicy{}).
    WithName("example").
    WithBudgets(
        rqlimit.Budget{Budget: 10, Interval: 60*time.Second},
    ).
    WithCosts(
		rqlimit.Cost{Host: "www.fleeb.com", Cost: 0,
			CostCalc: func(verb string, url *url.URL) int {
				if url.Query().Get("derb") != "" {
					return 5
				}
                return 1
			},
        },
    )
```

