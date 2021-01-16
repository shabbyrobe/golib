package interval

type (
	// Unit identifies the unit of the Interval, i.e. Seconds, Minutes, etc.
	Unit uint8

	// Qty is the number of Units in an Interval. It should only be used as a
	// component of an Interval, not as a way to represent many intervals (see
	// Span).
	Qty uint

	// Period identifies the number of Intervals that have passed since the
	// Unix Epoch.
	Period int64

	// Interval combines a Unit and a Qty into a single value.
	Interval uint16

	// Span combines an interval with a count. Where Qty is intended to only be used to
	// create an Interval (hence the small data type), Span is intended to represent one
	// or more of those intervals in a series of any length.
	// FIXME: re-enable once Unit is rolled out.
	// Span struct {
	//     Interval
	//     Num int64
	// }

	// Moment combines Period and Interval as a complete representation of a specific
	// moment of UTC time.
	Moment struct {
		Interval
		Period
	}

	// Range of time represented as a half open interval [Since, Until)
	Range struct {
		Interval
		Since, Until Period
	}
)
