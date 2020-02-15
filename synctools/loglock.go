package synctools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Mutex = sync.Mutex
type RWMutex = sync.RWMutex

// type Mutex = LoggingMutex
// type RWMutex = LoggingRWMutex

type RWLocker interface {
	sync.Locker

	RLocker() sync.Locker
	RLock()
	RUnlock()
}

var (
	LoggingMutexWriter io.Writer = os.Stdout

	LogFullStack = true

	next uint64
)

type LoggingMutex struct {
	sync.Mutex
}

func (l *LoggingMutex) Lock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "lock")
	l.Mutex.Lock()
	wlog(unsafe.Pointer(l), id, "locked")
}

func (l *LoggingMutex) Unlock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "unlock")
	l.Mutex.Unlock()
	wlog(unsafe.Pointer(l), id, "unlocked")
}

type LoggingRWMutex struct {
	sync.RWMutex
}

func (l *LoggingRWMutex) Lock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "lock")
	l.RWMutex.Lock()
	wlog(unsafe.Pointer(l), id, "locked")
}

func (l *LoggingRWMutex) Unlock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "unlock")
	l.RWMutex.Unlock()
	wlog(unsafe.Pointer(l), id, "unlocked")
}

func (l *LoggingRWMutex) RLock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "rlock")
	l.RWMutex.RLock()
	wlog(unsafe.Pointer(l), id, "rlocked")
}

func (l *LoggingRWMutex) RUnlock() {
	id := atomic.AddUint64(&next, 1)
	wlog(unsafe.Pointer(l), id, "runlock")
	l.RWMutex.RUnlock()
	wlog(unsafe.Pointer(l), id, "runlocked")
}

func wlog(p unsafe.Pointer, id uint64, event string) {
	n := time.Now()
	tm := n.Format("2006-01-02T15:04:05.") // .999999999Z07:00"
	tm += rightPad(strconv.FormatInt(int64(n.Nanosecond()), 10), '0', 9)
	tm += n.Format("Z07:00")

	_, file, line, _ := runtime.Caller(2)
	prefix := fmt.Sprintf("%s %p %d %10s ", tm, p, id, event)
	if !LogFullStack {
		fmt.Fprintf(LoggingMutexWriter, "%s%s:%d\n", prefix, file, line)
	} else {
		indent := strings.Repeat(" ", len(prefix))
		st := make([]uintptr, 5)
		_ = runtime.Callers(3, st)
		cf := runtime.CallersFrames(st)
		i := 0

		var buf bytes.Buffer
		for {
			f, more := cf.Next()
			if !more {
				break
			}
			if f.Func != nil {
				fmt.Fprintf(&buf, "%s%s:%d\n", prefix, f.File, f.Line)
				if i == 0 {
					prefix = indent
				}
				i++
			}
		}
		buf.WriteByte('\n')
		buf.WriteTo(LoggingMutexWriter)
	}
}

func rightPad(s string, c byte, total int) string {
	pad := total - len(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(string(c), pad)
}
