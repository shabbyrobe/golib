package nettools

import (
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestHostDefaultPort(t *testing.T) {
	tt := assert.WrapTB(t)

	tt.MustEqual("localhost:123", HostDefaultPort("localhost", "123"))
	tt.MustEqual("localhost:123", HostDefaultPort("localhost", ":123"))
	tt.MustEqual("localhost:123", HostDefaultPort("localhost:123", ""))
	tt.MustEqual("localhost:123", HostDefaultPort("localhost:123", "456"))
	tt.MustEqual("localhost", HostDefaultPort("localhost:", ""))
	tt.MustEqual("localhost:", HostDefaultPort("localhost:", ":"))
	tt.MustEqual("localhost:http", HostDefaultPort("localhost", ":http"))
	tt.MustEqual("localhost:http", HostDefaultPort("localhost:", ":http"))
	tt.MustEqual("localhost:http", HostDefaultPort("localhost:", "http"))
}
