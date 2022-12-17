// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This exists because of this issue:
// https://github.com/golang/go/issues/18482
//
// The code has been copy/pasted from crypto/tls and from this as-yet
// unmerged patch:
// https://go-review.googlesource.com/c/93255/

package tlsdial

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"time"
)

var emptyConfig tls.Config

func defaultConfig() *tls.Config {
	return &emptyConfig
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "tls: DialWithDialer timed out" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// DialContextWithDialer connects to the given network address using
// dialer.DialContext with the provided Context and then initiates a TLS
// handshake, returning the resulting TLS connection.
//
// Any timeout or deadline given in the dialer apply to connection and
// TLS handshake as a whole.
//
// The provided Context must be non-nil. If the context expires before
// the connection and TLS handshake are complete, an error is returned.
// Once TLS handshake completes successfully, any expiration of the
// context will not affect the connection.
//
// DialContextWithDialer interprets a nil configuration as equivalent to
// the zero configuration; see the documentation of Config for the defaults.
//
// Deprecated: supported in stdlib since 1.15.
func DialContextWithDialer(ctx context.Context, dialer *net.Dialer, network, addr string, config *tls.Config) (*tls.Conn, error) {
	// We want the Timeout and Deadline values from dialer to cover the
	// whole process: TCP connection and TLS handshake. This means that we
	// also need to start our own timers now.
	timeout := dialer.Timeout
	if !dialer.Deadline.IsZero() {
		deadlineTimeout := time.Until(dialer.Deadline)
		if timeout == 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}
	if timeout != 0 {
		subCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = subCtx
	}
	var errChannel chan error
	if ctx != context.Background() {
		errChannel = make(chan error, 2)

		subCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			<-subCtx.Done()
			switch subCtx.Err() {
			case context.DeadlineExceeded:
				errChannel <- timeoutError{}
			default:
				errChannel <- subCtx.Err()
			}
		}()
		ctx = subCtx
	}

	rawConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]
	if config == nil {
		config = defaultConfig()
	}

	// If no ServerName is set, infer the ServerName
	// from the hostname we're connecting to.
	if config.ServerName == "" {
		// Make a copy to avoid polluting argument or default.
		c := config.Clone()
		c.ServerName = hostname
		config = c
	}
	conn := tls.Client(rawConn, config)
	if ctx == context.Background() {
		err = conn.Handshake()
	} else {
		go func() {
			errChannel <- conn.Handshake()
		}()
		err = <-errChannel
	}
	if err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}

// DialWithDialer connects to the given network address using dialer.Dial and
// then initiates a TLS handshake, returning the resulting TLS connection. Any
// timeout or deadline given in the dialer apply to connection and TLS
// handshake as a whole.
//
// DialWithDialer interprets a nil configuration as equivalent to the zero
// configuration; see the documentation of Config for the defaults.
//
// Deprecated: supported in stdlib since 1.15.
func DialWithDialer(dialer *net.Dialer, network, addr string, config *tls.Config) (*tls.Conn, error) {
	return DialContextWithDialer(context.Background(), dialer, network, addr, config)
}
