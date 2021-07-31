package httpagent

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func assertUse(t testing.TB, lm *Limiter, expected bool, wait time.Duration, verb string, url string, at time.Time) {
	t.Helper()
	ok, dur := lm.Use(verb, mustURL(url), at)
	if expected != ok {
		t.Fatal(expected, "!=", ok)
	}
	if wait != dur {
		t.Fatal("expected", wait, "!= actual", dur)
	}
}

func TestLimiterUseHostnameOnly(t *testing.T) {
	for _, tc := range []struct {
		costPath   string
		requestURL string
	}{
		{"/", "http://k3jw.com/"},
		{"/", "http://k3jw.com"},
		{"", "http://k3jw.com"},
		{"", "http://k3jw.com/"},
	} {
		t.Run("", func(t *testing.T) {
			lm, err := NewLimiter(WithHeadroom(0))
			if err != nil {
				t.Fatal(err)
			}

			p1 := &SimplePolicy{
				name: "yep",
				budgets: []Budget{
					{Key: "limit", Budget: 3, Interval: 1 * time.Second},
				},
				costs: []Cost{
					{Host: "k3jw.com", Path: tc.costPath, Key: "limit", Cost: 2},
				},
			}
			lm.AddPolicies(p1)

			tm := time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)

			assertUse(t, lm, true, 0*time.Second, "GET", tc.requestURL, tm)
			assertUse(t, lm, false, 1*time.Second, "GET", tc.requestURL, tm)
		})
	}
}

func TestLimiterMultiCost(t *testing.T) {
	lm, err := NewLimiter(WithHeadroom(0))
	if err != nil {
		t.Fatal(err)
	}

	p1 := &SimplePolicy{
		name: "yep",
		budgets: []Budget{
			{Key: "one", Budget: 3, Interval: 1 * time.Second},
			{Key: "one", Budget: 6, Interval: 10 * time.Second},
		},
		costs: []Cost{
			{Host: "google.com", Path: "foo/bar", Key: "one", Cost: 2},
			{Host: "google.com", Path: "foo/baz", Key: "one", Cost: 1},
		},
	}
	lm.AddPolicies(p1)

	tm := time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)

	// This uses up our 1-second budget completely (bar: 2, baz: 1):
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm)
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/baz", tm)

	// These should be delayed 1 second:
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/bar", tm)
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/baz", tm)

	// These should succeed, but will use up the last of our 10-second budget:
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm.Add(2*time.Second))
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/baz", tm.Add(2*time.Second))

	// These should be delayed until the 10-second budget can afford them (when the very first
	// requests expire):
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/bar", tm.Add(9*time.Second))
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/baz", tm.Add(9*time.Second))

	// There is now budget free in the 10-second:
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm.Add(10*time.Second))
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/baz", tm.Add(10*time.Second))

	// The 1-second budget is used up, but the 10-second budget still has to wait 2 seconds for
	// some free resources:
	assertUse(t, lm, false, 2*time.Second, "GET", "http://google.com/foo/bar", tm.Add(10*time.Second))
	assertUse(t, lm, false, 2*time.Second, "GET", "http://google.com/foo/baz", tm.Add(10*time.Second))

	// The 1-second budget is free, but the 10-second budget still has 1 second left:
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/bar", tm.Add(11*time.Second))
	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/baz", tm.Add(11*time.Second))
}

func TestLimiterMultiPath(t *testing.T) {
	lm, err := NewLimiter(WithHeadroom(0))
	if err != nil {
		t.Fatal(err)
	}

	p1 := &SimplePolicy{
		name: "yep",
		budgets: []Budget{
			{Key: "one", Budget: 10, Interval: 10 * time.Second},
			{Key: "two", Budget: 10, Interval: 10 * time.Second},
		},
		costs: []Cost{
			{Host: "google.com", Path: "foo/bar", Key: "one", Cost: 3},
			{Host: "google.com", Path: "foo/baz", Key: "two", Cost: 2},
		},
	}
	lm.AddPolicies(p1)

	tm := time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm)
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm.Add(1*time.Second))
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm.Add(2*time.Second))

	// Only the /foo/bar endpoint should be limited at this time:
	assertUse(t, lm, false, 7*time.Second, "GET", "http://google.com/foo/bar", tm.Add(3*time.Second))
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/baz", tm.Add(3*time.Second))

	assertUse(t, lm, false, 1*time.Second, "GET", "http://google.com/foo/bar", tm.Add(9*time.Second))
	assertUse(t, lm, true, 0*time.Second, "GET", "http://google.com/foo/bar", tm.Add(10*time.Second))
}

var BenchmarkCosts int

func BenchmarkLimiterFindCosts(b *testing.B) {
	for _, costCount := range []int{1, 10, 100, 1000} {
		b.Run("", func(b *testing.B) {
			lm, _ := NewLimiter(WithHeadroom(0))

			costs := make([]Cost, 0, costCount)
			for i := 0; i < costCount; i++ {
				costs = append(costs,
					Cost{Host: "google.com", Path: fmt.Sprintf("foo/%d", i), Key: "one", Cost: 2},
				)
			}

			p1 := &SimplePolicy{
				name:  "yep",
				costs: costs,
				budgets: []Budget{
					{Key: "one", Budget: 3, Interval: 1 * time.Second},
					{Key: "one", Budget: 6, Interval: 10 * time.Second},
					{Key: "two", Budget: 6, Interval: 10 * time.Second},
				},
			}
			lm.AddPolicies(p1)

			url := mustURL("http://google.com/foo/bar")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// This uses up our 1-second budget completely (bar: 2, baz: 1):
				cs := lm.FindCosts(url)
				BenchmarkCosts += len(cs)
			}
		})
	}
}

var BenchmarkWait time.Duration

func BenchmarkLimiterSimple(b *testing.B) {
	lm, _ := NewLimiter(WithHeadroom(0))

	costCount := 1000
	costs := make([]Cost, 0, costCount)
	for i := 0; i < costCount; i++ {
		costs = append(costs,
			Cost{Host: "google.com", Path: fmt.Sprintf("foo/%d", i), Key: "one", Cost: 2},
		)
	}

	p1 := &SimplePolicy{
		name:  "yep",
		costs: costs,
		budgets: []Budget{
			{Key: "one", Budget: 3, Interval: 1 * time.Second},
			{Key: "one", Budget: 6, Interval: 10 * time.Second},
			{Key: "two", Budget: 6, Interval: 10 * time.Second},
		},
	}
	lm.AddPolicies(p1)

	tm := time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)

	url := mustURL("http://google.com/foo/bar")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This uses up our 1-second budget completely (bar: 2, baz: 1):
		_, wait := lm.Use("GET", url, tm)
		BenchmarkWait += wait
	}
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
