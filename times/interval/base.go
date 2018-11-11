package interval

type (
	// Span identifies the unit of the Interval, i.e. Seconds, Minutes, etc.
	Span uint8

	// Qty is the number of Spans in an Interval.
	Qty uint

	// Period identifies the number of Intervals that have passed since the
	// Unix Epoch.
	Period int64

	// Interval combines a Span and a Qty into a single value.
	Interval uint16
)
