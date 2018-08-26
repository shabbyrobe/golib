/*
Package errgroup is copy/pasted from https://godoc.org/golang.org/x/sync/errgroup.

It uses the stdlib context instead of x/context because my binaries are big enough
without two copies of the same dependency.

It seems AppEngine is the reason why this is "necessary" for now:
https://github.com/golang/go/issues/19781
*/
package errgroup
