# Unidecode

Unidecode implementation allowing operations on byte slices, in-place buffers
and strings, with full control over allocation.

Tables are built from the data in the Perl version, which is licensed under the
regrettable, FSF-rejected "Artistic License". This may be a derivative work as
a consequence, but if not, consider it MIT licensed.


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
