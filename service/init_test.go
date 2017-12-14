package service

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"
)

var (
	fuzzEnabled  bool
	fuzzTimeSec  float64
	fuzzTickNsec int64
	fuzzSeed     int64
)

func TestMain(m *testing.M) {
	flag.BoolVar(&fuzzEnabled, "service.fuzz", false, "Fuzz? Nope by default.")
	flag.Float64Var(&fuzzTimeSec, "service.fuzztime", float64(1*time.Second)/float64(time.Second), "Run the fuzzer for this many seconds")
	flag.Int64Var(&fuzzTickNsec, "service.fuzzticknsec", 0, "How frequently to tick in the fuzzer's loop.")
	flag.Int64Var(&fuzzSeed, "service.fuzzseed", 0, "Randomise the fuzz tester with this seed prior to every fuzz test")

	beforeCount := pprof.Lookup("goroutine").Count()
	code := m.Run()

	if code == 0 {
		// This little hack is to give things like "go OnServiceState" a chance
		// to finish - it routinely shows up in the profile
		time.Sleep(20 * time.Millisecond)

		after := pprof.Lookup("goroutine")
		afterCount := after.Count()

		diff := afterCount - beforeCount
		if diff > 0 {
			var buf bytes.Buffer
			after.WriteTo(&buf, 1)
			fmt.Fprintf(os.Stderr, "stray goroutines: %d\n%s\n", diff, buf.String())
			os.Exit(2)
		}
	}
	os.Exit(code)
}

type listenerCollectorEnd struct {
	err error
}

type listenerCollectorService struct {
	errs       []Error
	states     []State
	endWaiters []chan struct{}
	ends       []*listenerCollectorEnd
}

type listenerCollector struct {
	services map[Service]*listenerCollectorService
	lock     sync.Mutex
}

func newListenerCollector() *listenerCollector {
	return &listenerCollector{
		services: make(map[Service]*listenerCollectorService),
	}
}

func (t *listenerCollector) errs(service Service) (out []Error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	svc := t.services[service]
	if svc == nil {
		return
	}
	for _, e := range svc.errs {
		out = append(out, e)
	}
	return
}

func (t *listenerCollector) ends(service Service) (out []listenerCollectorEnd) {
	t.lock.Lock()
	defer t.lock.Unlock()
	svc := t.services[service]
	if svc == nil {
		return
	}
	for _, e := range svc.ends {
		out = append(out, *e)
	}
	return
}

func (t *listenerCollector) endWaiter(service Service) chan struct{} {
	t.lock.Lock()
	if t.services[service] == nil {
		t.services[service] = &listenerCollectorService{}
	}
	svc := t.services[service]
	w := make(chan struct{}, 1)
	svc.endWaiters = append(svc.endWaiters, w)
	t.lock.Unlock()

	return w
}

func (t *listenerCollector) OnServiceState(service Service, state State) {
	t.lock.Lock()
	if t.services[service] == nil {
		t.services[service] = &listenerCollectorService{}
	}
	svc := t.services[service]
	svc.states = append(svc.states, state)
	t.lock.Unlock()
}

func (t *listenerCollector) OnServiceError(service Service, err Error) {
	t.lock.Lock()
	if t.services[service] == nil {
		t.services[service] = &listenerCollectorService{}
	}
	svc := t.services[service]
	svc.errs = append(svc.errs, err)
	t.lock.Unlock()
}

func (t *listenerCollector) OnServiceEnd(service Service, err Error) {
	t.lock.Lock()
	if t.services[service] == nil {
		t.services[service] = &listenerCollectorService{}
	}
	svc := t.services[service]

	svc.ends = append(svc.ends, &listenerCollectorEnd{
		err: Cause(err),
	})
	if len(svc.endWaiters) > 0 {
		for _, w := range svc.endWaiters {
			close(w)
		}
		svc.endWaiters = nil
	}
	t.lock.Unlock()
}

type dummyListener struct {
}

func newDummyListener() *dummyListener {
	return &dummyListener{}
}

func (t *dummyListener) OnServiceState(service Service, state State) {
}

func (t *dummyListener) OnServiceError(service Service, err Error) {
}

func (t *dummyListener) OnServiceEnd(service Service, err Error) {
}

type dummyService struct {
	name         Name
	startFailure error
	startDelay   time.Duration
	runFailure   error
	runTime      time.Duration
	haltDelay    time.Duration
	haltingSleep bool
}

