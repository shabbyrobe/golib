package errtools

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

func mustPattern(t *testing.T, pattern string, in string) {
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
	in := fmt.Errorf("yep")
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil)
		ec.Do(in)
		return
	}()
	if !reflect.DeepEqual(ec, result) {
		t.Fatal()
	}
	mustPattern(t, `error at .*_test\.go.* #1 - yep`, ec.Error())
}

func TestCollectorSetOK(t *testing.T) {
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil)
		return
	}()
	if result != nil {
		t.Fatal()
	}
}

func TestCollectorSetMultiple(t *testing.T) {
	in := fmt.Errorf("yep")
	ec := &Collector{}
	result := func() (err error) {
		defer ec.Set(&err)
		ec.Do(nil, nil, in)
		return
	}()
	if !reflect.DeepEqual(ec, result) {
		t.Fatal()
	}
	mustPattern(t, `error at .*_test\.go.* #3 - yep`, ec.Error())
}

func TestCollectorPanic(t *testing.T) {
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
	if !reflect.DeepEqual(ec, result) {
		t.Fatal()
	}
	mustPattern(t, `error at .*_test\.go.* #3 - yep`, ec.Error())
}
