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
	threads  int
	waitRq   time.Duration
	waitConn time.Duration
}

func (sp *spammer) Flags(fs *cmdy.FlagSet) {
	fs.IntVar(&sp.total, "n", 100000, "total messages to send")
	fs.IntVar(&sp.threads, "c", 10, "threads")
	fs.DurationVar(&sp.waitRq, "wr", 0, "wait between requests")
	fs.DurationVar(&sp.waitConn, "wc", 0, "wait between connections")
}

func (sp *spammer) Config() *socketsrv.ConnectorConfig {
	config := socketsrv.DefaultConnectorConfig()
	config.Conn.ResponseTimeout = 20 * time.Second
	config.Conn.ReadTimeout = 20 * time.Second
	config.Conn.WriteTimeout = 20 * time.Second
	config.Conn.HeartbeatSendInterval = 60 * time.Second
	return config
}

type spammerClientCb func(handler socketsrv.Handler, opts ...socketsrv.ClientOption) (socketsrv.Client, error)

func (sp *spammer) Spam(ctx cmdy.Context, handler socketsrv.Handler, clientCb spammerClientCb) error {
	in := make([]byte, 10000)
	rand.Reader.Read(in)
	_ = in

	var wg sync.WaitGroup
	wg.Add(sp.threads)

	iter := sp.total / sp.threads
	left := sp.total % sp.threads

	s := time.Now()

	for thread := 0; thread < sp.threads; thread++ {
		citer := iter
		if thread == 0 {
			citer += left
		}
		if sp.waitConn > 0 {
			time.Sleep(sp.waitConn)
		}

		go func(thread int, iter int) {
			defer wg.Done()

			opts := []socketsrv.ClientOption{
				socketsrv.ClientDisconnect(func(connector *socketsrv.Connector, id socketsrv.ConnID, err error) {
					fmt.Println("disconnected:", id, err)
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
		}(thread, citer)
	}

	wg.Wait()

	since := time.Since(s)
	fmt.Println(since,
		since/time.Duration(sp.total),
		int64(sp.total)*int64(time.Second)/int64(since))

	return nil
}
