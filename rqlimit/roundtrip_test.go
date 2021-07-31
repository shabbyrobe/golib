package rqlimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type serveResult struct {
	url string
	at  time.Time
}

type testServer struct {
	*httptest.Server
	delay   time.Duration
	results chan serveResult
}

func (t *testServer) Run() *testServer {
	t.results = make(chan serveResult, 50)
	t.Server = httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if t.delay > 0 {
				time.Sleep(t.delay)
			}
			t.results <- serveResult{r.URL.String(), time.Now()}
		}))

	return t
}

func TestRoundTripperWaitLimit(t *testing.T) {
	ts := (&testServer{}).Run()
	defer ts.Close()
	client := ts.Client()

	policy := (&SimplePolicy{}).
		WithBudgets(Budget{Budget: 1, Interval: 60 * time.Second}).
		WithCosts(Cost{Cost: 1})

	limiter, _ := NewLimiter(WithHeadroom(0), WithPolicies(policy))
	client.Transport = NewRoundTripper(limiter, nil, 10*time.Millisecond, nil)
	if _, err := client.Get(ts.URL); err != nil {
		t.Fatal(err)
	}

	// Second request should exceed limit:
	if _, err := (client.Get(ts.URL)); err == nil {
		t.Fatal("expected error")
	}
}
