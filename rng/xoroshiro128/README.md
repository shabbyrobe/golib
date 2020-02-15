xoroshiro128+ RNG
=================

Implements http://vigna.di.unimi.it/xorshift/xoroshiro128plus.c

Can be about twice as quick as Go's RNG, with arguably lower quality output. If
speed matters much more than quality, this is probably your RNG.

```
pkg: github.com/shabbyrobe/golib/rng/xoroshiro128
BenchmarkSource-8       623339676            1.82 ns/op
BenchmarkStdlib-8       325350958            3.65 ns/op
```
