package errtools

import (
	"fmt"
	"runtime"
)

/*
Collector allows you to defer raising or accumulating an error
until after a series of procedural calls.

Collector it is intended to help cut down on boilerplate like this, in cases
where the object you are using returns errors but it is safe to continue using
the object if one call fails.

	doer := &Doer{}
	if err := doer.Do(1); err != nil {
		return err
	}
	if err := doer.Do(2); err != nil {
		return err
	}
	if err := doer.Do(3); err != nil {
		return err
	}

Collector allows you to assume that it's ok to keep doing things until the end
of a controlled block even if the first one fails, and then return the first
error that occurred. In complex procedures, Collector is far more succinct and
mirrors an idiom used internally in the library, which was itself cribbed from
the stdlib's xml package (see cachedWriteError).

To use a collector:

	func pants(doer *Doer) error {
		ec := &errtools.Collector{}
		ec.Do(
			doer.Do(1),
			doer.Do(2),
			doer.Do(3),
		)
		return ec.Cause()
	}

To use a collector with a named return:

	func pants(doer *Doer) (err error) {
		ec := &errtools.Collector{}
		defer ec.Set(&err)
		ec.Do(
			doer.Do(1),
			doer.Do(2),
			doer.Do(3),
		)
		return
	}

If you want to panic instead, follow the named return example, substituting
`defer ec.Set(&err)` with `defer ec.Panic()`

It is entirely the responsibility of the library's user to remember to call
either `ec.Set()`, `ec.Panic()` or `ec.Cause()`. If you don't, you'll be
swallowing errors.
*/
type Collector struct {
	File  string
	Line  int
	Index int
	Err   error
}

// Cause returns the underlying error.
func (e *Collector) Cause() error {
	return e.Err
}

// Error implements the error interface.
func (e *Collector) Error() string {
	return fmt.Sprintf("error at %s:%d #%d - %v", e.File, e.Line, e.Index, e.Err)
}

// Panic causes the collector to panic if any error has been collected.
//
// This should be called in a defer:
//
//	func pants() {
//		ec := &errtools.Collector{}
//		defer ec.Panic()
//		ec.Do(fmt.Errorf("this will panic at the end"))
//		fmt.Printf("This will print")
//	}
//
func (e *Collector) Panic() {
	if e.Err != nil {
		panic(e)
	}
}

// Set assigns the collector's internal error to an external error variable.
//
// This should be called in a defer with a named return to allow an error
// to be easily returned if one is collected:
//
//	func pants() (err error) {
//		ec := &xmlwriter.Collector{}
//		defer ec.Set(&err)
//		ec.Do(fmt.Errorf("this error will be returned by the pants function"))
//		fmt.Printf("This will print")
//	}
//
func (e *Collector) Set(err *error) {
	if e.Err != nil {
		*err = e
	}
}

// Do collects the first error in a list of errors and holds on to it.
//
// If you pass the result of multiple functions to Do, they will not be
// short circuited on failure - the first error is retained by the collector
// and the rest are discarded. It is only intended to be used when you know
// that subsequent calls after the first error are safe to make.
//
func (e *Collector) Do(errs ...error) {
	for i, err := range errs {
		if err != nil {
			_, file, line, _ := runtime.Caller(1)
			e.Err = err
			e.Index = i + 1
			e.File = file
			e.Line = line
			return
		}
	}
}

// Must collects the first error in a list of errors and panics with it.
func (e *Collector) Must(errs ...error) {
	for i, err := range errs {
		if err != nil {
			_, file, line, _ := runtime.Caller(1)
			e.Err = err
			e.Index = i + 1
			e.File = file
			e.Line = line
			panic(e)
		}
	}
}
