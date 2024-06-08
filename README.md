Go Toolbelt
===========

**NOTE**: This repo is moving to https://git.sr.ht/~shabbyrobe/golib. The new import path
base is go.shabbyrobe.org/golib, but you shouldn't be importing any of this directly
anyway, you should be manually vendoring it.

Packages and files added to this Github repo from 2020 onwards will be progressively
removed (this is when the "Expectation Management" section was introduced in largely its
current format). In the future, all packages will be removed, probably after 2025 or so.
The repo will be preserved with a single README, though probably with the history removed
(as it will still be available at sourcehut).

The Go module proxy should protect you if you still somehow depend on the removed code.

---

This is a collection of Go modules I use to augment the Go standard library in
my own personal projects. Please read the section titled "Expectation
Management" before using any of this code.

Some parts of this repo have not been used in a very long time and it could
stand a bit of spring cleaning.

All code is available under the MIT license unless otherwise stated in the
particular file or package.


## Expectation Management

This is a personal grab-bag of utility code that I add to in a very ad-hoc
fashion. *No API stability guarantees are made*, the code is *not guaranteed to
work*, and anything may be changed, renamed or removed at any time as I see fit.

Having said that, there are some useful things in here, some of which are
reasonably well tested, and you may get some use out of them.

If you wish to use any of this, I strongly recommend you copy-paste pieces
as-needed (including tests and license/attribution) into the `internal/` folder
of your projects rather than reference these modules directly. Stability of this
repo is in no way guaranteed.
