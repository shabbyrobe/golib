package bytestream

import "io"

// ByteStream allows a stream of bytes to be taken in discrete chunks, allowing
// things like buffering across reads to be masked from calling code.
//
// All implementers of ByteStream MUST cache errors, to be returned by Err(),
// and return the cached error for all byte-returning operations if one exists.
//
// This allows calling code to ignore errors until predefined "checkpoints", which
// are then checked using Err().
//
// Individual implementations are free to place their own limits in the maximum
// number of bytes that can be retrieved in a single call; that limit MUST be
// available via Limt(). Exceeding this limit MUST result in `errors.Is(err,
// io.ErrShortBuffer) == true`.
//
type ByteStream interface {
	io.ByteReader

	// Byte position since the start of the stream
	Tell() int64

	// The maximum number of bytes that you can Peek or Take without seeing an
	// io.ErrShortBuffer. This limit should be fixed for the entire lifetime of the
	// stream.
	//
	// This does not return the number of bytes remaining, and should not be used
	// to determine if TakeExactly would succeed.
	Limit() int64

	// Return any cached errors
	Err() error

	DiscardExactly(n int) (err error)
	DiscardUpTo(n int) error
	PeekExactly(n int) (o []byte, err error)
	PeekUpTo(n int) (o []byte, err error)
	TakeExactly(n int) (o []byte, err error)
	TakeUpTo(n int) (o []byte, err error)
}
