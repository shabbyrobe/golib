Go Intervals
============

`interval` is a simple way of storing human-centric intervals of time in a
compact representation.

It is split into four key concepts: `Span`, `Qty`, `Interval`, and `Period`.

`Span` is the amount of time that has passed. `Qty` is the number of `Spans`.
`Interval` is the combination of `Span` and `Qty`. `Period` is the index of
the interval relative to a fixed real time (Unix Epoch).

For example:

- `Span` == Minute
- `Qty` == Five of them
- `Interval` == `Qty(Five of them) * Span(Minute)`
- `Period` == The 5,121st `Interval` since the Unix Epoch.
- `Time` == Do some jiggery pokery with an `Interval` and a `Period`.

And in code:

```go
span := interval.Minute
qty := interval.Qty(5)
intvl := interval.New(qty, span)
period := interval.Period(5121)
periodTime := intvl.Time(period, time.UTC)
```


Intervals can be parsed as a string. Parsing is reasonably tolerant:

```go
intvl := interval.MustParse("1s")
intvl := interval.MustParse("10sec")
intvl := interval.MustParse("1min")
intvl := interval.MustParse("1hr")
intvl := interval.MustParse("1 hour")
intvl := interval.MustParse("1 day")
intvl := interval.MustParse("1 week")
intvl := interval.MustParse("1mo")
intvl := interval.MustParse("1 month")

// PANIC! ambiguous.
nope := interval.MustParse("1m") 
```