func (d *dummyService) ServiceName() Name {
	if d.name == "" {
		// This is a nasty cheat, don't do it in any real code!
		return Name(fmt.Sprintf("dummyService-%p", d))
	}
	return d.name
}

func (d *dummyService) Run(ctx Context) error {
	if d.startDelay > 0 {
		time.Sleep(d.startDelay)
	}
	if d.startFailure != nil {
		return d.startFailure
	}
	if err := ctx.Ready(d); err != nil {
		return err
	}
	if d.runTime > 0 {
		if d.haltingSleep {
			Sleep(ctx, d.runTime)
		} else {
			time.Sleep(d.runTime)
		}
	}
	if ctx.Halted() {
		if d.haltDelay > 0 {
			time.Sleep(d.haltDelay)
		}
		return nil
	} else {
		if d.runFailure == nil {
			return ErrServiceEnded
		}
		return d.runFailure
	}
}

type errorService struct {
	name       Name
	startDelay time.Duration
	errc       chan error
	buf        int
	init       bool
}

func (d *errorService) Init() *errorService {
	d.init = true
	if d.buf <= 0 {
		d.buf = 10
	}
	d.errc = make(chan error, d.buf)
	return d
}

func (d *errorService) ServiceName() Name {
	if d.name == "" {
		// This is a nasty cheat, don't do it in any real code!
		return Name(fmt.Sprintf("errorService-%p", d))
	}
	return d.name
}

func (d *errorService) Run(ctx Context) error {
	if !d.init {
		panic("call Init()!")
	}
	if d.startDelay > 0 {
		after := time.After(d.startDelay)
		for {
			select {
			case err := <-d.errc:
				ctx.OnError(d, err)
			case <-after:
				goto startDone
			}
		}
	startDone:
	}
	if err := ctx.Ready(d); err != nil {
		return err
	}
	for {
		select {
		case err := <-d.errc:
			ctx.OnError(d, err)
		case <-ctx.Halt():
			return nil
		}
	}
}

type blockingService struct {
	name         Name
	startFailure error
	runFailure   error
	startDelay   time.Duration
	haltDelay    time.Duration
	init         bool
}

func (d *blockingService) Init() *blockingService {
	d.init = true
	return d
}

func (d *blockingService) ServiceName() Name {
	if d.name == "" {
		// This is a nasty cheat, don't do it in any real code!
		return Name(fmt.Sprintf("blockingService-%p", d))
	}
	return d.name
}

func (d *blockingService) Run(ctx Context) error {
	// defer fmt.Println("dummy ENDED", d.ServiceName())
	// fmt.Println("RUNNING", d.ServiceName())

	if !d.init {
		panic("call Init()!")
	}
	if d.startDelay > 0 {
		time.Sleep(d.startDelay)
	}
	if d.startFailure != nil {
		return d.startFailure
	}
	if err := ctx.Ready(d); err != nil {
		return err
	}

	<-ctx.Halt()
	if d.haltDelay > 0 {
		time.Sleep(d.haltDelay)
	}
	return d.runFailure
}

// Testing tools copypasta

func WrapTB(tb testing.TB) T { tb.Helper(); return T{TB: tb} }

type T struct{ testing.TB }

const frameDepth = 2

func (tb T) MustFloatNear(epsilon float64, expected float64, actual float64, v ...interface{}) {
	tb.Helper()
	_ = tb.floatNear(true, epsilon, expected, actual, v...)
}

func (tb T) FloatNear(epsilon float64, expected float64, actual float64, v ...interface{}) bool {
	tb.Helper()
	return tb.floatNear(false, epsilon, expected, actual, v...)
}

func (tb T) floatNear(fatal bool, epsilon float64, expected float64, actual float64, v ...interface{}) bool {
	tb.Helper()
	near := IsFloatNear(epsilon, expected, actual)
	if !near {
		_, file, line, _ := runtime.Caller(frameDepth)
		msg := ""
		if len(v) > 0 {
			msg, v = v[0].(string), v[1:]
		}
		v = append([]interface{}{expected, actual, epsilon, filepath.Base(file), line}, v...)
		msg = fmt.Sprintf("\nfloat abs(%f - %f) > %f at %s:%d\n"+msg, v...)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
	}
	return near
}

