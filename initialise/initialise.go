// Simple copy-pasta for initialising structs in a manual recursive-descent

package initialise

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func IsFailed(err error) bool {
	var failed *Failed
	return errors.As(err, &failed)
}

type Initialisable interface {
	// Initialise and validate the struct. The struct may be modified by
	// this function.
	Initialise(vctx *Context)
}

type Failed struct {
	name string
	errs []error
}

func (v *Failed) Name() string { return v.name }

func (v *Failed) Errs() []error { return v.errs }

func (v *Failed) Error() string {
	return fmt.Sprintf("initialise failed: %q", v.name)
}

func Initialise(name string, v Initialisable) (rerr error) {
	defer func() {
		// Panic takes precedence over validation errors:
		if err := recover(); err != nil {
			rerr = err.(error)
		}
	}()

	vctx := Context{}
	v.Initialise(&vctx)
	if len(vctx.Errors) > 0 {
		errs := make([]error, len(vctx.Errors))
		for i := 0; i < len(errs); i++ {
			errs[i] = vctx.Errors[i]
		}
		return &Failed{name: name, errs: errs}
	}

	return nil
}

type Error struct {
	Path    string
	Message string
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Path, err.Message)
}

type Context struct {
	Errors []*Error
	path   []any
}

func (val *Context) Addf(msg string, args ...interface{}) {
	val.Errors = append(val.Errors, &Error{
		Path:    val.pathString(),
		Message: fmt.Sprintf(msg, args...),
	})
}

func (val *Context) PushKey(key string) {
	val.path = append(val.path, key)
}

func (val *Context) PushIndex(idx int) {
	val.path = append(val.path, idx)
}

func (val *Context) Pop() {
	val.path = val.path[:len(val.path)-1]
}

func (val *Context) DrillIndex(idx int, initialise func() error) {
	val.PushIndex(idx)
	defer val.Pop()
	initialise()
}

func (val *Context) DrillKey(key string, initialise func()) {
	val.PushKey(key)
	defer val.Pop()
	initialise()
}

func (val *Context) pathString() string {
	if len(val.path) == 0 {
		return "/"
	}
	var sb strings.Builder
	for _, part := range val.path {
		sb.WriteByte('/')
		switch part := part.(type) {
		case string:
			sb.WriteString(part)
		case int:
			sb.WriteString(strconv.Itoa(part))
		default:
			panic(fmt.Errorf("unexpected path segment %T", part))
		}
	}
	return sb.String()
}
