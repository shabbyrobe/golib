package errtools

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func mustPattern(t assert.T, pattern string, in string) {
	t.Helper()
	ptn, _ := regexp.Compile(pattern)
	if !ptn.MatchString(in) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\tptn: %#v\n\n\tgot: %#v\033[39m\n\n",
			filepath.Base(file), line, pattern, in)
		t.FailNow()
	}
}

func TestCollectorSet(t *testing.T) {
	tt := assert.WrapTB(t)
	in := fmt.Errorf("yep")
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil)
		ec.Do(in)
		return
	}()
	tt.Equals(ec, result)
	mustPattern(tt, `error at .*_test\.go.* #1 - yep`, ec.Error())
}

func TestCollectorSetOK(t *testing.T) {
	tt := assert.WrapTB(t)
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil)
		return
	}()
	tt.Equals(nil, result)
}

func TestCollectorSetMultiple(t *testing.T) {
	tt := assert.WrapTB(t)
	in := fmt.Errorf("yep")
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil, nil, in)
		return
	}()
	tt.Equals(ec, result)
	mustPattern(tt, `error at .*_test\.go.* #3 - yep`, ec.Error())
}

func TestCollectorPanic(t *testing.T) {
	tt := assert.WrapTB(t)
	in := fmt.Errorf("yep")
	ec := &Collector{}
	result := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		func() {
			defer ec.Panic()
			ec.Do(nil, nil, in)
			return
		}()
		return
	}()
	tt.Equals(ec, result)
	mustPattern(tt, `error at .*_test\.go.* #3 - yep`, ec.Error())
}
