package errtools

// This is very similar to github.com/pkg/errors.Cause

type Causer interface {
	Cause() error
}

func Causes(err error) (out []error) {
	for err != nil {
		out = append(out, err)
		cause, ok := err.(Causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return
}

func Cause(err error) error {
	var last error

	for err != nil {
		cause, ok := err.(Causer)
		if !ok {
			break
		}
		err = cause.Cause()

		if err == last {
			break
		}

		last = err
	}

	return err
}

func ParentCause(err error) error {
	cause, ok := err.(Causer)
	if ok {
		return cause.Cause()
	} else {
		return err
	}
}
