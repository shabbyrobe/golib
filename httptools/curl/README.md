Curl HTTP Transport for Go's http.Client
========================================

Implements an http.RoundTripper for Go's http.Client that uses your local
curl binary to execute the request.


## Usage

```go
rt := curl.Transport{}
hc := http.Client{Transport: &rt}
rs, _ := hc.Get("https://en.wikipedia.org/")
```


## Why!?

Mostly to see if it could be done. No real reason beyond that.


## Expectation Management

This is an experiment. *No API stability guarantees are made* and the code is
*not guaranteed to work*.

I recommend you copy-paste pieces you want to use as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.

