package interval

import (
	"math/rand"
	"time"
)

var unitsLen = len(Units)

func tm(rfc3339 string) time.Time {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		panic(err)
	}
	return t
}

func randomInterval(rng *rand.Rand) Interval {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}
	unit := Unit(gimmeRandom(unitsLen) + int(firstUnit))
	qty := Qty(gimmeRandom(int(unit.MaxQty())))
	return OfValid(qty, unit)
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

	var fromUnit Unit
	for fromUnit == 0 || fromUnit == firstUnit {
		fromUnit = Unit(gimmeRandom(unitsLen) + int(firstUnit))
	}

	var byUnit Unit
	for byUnit == 0 || byUnit == lastUnit || byUnit >= fromUnit {
		byUnit = Unit(gimmeRandom(unitsLen) + int(firstUnit))
	}

	for {
		from = OfValid(randomQty(rng, fromUnit), fromUnit)
		by = OfValid(randomQty(rng, byUnit), byUnit)
		if from.CanDivideBy(by) {
			return from, by
		}
	}
}

func randomQty(rng *rand.Rand, unit Unit) Qty {
	gimmeRandom := rand.Intn
	if rng != nil {
		gimmeRandom = rng.Intn
	}

	// Qty must not be zero
	v := gimmeRandom(int(unit.MaxQty())-1) + 1
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
