# Sorttools

Contains copy-pastable tools for sorting.

Includes several copy-and-paste dupes of Go's sort library, hacked up to support
primitive types without interface boxing:

- SortUint64s
- SortInt64s
- SortFloat64s

They can be a lot faster than the stdlib's pseudo-generic approach, which
probably doesn't matter for almost every program, but it can really help when
it turns out that it does:

    BenchmarkSortUint64sThis-8              18840218                58.1 ns/op             0 B/op          0 allocs/op
    BenchmarkSortUint64sInterface-8          6989876               164 ns/op              32 B/op          1 allocs/op
    BenchmarkSortUint64sCallback-8           4555851               255 ns/op              64 B/op          2 allocs/op


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
