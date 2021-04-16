package frap

// Find rational approximation for given real number.
//
// Based on David Eppstein's algorithm, available here:
// https://www.ics.uci.edu/~eppstein/numth/frap.c
//
//	based on the theory of continued fractions
// 	if x = a1 + 1/(a2 + 1/(a3 + 1/(a4 + ...)))
// 	then best approximation is found by truncating this series
// 	(with some adjustments in the last term).
//
// 	Note the fraction can be recovered as the first column of the matrix
// 	 ( a1 1 ) ( a2 1 ) ( a3 1 ) ...
// 	 ( 1  0 ) ( 1  0 ) ( 1  0 )
// 	Instead of keeping the sequence of continued fraction terms,
// 	we just keep the last partial product of these matrices.
//
func RationalApprox(rat float64, maxDenom int) (num, denom int) {
	type mat [2][2]int

	var m mat = [2][2]int{{1, 0}, {0, 1}}
	var x, startx = rat, rat
	var ai int
	_ = startx

	// loop finding terms until denom gets too big
	for {
		ai = int(x)
		nextDenom := (m[1][0] * ai) + m[1][1]
		if nextDenom > maxDenom {
			break
		}
		t := m[0][0]*ai + m[0][1]
		m[0][1], m[0][0] = m[0][0], t

		t = m[1][0]*ai + m[1][1]
		m[1][1], m[1][0] = m[1][0], t

		if x == float64(ai) { // AF: division by zero
			break
		}

		x = 1 / (x - float64(ai))
		if x > float64(0x7FFFFFFF) {
			break // AF: representation failure
		}
	}

	// now remaining x is between 0 and 1/ai
	// approx as either 0 or 1/m where m is max that will fit in maxden
	// first try zero
	return m[0][0], m[1][0]

	// Where 'startx' is the original input float64, the original algo
	// calculates the error amount like this:
	//   startx - (float64(m[0][0]) / float64(m[1][0]))

	// Original algo also contains this secondary option, which only seems to
	// to produce a worse result?
	//   ai = (maxden - m[1][1]) / m[1][0];
	//   m[0][0] = m[0][0] * ai + m[0][1];
	//   m[1][0] = m[1][0] * ai + m[1][1];
	//   printf("%ld/%ld, error = %e\n", m[0][0], m[1][0], startx - ((double) m[0][0] / (double) m[1][0]));
}
