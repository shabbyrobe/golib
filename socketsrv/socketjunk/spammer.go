package main

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/shabbyrobe/cmdy"
	"github.com/shabbyrobe/golib/socketsrv"
)

type spammer struct {
	total    int
	conns    int
	waitRq   time.Duration
	waitConn time.Duration
}

func (sp *spammer) Flags(fs *cmdy.FlagSet) {
	fs.IntVar(&sp.total, "n", 100000, "total messages to send")
	fs.IntVar(&sp.conns, "c", 10, "connections")
	fs.DurationVar(&sp.waitRq, "wr", 0, "wait between requests")
	fs.DurationVar(&sp.waitConn, "wc", 0, "wait between connections")
}

func (sp *spammer) Dialer(neg socketsrv.Negotiator) socketsrv.Dialer {
	config := socketsrv.DefaultDialer(neg)
	config.ResponseTimeout = 20 * time.Second
	config.ReadTimeout = 20 * time.Second
	config.WriteTimeout = 20 * time.Second
	config.HeartbeatSendInterval = 60 * time.Second
	return config
}

type spammerClientCb func(handler socketsrv.Handler, opts ...socketsrv.ClientOption) (socketsrv.Client, error)

func (sp *spammer) Spam(ctx cmdy.Context, handler socketsrv.Handler, clientCb spammerClientCb) error {
	in := make([]byte, 10000)
	rand.Reader.Read(in)
	_ = in

	var wgt, wgConn sync.WaitGroup
	wgt.Add(sp.conns)

	iter := sp.total / sp.conns
	left := sp.total % sp.conns

	s := time.Now()

	for conn := 0; conn < sp.conns; conn++ {
		citer := iter
		if conn == 0 {
			citer += left
		}
		if sp.waitConn > 0 {
			time.Sleep(sp.waitConn)
		}

		go func(thread int, iter int) {
			defer wgt.Done()

			opts := []socketsrv.ClientOption{
				socketsrv.ClientConnect(func(id socketsrv.ConnID) {
					wgConn.Add(1)
				}),
				socketsrv.ClientDisconnect(func(id socketsrv.ConnID, err error) {
					fmt.Println("disconnected:", id, err)
					wgConn.Done()
				}),
			}

			client, err := clientCb(handler, opts...)
			if err != nil {
				fmt.Println("thread", thread, "failed:", err)
				return
			}

			defer client.Close()

			rq := &TestRequest{
				Foo: fmt.Sprintf("%d", thread),
			}
			rs := make(chan socketsrv.Result, 1)
			for i := 0; i < iter; i++ {
				err := (client.Send(ctx, rq, rs))
				if err != nil {
					fmt.Println("thread", thread, "failed:", err)
					return
				}
				rsp := <-rs
				_ = rsp
				if sp.waitRq > 0 {
					time.Sleep(sp.waitRq)
				}
				// fmt.Printf("%#v\n", rsp)
			}
		}(conn, citer)
	}

	wgt.Wait()
	wgConn.Wait()

	since := time.Since(s)
	fmt.Println(since,
		since/time.Duration(sp.total),
		int64(sp.total)*int64(time.Second)/int64(since))

	return nil
}
