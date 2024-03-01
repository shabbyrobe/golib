package ring

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

type RingFactory[T any] func(n int) Ring[T]

func makeRingFactories[T any]() map[string]RingFactory[T] {
	return map[string]RingFactory[T]{
		"fixed":           func(n int) Ring[T] { return NewFixedRing[T](n) },
		"dynamic":         func(n int) Ring[T] { return NewDynamicRing[T](0, n) },
		"dynamicfull":     func(n int) Ring[T] { return NewDynamicRing[T](n, n) },
		"dynamicsync":     func(n int) Ring[T] { return NewDynamicSyncRing[T](0, n) },
		"dynamicsyncfull": func(n int) Ring[T] { return NewDynamicSyncRing[T](n, n) },
	}
}

func assertTail[T comparable](t *testing.T, buf Ring[T], n int, vs ...T) {
	t.Helper()
	into := make([]T, n)
	result := into[:buf.PeekTail(into)]
	if vs == nil {
		vs = []T{}
	}
	if !reflect.DeepEqual(result, vs) {
		t.Fatal("expected", "vs", vs, "!= actual", result)
	}
}

func assertHead[T comparable](t *testing.T, buf Ring[T], n int, vs ...T) {
	t.Helper()
	into := make([]T, n)
	result := into[:buf.PeekHead(into)]
	if vs == nil {
		vs = []T{}
	}
	if !reflect.DeepEqual(result, vs) {
		t.Fatal("expected", "vs", vs, "!= actual", result)
	}
}

func assertNext[T comparable](t *testing.T, buf Ring[T], n int, vs ...T) {
	t.Helper()

	assertHead(t, buf, n, vs...) // Might as well.

	into := make([]T, n)
	result := into[:buf.Next(into)]
	if vs == nil {
		vs = []T{}
	}
	if !reflect.DeepEqual(result, vs) {
		t.Fatal("expected", "vs", vs, "!= actual", result)
	}
}

func TestBufferOverwrite(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s/7-elements", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			buf.Add("6")
			buf.Add("7")
			assertNext(t, buf, 5, "5", "6", "7")
		})

		t.Run(fmt.Sprintf("%s/6-elements", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			buf.Add("6")
			assertNext(t, buf, 5, "4", "5", "6")
		})

		t.Run(fmt.Sprintf("%s/5-elements", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			assertNext(t, buf, 5, "3", "4", "5")
		})
	}
}

func TestBufferWrapAfterEmptyPriorToFilling(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Next(make([]string, 2))
			assertTail(t, buf, 1)
			buf.Add("3")
			assertTail(t, buf, 1, "3")
			buf.Add("4")
			assertTail(t, buf, 2, "3", "4")
			buf.Add("5")
			assertTail(t, buf, 3, "3", "4", "5")
			buf.Add("6")
			assertTail(t, buf, 3, "4", "5", "6")
			buf.Add("7")
			assertNext(t, buf, 5, "5", "6", "7")
		})
	}
}

func TestBufferWrapAfterUnfilledRingShrunkButNotDrained(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Next(make([]string, 2))
			assertTail(t, buf, 1, "3")
			buf.Add("4")
			assertTail(t, buf, 2, "3", "4")
			buf.Add("5")
			assertTail(t, buf, 3, "3", "4", "5")
			buf.Add("6")
			assertTail(t, buf, 3, "4", "5", "6")
			buf.Add("7")
			assertNext(t, buf, 5, "5", "6", "7")
		})
	}
}

func TestBufferTail(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s/len=3,off=0", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			assertTail(t, buf, 1, "3")
			assertTail(t, buf, 2, "2", "3")
			assertTail(t, buf, 3, "1", "2", "3")
			assertTail(t, buf, 4, "1", "2", "3")
		})

		t.Run(fmt.Sprintf("%s/len=3,off=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			assertTail(t, buf, 1, "4")
			assertTail(t, buf, 2, "3", "4")
			assertTail(t, buf, 3, "2", "3", "4")
			assertTail(t, buf, 4, "2", "3", "4")
		})

		t.Run(fmt.Sprintf("%s/len=3,off=2", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			assertTail(t, buf, 1, "5")
			assertTail(t, buf, 2, "4", "5")
			assertTail(t, buf, 3, "3", "4", "5")
			assertTail(t, buf, 4, "3", "4", "5")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=0,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			assertNext(t, buf, 1, "1")

			assertTail(t, buf, 1, "3")
			assertTail(t, buf, 2, "2", "3")
			assertTail(t, buf, 3, "2", "3")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=1,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			assertNext(t, buf, 1, "2")

			assertTail(t, buf, 1, "4")
			assertTail(t, buf, 2, "3", "4")
			assertTail(t, buf, 3, "3", "4")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=2,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			assertNext(t, buf, 1, "3")

			assertTail(t, buf, 1, "5")
			assertTail(t, buf, 2, "4", "5")
			assertTail(t, buf, 3, "4", "5")
		})
	}
}

