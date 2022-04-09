# Unstructured data wrapper

Companion to reflect.Value, intended for complex deserialisation scenarios (most likely
configuration) where tight control of rich errors is desired and performance is not a
concern.


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.

This library will _never_ have a v1.0, so you should hard-pin your version if
you choose not to vendor.

