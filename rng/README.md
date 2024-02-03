# RNG

Alternative `math/rand.Source` implementations.

Probably not even worth using these; Go's RNG is very fast these days.
Having said that, you can potentially double your performance if you're
happy to accept the weaknessess and tradeoffs present in these RNGs.

Also, the stdlib is getting a newer, better rand library as of 1.22,
so this will hopefully be even less necessary.


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
