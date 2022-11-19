# Action loop boilerplate

Beyond a certain level of complexity, serial execution is much easier to reason
about than a rats-nest of mutexes and goroutines coordinating themselves in the
aether.

This library provides a starting point for serial execution of structs
representing function calls. It can be very useful when strapped directly to an
API to drive the state at the heart of a complex service. See example_test.go
for a fully worked example.

**This is intended to be copy-pasted into a project's `internal` folder and hit
with a rock until it resembles the shape you need.** It **should not** be
included in a project directly.

It'll be plenty fast enough for lots of workloads as-is, as long as you make
sure you don't block the loop, but there's lots of performance to be eked out
if you're happy to trade the simplicity of this solution for something more
complex.

```
cpu: AMD Ryzen 9 5950X 16-Core Processor
758.0 ns/op
176 B/op
3 allocs/op
```

## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
