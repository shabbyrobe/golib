package broadcast

type Listener[T any] struct {
	drop bool
	c    chan T
}

func (l Listener[T]) C() <-chan T {
	return l.c
}

func NewListener[T any](sz int, drop bool) Listener[T] {
	c := make(chan T, sz)
	return Listener[T]{
		drop: drop,
		c:    c,
	}
}

func ToListener[T any](ch chan T, drop bool) Listener[T] {
	return Listener[T]{
		drop: drop,
		c:    ch,
	}
}

type Broadcaster[T any] struct {
	in   chan T
	add  chan Listener[T]
	rem  chan Listener[T]
	out  map[Listener[T]]struct{}
	stop chan struct{}
	done chan struct{}
}

func New[T any](bufsz int) *Broadcaster[T] {
	b := &Broadcaster[T]{
		in:   make(chan T, bufsz),
		add:  make(chan Listener[T]),
		rem:  make(chan Listener[T]),
		out:  map[Listener[T]]struct{}{},
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go b.background()
	return b
}

func (b *Broadcaster[T]) Send(item T) {
	select {
	case b.in <- item:
	case <-b.stop:
		panic("send to a stopped broadcaster")
	}
}

func (b *Broadcaster[T]) TrySend(item T) bool {
	select {
	case b.in <- item:
		return true
	case <-b.stop:
		panic("send to a stopped broadcaster")
	default:
		return false
	}
}

func (b *Broadcaster[T]) Channel(sz int, drop bool) (ch <-chan T, done func(), err error) {
	l := b.Listen(sz, drop)
	done = func() { b.Remove(l) }
	return l.C(), done, nil
}

func (b *Broadcaster[T]) Listen(sz int, drop bool) Listener[T] {
	l := NewListener[T](sz, drop)
	b.Add(l)
	return l
}

func (b *Broadcaster[T]) Add(l Listener[T]) {
	select {
	case b.add <- l:
	case <-b.stop:
		panic("add to a stopped broadcaster")
	}
}

func (b *Broadcaster[T]) Remove(l Listener[T]) {
	select {
	case b.rem <- l:
	case <-b.stop:
		panic("remove from a stopped broadcaster")
	}
}

func (b *Broadcaster[T]) Close() error {
	close(b.stop)
	<-b.done
	return nil
}

func (b *Broadcaster[T]) background() {
	for {
		select {
		case in := <-b.in:
		broadcast:
			for out := range b.out {
				if out.drop {
					select {
					default:
					case out.c <- in:
					case <-b.stop:
						break broadcast
					}

				} else {
					select {
					case out.c <- in:
					case <-b.stop:
						break broadcast
					}
				}
			}

		case <-b.stop:
			close(b.done)
			return

		case ch := <-b.add:
			b.out[ch] = struct{}{}

		case ch := <-b.rem:
			delete(b.out, ch)
		}
	}
}