func TestBufferHead(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s/len=3,off=0", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			assertHead(t, buf, 1, "1")
			assertHead(t, buf, 2, "1", "2")
			assertHead(t, buf, 3, "1", "2", "3")
			assertHead(t, buf, 4, "1", "2", "3")
		})

		t.Run(fmt.Sprintf("%s/len=3,off=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			assertHead(t, buf, 1, "2")
			assertHead(t, buf, 2, "2", "3")
			assertHead(t, buf, 3, "2", "3", "4")
			assertHead(t, buf, 4, "2", "3", "4")
		})

		t.Run(fmt.Sprintf("%s/len=3,off=2", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			assertHead(t, buf, 1, "3")
			assertHead(t, buf, 2, "3", "4")
			assertHead(t, buf, 3, "3", "4", "5")
			assertHead(t, buf, 4, "3", "4", "5")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=0,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			assertNext(t, buf, 1, "1")

			assertHead(t, buf, 1, "2")
			assertHead(t, buf, 2, "2", "3")
			assertHead(t, buf, 3, "2", "3")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=1,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			assertNext(t, buf, 1, "2")

			assertHead(t, buf, 1, "3")
			assertHead(t, buf, 2, "3", "4")
			assertHead(t, buf, 3, "3", "4")
		})

		t.Run(fmt.Sprintf("%s/len=2,off=2,shift=1", factoryName), func(t *testing.T) {
			buf := ringFactory(3)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")
			assertNext(t, buf, 1, "3")

			assertHead(t, buf, 1, "4")
			assertHead(t, buf, 2, "4", "5")
			assertHead(t, buf, 3, "4", "5")
		})
	}
}

func TestBufferNextOne(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s", factoryName), func(t *testing.T) {
			buf := ringFactory(5)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")

			assertNext(t, buf, 1, "1")
			assertNext(t, buf, 1, "2")
			assertNext(t, buf, 1, "3")
			assertNext(t, buf, 1)
			assertTail(t, buf, 1)
		})
	}
}

func TestBufferNextTwo(t *testing.T) {
	for factoryName, ringFactory := range makeRingFactories[string]() {
		t.Run(fmt.Sprintf("%s", factoryName), func(t *testing.T) {
			buf := ringFactory(5)
			buf.Add("1")
			buf.Add("2")
			buf.Add("3")
			buf.Add("4")
			buf.Add("5")

			assertNext(t, buf, 2, "1", "2")
			assertNext(t, buf, 2, "3", "4")
			assertNext(t, buf, 2, "5")
			assertNext(t, buf, 2)
			assertTail(t, buf, 2)
		})
	}
}

func TestBufferSpam(t *testing.T) {
	seed := int64(2)
	rng := rand.New(rand.NewSource(seed))

	scratch := make([]int, 10000)
	iters := 5

	for iter := 0; iter < iters; iter++ {
		for _, c := range []struct {
			initial int
			limit   int
			events  int // adds or removes
		}{
			{initial: 0, limit: 5, events: 50},
			{initial: 0, limit: 5, events: 5000},
			{initial: 1, limit: 5, events: 50},
			{initial: 1, limit: 5, events: 5000},
			{initial: 5, limit: 5, events: 50},
			{initial: 5, limit: 5, events: 5000},

			{initial: 0, limit: 100, events: 1000},
			{initial: 0, limit: 100, events: 100000},
			{initial: 1, limit: 100, events: 1000},
			{initial: 1, limit: 100, events: 100000},
			{initial: 100, limit: 100, events: 1000},
			{initial: 100, limit: 100, events: 100000},

			{initial: 0, limit: 500, events: 1000},
			{initial: 0, limit: 500, events: 100000},
			{initial: 1, limit: 500, events: 1000},
			{initial: 1, limit: 500, events: 100000},
			{initial: 500, limit: 500, events: 1000},
			{initial: 500, limit: 500, events: 100000},
		} {
			for factoryName, ringFactory := range makeRingFactories[int]() {
				t.Run(fmt.Sprintf("%s/iter=%d/initial=%d/limit=%d/events=%d", factoryName, iter, c.initial, c.limit, c.events), func(t *testing.T) {
					ring := ringFactory(c.limit)
					empties := 0
					fulls := 0
					lens := make([]int, c.limit+1)

					for i := 0; i < c.events; i++ {
						add := ring.Len() == 0 || rng.Intn(2) == 1
						if add {
							cnt := rng.Intn(100) + 1
							for j := 0; j < cnt; j++ {
								ring.Add(1)
							}
							if ring.Len() == ring.Cap() {
								fulls++
							}
						} else {
							cnt := rng.Intn(100) + 1
							ring.Next(scratch[:cnt])
							if ring.Len() == 0 {
								empties++
							}
						}

						lens[ring.Len()]++
					}

					if empties == 0 {
						t.Fatal()
					} else if fulls == 0 {
						t.Fatal()
					}
				})
			}
		}
	}
}
