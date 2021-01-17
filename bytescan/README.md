Byte Scanner
============

Clone of `bufio.Scanner` that works on static byte slices. Much faster than
`bufio.NewScanner(bytes.NewReader(b))` for this case, with an identical API.

Compatible with all the default split functions, i.e. `bufio.ScanLines`,
`bufio.ScanWords`.

Code is lifted directly from the stdlib, so the copyright is with The Go Authors
and the license is permissive.


## Expectation Management

I have hacked this together quickly to solve a specific problem that doesn't
require high accuracy. I can't guarantee I've caught all the corner cases yet.

As with all modules in this collection, I _strongly_ recommend you vendor this
into your project, in an `internal/` directory, if you decide to use it as I
may change it without warning or regard to backward compatibility.


## Silly Benchmarks Game

Benchmarks run on my i7-8550U CPU @ 1.80GHz:

    BenchmarkByteScanSplitChunk-8             738702              1649 ns/op              16 B/op          1 allocs/op
    BenchmarkBufioScanSplitChunk-8            246861              4206 ns/op            4160 B/op          3 allocs/op
    BenchmarkByteScanLines-8                  426286              2946 ns/op               0 B/op          0 allocs/op
    BenchmarkBufioScanLines-8                 177150              5741 ns/op            4144 B/op          2 allocs/op

