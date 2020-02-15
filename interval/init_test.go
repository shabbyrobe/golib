package interval

import (
	"math/rand"
	"time"
)

var spansLen = len(Spans)

func randomInterval(rng *rand.Rand) Interval {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}
	span := Span(gimmeRandom(spansLen) + int(firstSpan))
	qty := Qty(gimmeRandom(int(span.MaxQty())))
	return OfValid(qty, span)
}

func randomDifferentIntervals(rng *rand.Rand) (a, b Interval) {
	from := randomInterval(nil)
	to := from
	for from == to {
		to = randomInterval(nil)
	}
	return from, to
}

func randomDivisibleIntervals(rng *rand.Rand) (from, by Interval) {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}

	var fromSpan Span
	for fromSpan == 0 || fromSpan == firstSpan {
		fromSpan = Span(gimmeRandom(spansLen) + int(firstSpan))
	}

	var bySpan Span
	for bySpan == 0 || bySpan == lastSpan || bySpan >= fromSpan {
		bySpan = Span(gimmeRandom(spansLen) + int(firstSpan))
	}

	for {
		from = OfValid(randomQty(rng, fromSpan), fromSpan)
		by = OfValid(randomQty(rng, bySpan), bySpan)
		if from.CanDivideBy(by) {
			return from, by
		}
	}
}

func randomQty(rng *rand.Rand, span Span) Qty {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}

	// Qty must not be zero
	v := gimmeRandom(int(span.MaxQty())-1) + 1
	return Qty(v)
}

func randomPeriod(rng *rand.Rand, intvl Interval) Period {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}

	// This is an int overflow waiting to happen - if you have an interval of
	// 150 years, the period can't get too big otherwise you will overflow the
	// 64 bit time representation. This is not currently well handled by this
	// lib.
	max := intvl.Period(time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC))
	return Period(gimmeRandom(int(max)))
}