// MustAssert immediately fails the test if the condition is false.
func (tb T) MustAssert(condition bool, v ...interface{}) {
	tb.Helper()
	_ = tb.assert(true, condition, v...)
}

// Assert fails the test if the condition is false.
func (tb T) Assert(condition bool, v ...interface{}) bool {
	tb.Helper()
	return tb.assert(false, condition, v...)
}

func (tb T) assert(fatal bool, condition bool, v ...interface{}) bool {
	tb.Helper()
	if !condition {
		_, file, line, _ := runtime.Caller(frameDepth)
		msg := ""
		if len(v) > 0 {
			msgx := v[0]
			v = v[1:]
			if msgx == nil {
				msg = "<nil>"
			} else if err, ok := msgx.(error); ok {
				msg = err.Error()
			} else {
				msg = msgx.(string)
			}
		}
		v = append([]interface{}{filepath.Base(file), line}, v...)
		msg = fmt.Sprintf("\nassertion failed at %s:%d\n"+msg, v...)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
	}
	return condition
}

// MustOK errors and terminates the test at the first error found in the arguments.
// It allows multiple return value functions to be passed in directly.
func (tb T) MustOK(errs ...interface{}) {
	tb.Helper()
	_ = tb.ok(true, errs...)
}

// OK errors the test at the first error found in the arguments, but continues
// running the test. It allows multiple return value functions to be passed in
// directly.
func (tb T) OK(errs ...interface{}) bool {
	tb.Helper()
	return tb.ok(false, errs...)
}

func (tb T) ok(fatal bool, errs ...interface{}) bool {
	tb.Helper()
	for _, err := range errs {
		if _, ok := err.(*testing.T); ok {
			panic("unexpected testing.T in call to OK()")
		} else if _, ok := err.(T); ok {
			panic("unexpected testtools.T in call to OK()")
		}
		if err, ok := err.(error); ok && err != nil {
			_, file, line, _ := runtime.Caller(frameDepth)
			msg := fmt.Sprintf("\nunexpected error at %s:%d\n%s", filepath.Base(file), line, err.Error())
			if fatal {
				tb.Fatal(msg)
			} else {
				tb.Error(msg)
			}
			return false
		}
	}
	return true
}

// MustExact immediately fails the test if exp is not equal to act.
func (tb T) MustExact(exp, act interface{}, v ...interface{}) {
	tb.Helper()
	_ = tb.exact(true, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) Exact(exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	return tb.exact(false, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) exact(fatal bool, exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	if exp != act {
		extra := ""
		if len(v) > 0 {
			extra = fmt.Sprintf(" - "+v[0].(string), v[1:]...)
		}

		_, file, line, _ := runtime.Caller(frameDepth)
		msg := CompareMsgf(exp, act, "\nexact failed at %s:%d%s", filepath.Base(file), line, extra)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
		return false
	}
	return true
}

// MustEqual immediately fails the test if exp is not equal to act based on
// reflect.DeepEqual()
func (tb T) MustEqual(exp, act interface{}, v ...interface{}) {
	tb.Helper()
	_ = tb.equals(true, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) Equals(exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	return tb.equals(false, exp, act, v...)
}

// Equal fails the test if exp is not equal to act.
func (tb T) equals(fatal bool, exp, act interface{}, v ...interface{}) bool {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		extra := ""
		if len(v) > 0 {
			extra = fmt.Sprintf(" - "+v[0].(string), v[1:]...)
		}

		_, file, line, _ := runtime.Caller(frameDepth)
		msg := CompareMsgf(exp, act, "\nequal failed at %s:%d%s", filepath.Base(file), line, extra)
		if fatal {
			tb.Fatal(msg)
		} else {
			tb.Error(msg)
		}
		return false
	}
	return true
}

func CompareMsg(exp, act interface{}) string {
	return fmt.Sprintf("\nexp: %+v\ngot: %+v", exp, act)
}

func CompareMsgf(exp, act interface{}, msg string, args ...interface{}) string {
	msg = fmt.Sprintf(msg, args...)
	return fmt.Sprintf("%v%v", msg, CompareMsg(exp, act))
}

func IsFloatNear(epsilon, expected, actual float64) bool {
	diff := expected - actual
	return diff == 0 || (diff < 0 && diff > -epsilon) || (diff > 0 && diff < epsilon)
}
