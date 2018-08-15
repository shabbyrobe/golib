package socketsrv

import "errors"

var (
	errConnShutdown       = errors.New("socketsrv: connection shutdown")
	errResponseTimeout    = errors.New("socketsrv: response timeout")
	errConnSendNotRunning = errors.New("socketsrv: send to conn which is not running")
	errReadTimeout        = errors.New("socketsrv: read timeout")
)
