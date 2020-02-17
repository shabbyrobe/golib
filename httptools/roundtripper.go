package httptools

import "net/http"

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(rq *http.Request) (rs *http.Response, err error) {
	return f(rq)
}
