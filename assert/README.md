# Simple Go Assertions

**NOTE 20231228: This is deprecated, I don't use it anymore. I just use plain
ol' `t.Fatal` and friends, and github.com/google/go-cmp/cmp for anything fancy.**


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
