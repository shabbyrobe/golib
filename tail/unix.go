package tail

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"syscall"
)

type Unix struct {
	prev       int
	persistent bool
	closed     atomic.Int64
	file       string
	bin        string
	cmd        *exec.Cmd
	rdr        io.ReadCloser
	line       string
	lines      chan string
	done       chan error
	err        error
	stop       chan struct{}
}

type UnixOption func(u *Unix)

func UnixPrev(prev int) UnixOption {
	return func(u *Unix) { u.prev = prev }
}

func NewUnix(file string, opts ...UnixOption) (*Unix, error) {
	if file == "" {
		return nil, fmt.Errorf("empty file name")
	}
	if file == "-" {
		return nil, fmt.Errorf("cannot follow stdin")
	}

	u := &Unix{
		file:       file,
		persistent: true,
		prev:       10,
		lines:      make(chan string, 1),
		done:       make(chan error, 1),
		stop:       make(chan struct{}),
	}
	for _, opt := range opts {
		opt(u)
	}
	if u.bin == "" {
		u.bin = "tail"
	}
	look, err := exec.LookPath(u.bin)
	if err != nil {
		return nil, err
	}
	u.bin = look

	args := []string{
		"-n", strconv.Itoa(u.prev),
	}
	if u.persistent {
		args = append(args, "-F")
	} else {
		args = append(args, "-f")
	}
	args = append(args, u.file)

	var errbuf bytes.Buffer
	u.cmd = exec.Command(u.bin, args...)
	u.cmd.Stderr = &errbuf
	u.rdr, err = u.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := u.cmd.Start(); err != nil {
		u.rdr.Close()
		return nil, err
	}

	go func() {
		defer close(u.lines)
		scn := bufio.NewScanner(u.rdr)
		for scn.Scan() {
			select {
			case u.lines <- scn.Text():
			case <-u.stop:
			}
		}
	}()

	go func() {
		select {
		case u.done <- u.cmd.Wait():
		default:
		}
	}()

	return u, nil
}

func (u *Unix) Close() error {
	if !u.closed.CompareAndSwap(0, 1) {
		return u.err
	}

	// Do this before closing the stdout reader
	terminate(u.cmd.Process)

	u.rdr.Close()

	close(u.stop)
	err := <-u.done
	if isErrTerminated(err) {
		err = nil
	}
	u.err = err
	return u.err
}

func (u *Unix) Scan() bool {
	line, ok := <-u.lines
	if ok {
		u.line = line
		return true
	} else {
		return false
	}
}

func (u *Unix) Line() string {
	return u.line
}

func terminate(proc *os.Process) {
	proc.Signal(syscall.SIGTERM)
}

func isErrTerminated(err error) bool {
	exErr := (*exec.ExitError)(nil)
	if !errors.As(err, &exErr) {
		return false
	}

	sys := exErr.Sys()
	waitStatus, ok := sys.(syscall.WaitStatus)
	if !ok {
		return false
	}

	return waitStatus.Signal() == syscall.SIGTERM
}
