package httpagent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type RoundTripper struct {
	ctx       context.Context
	transport http.RoundTripper
	limiter   *Limiter
	waitLimit time.Duration
	log       Logger
	lock      sync.Mutex
}

var _ http.RoundTripper = &RoundTripper{}

const DefaultRoundTripWaitLimit time.Duration = 300 * time.Second

func NewRoundTripper(
	l *Limiter,
	t http.RoundTripper,
	waitLimit time.Duration,
	logger Logger,
) *RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	if logger == nil {
		logger = &nilLogger{}
	}
	if waitLimit == 0 {
		waitLimit = DefaultRoundTripWaitLimit
	}
	lt := &RoundTripper{
		transport: t,
		limiter:   l,
		log:       logger,
		waitLimit: waitLimit,
	}
	return lt
}

func (lt *RoundTripper) useLimit(verb string, url *url.URL, at time.Time) (ok bool, wait time.Duration) {
	lt.lock.Lock()
	defer lt.lock.Unlock()

	return lt.limiter.Use(verb, url, at)
}

func (lt *RoundTripper) RoundTrip(rq *http.Request) (*http.Response, error) {
	var start = time.Now()
	var lastAttempt = false

	for {
		ok, wait := lt.useLimit(rq.Method, rq.URL, time.Now())
		if ok {
			break
		}
		if lastAttempt {
			return nil, fmt.Errorf("limiter: roundtrip for url %q took longer than %s", rq.URL, lt.waitLimit)
		}

		if lt.waitLimit > 0 {
			maxWait := lt.waitLimit - time.Since(start)
			if maxWait <= wait {
				lastAttempt = true
				wait = maxWait
			}
		}

		lt.log.Printf("roundtripper delayed %s: %q", wait, rq.URL)
		select {
		case <-time.After(wait):
		case <-rq.Context().Done():
			return nil, rq.Context().Err()
		}
	}

	return lt.transport.RoundTrip(rq)
}
