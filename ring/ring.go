package ring

import "sync"

type Ring[T any] interface {
	Len() int
	Cap() int
	Next(into []T) int
	Add(item T)
	PeekTail(into []T) (n int)
	PeekHead(into []T) (n int)
}

type DynamicRing[T any] struct {
	ring *FixedRing[T]
	cap  int
}

func NewDynamicRing[T any](initial int, limit int) *DynamicRing[T] {
	if limit <= 0 {
		panic("0-limit ring")
	}
	if initial <= 0 {
		initial = 1
	}
	if limit < initial {
		limit = initial
	}
	return &DynamicRing[T]{
		ring: NewFixedRing[T](initial),
		cap:  limit,
	}
}

func (r *DynamicRing[T]) Len() int                  { return r.ring.Len() }
func (r *DynamicRing[T]) Cap() int                  { return r.cap }
func (r *DynamicRing[T]) Next(into []T) int         { return r.ring.Next(into) }
func (r *DynamicRing[T]) PeekTail(into []T) (n int) { return r.ring.PeekTail(into) }
func (r *DynamicRing[T]) PeekHead(into []T) (n int) { return r.ring.PeekHead(into) }

func (r *DynamicRing[T]) Add(item T) {
	// If we aren't done growing and the ring is full, time to grow:
	if r.ring.Cap() < r.cap && r.ring.Len() == r.ring.Cap() {
		newCap := grow(r.ring.Cap()+1, r.ring.Cap())
		if newCap > r.cap {
			newCap = r.cap
		}

		// Assert that this shouldn't happen, at least until we're sure it's stable
		if newCap == r.ring.Cap() {
			panic("unexpected cap")
		}

		next := NewFixedRing[T](newCap)
		n := r.ring.Next(next.items)
		next.len = n
		next.next = n
		r.ring = next
	}

	r.ring.Add(item)
	return
}

type DynamicSyncRing[T any] struct {
	ring *DynamicRing[T]
	mu   sync.Mutex
}

func NewDynamicSyncRing[T any](initial int, limit int) *DynamicSyncRing[T] {
	return &DynamicSyncRing[T]{
		ring: NewDynamicRing[T](initial, limit),
	}
}

func (r *DynamicSyncRing[T]) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ring.Len()
}

func (r *DynamicSyncRing[T]) Cap() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ring.Cap()
}

func (r *DynamicSyncRing[T]) Next(into []T) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ring.Next(into)
}

func (r *DynamicSyncRing[T]) PeekTail(into []T) (n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ring.PeekTail(into)
}

func (r *DynamicSyncRing[T]) PeekHead(into []T) (n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ring.PeekHead(into)
}

func (r *DynamicSyncRing[T]) Add(item T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ring.Add(item)
}

type FixedRing[T any] struct {
	items []T
	head  int // Index of the first item.
	next  int // Index of the next slot to write in.
	len   int
}

func NewFixedRing[T any](sz int) *FixedRing[T] {
	if sz <= 0 {
		panic("ring buffer: max must be > 0")
	}
	return &FixedRing[T]{
		items: make([]T, sz),
	}
}

func (r *FixedRing[T]) Len() int { return r.len }
func (r *FixedRing[T]) Cap() int { return len(r.items) }

func (r *FixedRing[T]) Next(into []T) (n int) {
	if r.len == 0 {
		return 0
	}

	end := len(into)
	if end > r.len {
		end = r.len
	}

	rdPos := r.head
	for ; n < end; n++ {
		into[n] = r.items[rdPos]
		rdPos++
		if rdPos == len(r.items) {
			rdPos = 0
		}
	}

	r.head = rdPos
	r.len -= n

	return n
}

func (r *FixedRing[T]) Add(item T) {
	r.items[r.next] = item
	if r.len >= len(r.items) {
		r.head++
		if r.head >= r.len {
			r.head = 0
		}
	} else {
		r.len++
	}
	r.next++
	if r.next >= len(r.items) {
		r.next = 0
	}
}

func (r *FixedRing[T]) PeekHead(into []T) (n int) {
	if r.len == 0 {
		return 0
	}

	left := len(into)
	if left > r.len {
		left = r.len
	}

	if r.head < r.next {
		rdStart, rdEnd := r.head, r.next
		rdSz := rdEnd - rdStart
		if rdSz > left {
			rdEnd = rdStart + left
		}
		return copy(into, r.items[rdStart:rdEnd])

	} else {
		rdStart, rdEnd := r.head, len(r.items)
		rdSz := rdEnd - rdStart
		if rdSz > left {
			rdEnd = rdStart + left
			return copy(into, r.items[rdStart:rdEnd])
		}

		c := copy(into, r.items[rdStart:rdEnd])
		n += c
		left -= c

		rdStart, rdEnd = 0, r.next
		rdSz = rdEnd - rdStart
		if rdSz > left {
			rdEnd = rdStart + left
		}
		n += copy(into[n:], r.items[rdStart:rdEnd])

		return n
	}
}

func (r *FixedRing[T]) PeekTail(into []T) (n int) {
	if r.len == 0 {
		return 0
	}

	left := len(into)
	if left > r.len {
		left = r.len
	}

	if r.head < r.next {
		start := r.head
		if r.next-r.head > left {
			start = r.next - left
		}
		return copy(into, r.items[start:r.next])

	} else {
		wrStart := 0
		rdStart, rdEnd := 0, r.next
		rdSz := rdEnd - rdStart
		if rdSz > left {
			rdStart = rdEnd - left
			return copy(into, r.items[rdStart:rdEnd])
		} else {
			wrStart = left - rdSz
		}

		c := copy(into[wrStart:], r.items[rdStart:rdEnd])
		n += c
		left -= c

		rdStart, rdEnd = r.head, len(r.items)
		rdSz = rdEnd - rdStart
		if rdSz > left {
			rdStart = rdEnd - left
		}
		n += copy(into, r.items[rdStart:rdEnd])
		return n
	}
}

// Go's slice growing algorithm as at 1.21:
func grow(targetCap, oldCap int) int {
	const threshold = 256

	dbl := oldCap * 2
	if dbl < targetCap {
		return targetCap

	} else if oldCap < threshold {
		return dbl

	} else {
		newCap := oldCap
		for 0 < newCap && newCap < targetCap {
			newCap += (newCap + 3*threshold) / 4
		}
		if newCap <= 0 {
			newCap = targetCap
		}
		return newCap
	}
}
