package logtools

import "log"

var _ Logger = &log.Logger{}
