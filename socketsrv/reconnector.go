package socketsrv

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ReconnectDelayer func(err error) time.Duration

func ReconnectEvery(d time.Duration) ReconnectDelayer {
	return func(err error) time.Duration { return d }
}

// Reconnector is a Client that will re-establish a connection in a loop after
// it is broken.
type Reconnector struct {
	dial ReconnectorDialFunc

	client  Client
	mu      sync.RWMutex
	stop    chan struct{}
	stopped bool
	running bool
	delayer ReconnectDelayer
}

type ReconnectorDialFunc func(ctx context.Context, dc OnClientDisconnect) (Client, error)

var _ Client = &Reconnector{}

func NewReconnector(delayer ReconnectDelayer, dial ReconnectorDialFunc) *Reconnector {
	if delayer == nil {
		panic("delayer was nil")
	}
	if dial == nil {
		panic("dial was nil")
	}
	return &Reconnector{
		dial:    dial,
		delayer: delayer,
		stop:    make(chan struct{}),
	}
}

func (rc *Reconnector) Close() error {
	rc.mu.Lock()
	if rc.stopped {
		rc.mu.Unlock()
		return fmt.Errorf("socketsrv: reconnector already stopped")
	}

	if !rc.running {
		rc.mu.Unlock()
		return fmt.Errorf("socketsrv: client not running")
	}

	rc.stopped = true
	close(rc.stop)
	var err error
	if rc.client != nil {
		err = rc.client.Close()
	}
	rc.mu.Unlock()
	return err
}

// ID returns the ConnID of the underlying connection. This will change as the
// client reconnects.
func (rc *Reconnector) ID() (out ConnID) {
	rc.mu.RLock()
	if rc.client != nil {
		out = rc.client.ID()
	}
	rc.mu.RUnlock()
	return out
}

func (rc *Reconnector) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	rc.mu.RLock()
	if rc.stopped {
		rc.mu.RUnlock()
		return errNotRunning
	}
	client := rc.client
	rc.mu.RUnlock()
	if client != nil {
		return client.Send(ctx, msg, recv)
	}
	return errNotRunning
}

func (rc *Reconnector) Request(ctx context.Context, msg Message) (resp Message, rerr error) {
	rc.mu.RLock()
	if rc.stopped {
		rc.mu.RUnlock()
		return nil, errNotRunning
	}
	client := rc.client
	rc.mu.RUnlock()
	if client != nil {
		return client.Request(ctx, msg)
	}
	return nil, fmt.Errorf("socketsrv: client not available")
}

func (rc *Reconnector) Dial(ctx context.Context) error {
	{
		rc.mu.Lock()
		if rc.running {
			rc.mu.Unlock()
			return errAlreadyRunning
		}
		rc.running = true
		rc.mu.Unlock()
	}

	errc := make(chan error, 1)

	{ // Setup: initiate the first connection, then background ourselves when
		// it succeeds or fails.
		client, err := rc.dial(ctx, func(id ConnID, err error) {
			errc <- err
		})
		if err != nil {
			errc <- err
		}
		rc.mu.Lock()
		rc.client = client
		rc.mu.Unlock()
	}

	go rc.run(errc)

	return nil
}

func (rc *Reconnector) run(errc chan error) {
	// 'next' allows the reactor thread to unblock the connector thread when
	// it is time to attempt a connection:
	var next = make(chan struct{}, 1)
	var done = make(chan struct{})
	var cctx = &chanContext{done}

	var wg sync.WaitGroup
	wg.Add(1)

	// Shutdown:
	defer func() {
		close(done)
		_ = rc.Close()
		wg.Wait()

		rc.mu.Lock()
		rc.running = false
		rc.mu.Unlock()
	}()

	// Connector thread. This will establish connections with the remote, and
	// when they succeed, will block and wait for the signal from the reactor
	// thread to attempt to connect again or stop.
	go func() {
		defer wg.Done()

		for {
			select {
			case _, ok := <-next:
				if !ok {
					return
				}
			case <-done:
				return
			}

			// OnClientDisconnect will only receive an error if dial does not return
			// an error; we are guaranteed only one send to errc per iteration of the
			// connector:
			client, err := rc.dial(cctx, func(id ConnID, err error) {
				select {
				case errc <- err:
				case <-done:
				}
			})
			rc.mu.Lock()
			rc.client = client
			rc.mu.Unlock()

			if err != nil {
				select {
				case errc <- err:
				case <-done:
					return
				}
				continue
			}
		}
	}()

	{ // Reactor. This collects signals from all the disparate sources and makes
		// decisions about when to connect:
		for {
			// If the initial connection failed, errc will already have an error waiting
			// for us in it, otherwise we should assume the connection is running.
			select {
			case err := <-errc:
				delay := rc.delayer(err)
				if halted := rc.sleep(delay); halted {
					return

				}
				select {
				case next <- struct{}{}:
				case <-cctx.Done():
					return
				}

			case <-cctx.Done():
				return

			case <-rc.stop:
				close(next)
				return
			}
		}
	}
}

func (rc *Reconnector) sleep(d time.Duration) (halted bool) {
	// minHaltableSleep is a performance hack. It's probably not a
	// one-size-fits all constant but it'll do for now.
	if d < 50*time.Millisecond {
		time.Sleep(d)
		select {
		case <-rc.stop:
			return true
		default:
			return false
		}
	}
	select {
	case <-time.After(d):
		return false
	case <-rc.stop:
		return true
	}
}
