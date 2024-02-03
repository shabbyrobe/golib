# Panic Merchant

Unhandled panics in goroutines will bypass your shutdown routines. This is
a simple utility to help centralise panic handling.

## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I **strongly** recommend you copy-paste pieces as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.
