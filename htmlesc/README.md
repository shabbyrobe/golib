# HTML Unescape for byte slices

Copy-paste of just enough of the `html.Unescape` function from the standard library
to work on byte slices directly instead of strings.

This should hopefully be unnecessary if the changes that land in 1.22 work
as advertised: https://github.com/golang/go/issues/2205


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
