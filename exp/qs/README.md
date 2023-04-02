# Experiment with query string parsing functions and generics

Is it possible to make a viable replacement for the abandoned gorilla/schema
that makes manually decoding a query string into a struct a bit less brutal
in Go?

It would be nicer if we could make it read a bit more left-to-right but that's
impossible with generics at the moment.

Example:

```go
type ID int

type Thingo struct {
    ID     ID
    Stuff  int
    Nilly  *int
    Yep    int64
    Floats []float64
    IP     *net.IP
}

func (t *Thingo) DecodeQueryValues(values qs.Values) error {
    q := qs.NewLoader(values)
    t.ID = qs.Val(qs.AnyInt[ID](q.First("id")))
    t.Stuff = qs.Val(qs.Int(q.First("stuff")))
    t.Nilly = qs.Ptr(qs.Int(q.First("nilly")))
    t.Yep = qs.Val(qs.Int64(q.First("yep")))
    t.Floats = qs.Val(qs.Float64s(q.Get("floats")))
    t.IP = qs.Val(qs.Text[net.IP](qs.First("ip")))
    return loader.Err()
}
```

## Expectation Management

Don't use this for anything.

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.
