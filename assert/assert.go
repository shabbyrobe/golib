package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func WrapTB(tb testing.TB) T { tb.Helper(); return T{TB: tb} }

type T struct{ testing.TB }

// frameDepth is the number of frames to strip off the callstack when reporting the line
// where an error occurred.
const frameDepth = 2

func CompareMsg(exp, act interface{}) string {
	return fmt.Sprintf("\nexp: %+v\ngot: %+v", exp, act)
}

func CompareMsgf(exp, act interface{}, msg string, args ...interface{}) string {
	msg = fmt.Sprintf(msg, args...)
	return fmt.Sprintf("%v%v", msg, CompareMsg(exp, act))
}

func IsFloatNear(epsilon, expected, actual float64) bool {
	diff := expected - actual
	return diff == 0 || (diff < 0 && diff > -epsilon) || (diff > 0 && diff < epsilon)
}

func (tb T) MustFloatNear(epsilon float64, expected float64, actual float64, v ...interface{}) {
	tb.Helper()
	_ = tb.floatNear(true, epsilon, expected, actual, v...)
}

func (tb T) FloatNear(epsilon float64, expected float64, actual float64, v ...interface{}) bool {
	tb.Helper()
	return tb.floatNear(false, epsilon, expected, actual, v...)
}

func (tb T) floatNear(fatal bool, epsilon float64, expected float64, actual float64, v ...interface{}) bool {
	tb.Helper()
	near := IsFloatNear(epsilon, expected, actual)
	if !near {
		_, file, line, _ := runtime.Caller(frameDepth)
		msg := ""
		if len(v) > 0 {
			msg, v = v[0].(string), v[1:]
		}
		v = append([]interface{}{expected, actual, epsilon, filepath.Base(file), line}, v...)
		msg = fmt.Sprintf("\nfloat abs(%f - %f) > %f at %s:%d\n"+msg, v...)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
	}
	return near
}

// MustAssert immediately fails the test if the condition is false.
func (tb T) MustAssert(condition bool, v ...interface{}) {
	tb.Helper()
	_ = tb.assert(true, condition, v...)
}

// Assert fails the test if the condition is false.
func (tb T) Assert(condition bool, v ...interface{}) bool {
	tb.Helper()
	return tb.assert(false, condition, v...)
}

func (tb T) assert(fatal bool, condition bool, v ...interface{}) bool {
	tb.Helper()
	if !condition {
		_, file, line, _ := runtime.Caller(frameDepth)
		msg := ""
		if len(v) > 0 {
			msgx := v[0]
			v = v[1:]
			if msgx == nil {
				msg = "<nil>"
			} else if err, ok := msgx.(error); ok {
				msg = err.Error()
			} else {
				msg = msgx.(string)
			}
		}
		v = append([]interface{}{filepath.Base(file), line}, v...)
		msg = fmt.Sprintf("\nassertion failed at %s:%d\n"+msg, v...)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
	}
	return condition
}

// MustOKAll errors and terminates the test at the first error found in the arguments.
// It allows multiple return value functions to be passed in directly.
func (tb T) MustOKAll(errs ...interface{}) {
	tb.Helper()
	_ = tb.okAll(true, errs...)
}

// OKAll errors the test at the first error found in the arguments, but continues
// running the test. It allows multiple return value functions to be passed in
// directly.
func (tb T) OKAll(errs ...interface{}) bool {
	tb.Helper()
	return tb.okAll(false, errs...)
}

func (tb T) okAll(fatal bool, errs ...interface{}) bool {
	tb.Helper()
	for _, err := range errs {
		if _, ok := err.(*testing.T); ok {
			panic("unexpected testing.T in call to OK()")
		} else if _, ok := err.(T); ok {
			panic("unexpected testtools.T in call to OK()")
		}
		if err, ok := err.(error); ok && err != nil {
			if !tb.ok(fatal, err) {
				return false
			}
		}
	}
	return true
}

func (tb T) MustOK(err error) {
	tb.Helper()
	_ = tb.ok(true, err)
}

func (tb T) OK(err error) bool {
	tb.Helper()
	return tb.ok(true, err)
}

func (tb T) ok(fatal bool, err error) bool {
	tb.Helper()
	if err == nil {
		return true
	}
	_, file, line, _ := runtime.Caller(frameDepth)
	msg := fmt.Sprintf("\nunexpected error at %s:%d\n%s", filepath.Base(file), line, err.Error())
	if fatal {
		tb.Fatal(msg)
	} else {
		tb.Error(msg)
	}
	return false
}

// MustExact immediately fails the test if exp is not equal to act.
func (tb T) MustExact(exp, act interface{}, v ...interface{}) {
	tb.Helper()
	_ = tb.exact(true, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) Exact(exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	return tb.exact(false, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) exact(fatal bool, exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	if exp != act {
		extra := ""
		if len(v) > 0 {
			extra = fmt.Sprintf(" - "+v[0].(string), v[1:]...)
		}

		_, file, line, _ := runtime.Caller(frameDepth)
		msg := CompareMsgf(exp, act, "\nexact failed at %s:%d%s", filepath.Base(file), line, extra)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
		return false
	}
	return true
}

// MustEqual immediately fails the test if exp is not equal to act based on
// reflect.DeepEqual()
func (tb T) MustEqual(exp, act interface{}, v ...interface{}) {
	tb.Helper()
	_ = tb.equals(true, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) Equals(exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	return tb.equals(false, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) equals(fatal bool, exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		extra := ""
		if len(v) > 0 {
			extra = fmt.Sprintf(" - "+v[0].(string), v[1:]...)
		}

		_, file, line, _ := runtime.Caller(frameDepth)
		msg := CompareMsgf(exp, act, "\nequal failed at %s:%d%s", filepath.Base(file), line, extra)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
		return false
	}
	return true
}
