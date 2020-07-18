package bytestream

// ByteStream allows a stream of bytes to be taken in discrete chunks, allowing
// things like buffering across reads to be masked from calling code.
//
// All implementers of ByteStream MUST cache errors, to be returned by Err(),
// and return the cached error for all byte-returning operations if one exists.
//
// This allows calling code to ignore errors until predefined "checkpoints", which
// are then checked using Err().
type ByteStream interface {
	// Byte position since the start of the stream
	Tell() int64

	// Return any cached errors
	Err() error

	DiscardExactly(n int) (err error)
	DiscardUpTo(n int) error
	PeekExactly(n int) (o []byte, err error)
	PeekUpTo(n int) (o []byte, err error)
	TakeExactly(n int) (o []byte, err error)
	TakeUpTo(n int) (o []byte, err error)
}
