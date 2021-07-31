package httpagent

import (
	"fmt"
	"net/url"
	"time"
)

func ExampleSimpleCostAppliesToAllRequests() {
	policy := (&SimplePolicy{}).
		WithBudgets(Budget{Budget: 1, Interval: 60 * time.Second}).
		WithCosts(Cost{Cost: 1})

	limiter, _ := NewLimiter(WithHeadroom(0), WithPolicies(policy))

	u1, _ := url.Parse("http://first.example")
	ok, wait := limiter.Use("GET", u1, time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	u2, _ := url.Parse("http://second.example")
	ok, wait = limiter.Use("GET", u2, time.Date(2021, 1, 1, 12, 0, 59, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	u3, _ := url.Parse("http://third.example")
	ok, wait = limiter.Use("GET", u3, time.Date(2021, 1, 1, 12, 1, 0, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	// Output:
	// ok: true wait: 0s
	// ok: false wait: 1s
	// ok: true wait: 0s
}

func ExamplePolicyWithDefaultKey() {
	policy := (&SimplePolicy{}).
		WithName("example").
		WithBudgets(
			Budget{Budget: 1, Interval: 60 * time.Second},
		).
		WithCosts(
			Cost{Host: "www.fleeb.com", Cost: 1},
		)

	limiter, _ := NewLimiter(WithHeadroom(0), WithPolicies(policy))
	u, _ := url.Parse("http://www.fleeb.com")

	ok, wait := limiter.Use("GET", u, time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	ok, wait = limiter.Use("GET", u, time.Date(2021, 1, 1, 12, 0, 59, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	ok, wait = limiter.Use("GET", u, time.Date(2021, 1, 1, 12, 1, 0, 0, time.UTC))
	fmt.Println("ok:", ok, "wait:", wait)

	// Output:
	// ok: true wait: 0s
	// ok: false wait: 1s
	// ok: true wait: 0s
}
