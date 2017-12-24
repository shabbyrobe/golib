package stringtools

func similarBytes(a, b []byte, alen, blen int) int {
	var l int
	var max int
	var pos1, pos2 int
	for p := 0; p < alen; p++ {
		for q := 0; q < alen; q++ {
			for l = 0; (p+l < alen) && (q+l < blen) && (a[p+l] == b[q+l]); l++ {
				// ^ this is why people hate C. if you program in C, never do this.
			}
			if l > max {
				max = l
				pos1 = p
				pos2 = q
			}
		}
	}

	sum := max
	ab, bb := []byte(a), []byte(b)
	if sum > 0 {
		if pos1 > 0 && pos2 > 0 {
			sum += similarBytes(ab[0:pos1], bb[0:pos2], pos1, pos2)
		}
		if (pos1+max < alen) && (pos2+max < blen) {
			as, al := pos1+max, alen-pos1-max
			bs, bl := pos2+max, blen-pos2-max
			sum += similarBytes(ab[as:(al+as)], bb[bs:(bl+bs)], al, bl)
		}
	}

	return sum
}

func Similar(a, b string) (sum int, pct float64) {
	// TODO: non-recursive solution

	alen, blen := len(a), len(b)

	if alen == 0 && blen == 0 {
		return 0, 1.0
	}

	sum = similarBytes([]byte(a), []byte(b), alen, blen)
	pct = float64(sum) * 2.0 / float64(alen+blen)
	return sum, pct
}

func SimilarPercent(a, b string) float64 {
	alen, blen := len(a), len(b)

	if alen == 0 && blen == 0 {
		return 1.0
	}

	sum := similarBytes([]byte(a), []byte(b), alen, blen)
	return float64(sum) * 2.0 / float64(alen+blen)
}
