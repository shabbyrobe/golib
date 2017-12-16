// +build ignore

package service

import "testing"

func mustStartHalt(tt T, ss ...statService) {
	tt.Helper()
	for idx, s := range ss {
		tt.MustEqual(1, s.Starts(), "%d - %s", idx, s.ServiceName())
		tt.MustEqual(1, s.Halts(), "%d - %s", idx, s.ServiceName())
	}
}

func TestGroup(t *testing.T) {
	tt := WrapTB(t)

	s1 := (&blockingService{}).Init()
	s2 := (&blockingService{}).Init()
	r := NewRunner(newDummyListener())
	g := NewGroup("yep", []Service{s1, s2})

	tt.MustOK(r.StartWait(g, dto))
	tt.MustOK(r.Halt(g, dto))
	mustStartHalt(tt, s1, s2)
}

func TestGroupEndOne(t *testing.T) {
	tt := WrapTB(t)

	s1 := (&blockingService{}).Init()
	s2 := (&blockingService{}).Init()
	s3 := &dummyService{runTime: 2 * tscale}

	lc := newListenerCollector()
	r := NewRunner(lc)
	g := NewGroup("yep", []Service{s1, s2, s3})

	tt.MustOK(r.StartWait(g, dto))
	<-lc.endWaiter(g)
	mustStartHalt(tt, s1, s2, s3)
	tt.MustEqual(1, len(lc.ends(g)))

	tt.MustOK(EnsureHalt(r, g, dto))
}

func TestGroupEndMultiple(t *testing.T) {
	tt := WrapTB(t)

	s1 := (&blockingService{}).Init()
	s2 := &dummyService{runTime: 2 * tscale}
	s3 := &dummyService{runTime: 2 * tscale}

	lc := newListenerCollector()
	r := NewRunner(lc)
	g := NewGroup("yep", []Service{s1, s2, s3})

	tt.MustOK(r.StartWait(g, dto))
	<-lc.endWaiter(g)
	mustStartHalt(tt, s1, s2, s3)
	tt.MustEqual(1, len(lc.ends(g)))

	tt.MustOK(EnsureHalt(r, g, dto))
}

func TestGroupEndAll(t *testing.T) {
	tt := WrapTB(t)

	s1 := &dummyService{runTime: 2 * tscale}
	s2 := &dummyService{runTime: 2 * tscale}
	s3 := &dummyService{runTime: 2 * tscale}

	lc := newListenerCollector()
	r := NewRunner(lc)
	g := NewGroup("yep", []Service{s1, s2, s3})

	tt.MustOK(r.StartWait(g, dto))
	<-lc.endWaiter(g)
	mustStartHalt(tt, s1, s2, s3)
	tt.MustEqual(1, len(lc.ends(g)))

	tt.MustOK(EnsureHalt(r, g, dto))
}
