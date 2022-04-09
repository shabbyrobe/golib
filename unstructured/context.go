package unstructured

type Context interface {
	AddError(err error) error

	// Pop the most recent error from the end of the error list and return it:
	PopError() error

	// Shift the most recent error from the head of the error list and return it:
	ShiftError() error
}
