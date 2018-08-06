package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/pkg/profile"
	"github.com/shabbyrobe/cmdy"
	"github.com/shabbyrobe/cmdy/args"
	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/go-service/services"
	"github.com/shabbyrobe/go-service/serviceutil"
	"github.com/shabbyrobe/golib/iotools/bytewriter"
	"github.com/shabbyrobe/golib/socketsrv"
	"github.com/shabbyrobe/golib/socketsrv/wsocketsrv"
)

var proto Proto

func main() {
	if err := run(); err != nil {
		cmdy.Fatal(err)
	}
}

func run() error {
	bld := func() (cmdy.Command, error) {
		return cmdy.NewGroup("socketjunk", cmdy.Builders{
			"tcpclient": func() (cmdy.Command, error) { return &tcpClientCommand{}, nil },
			"tcpserver": func() (cmdy.Command, error) { return &tcpServerCommand{}, nil },
			"wsclient":  func() (cmdy.Command, error) { return &wsClientCommand{}, nil },
			"wsserver":  func() (cmdy.Command, error) { return &wsServerCommand{}, nil },
		}), nil
	}

	return cmdy.Run(context.Background(), os.Args[1:], bld)
}

type tcpClientCommand struct {
	host string
}

func (cl *tcpClientCommand) Synopsis() string     { return "tcpclient" }
func (cl *tcpClientCommand) Flags() *cmdy.FlagSet { return nil }

func (cl *tcpClientCommand) Args() *args.ArgSet {
	as := args.NewArgSet()
	as.StringOptional(&cl.host, "host", "localhost:9631", "host")
	return as
}

func (cl *tcpClientCommand) Run(ctx cmdy.Context) error {
	// defer profile.Start().Stop()

	config := socketsrv.ConnectorConfig{}
	connector := socketsrv.NewConnector(config, proto)

	h := &ServerHandler{}
	client, err := connector.NetClient(ctx, "tcp", cl.host, h)
	if err != nil {
		return err
	}

	s := time.Now()
	for i := 0; i < 10000; i++ {
		(client.Request(ctx, &TestRequest{}))
	}
	spew.Dump(time.Since(s))

	return nil
}

type tcpServerCommand struct {
	host string
}

func (sc *tcpServerCommand) Synopsis() string   { return "tcpserver" }
func (sc *tcpServerCommand) Args() *args.ArgSet { return nil }

func (sc *tcpServerCommand) Flags() *cmdy.FlagSet {
	fs := cmdy.NewFlagSet()
	fs.StringVar(&sc.host, "host", ":9631", "host")
	return fs
}

func (sc *tcpServerCommand) Run(ctx cmdy.Context) error {
	defer profile.Start().Stop()

	var config socketsrv.ServerConfig
	ln, err := socketsrv.Listen("tcp", sc.host)
	if err != nil {
		return err
	}

	handler := &ServerHandler{}

	srv := socketsrv.NewServer(config, ln, proto, handler)
	ender := service.NewEndListener(1)
	svc := service.New(service.Name(sc.host), srv).WithEndListener(ender)
	if err := service.StartTimeout(5*time.Second, services.Runner(), svc); err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", sc.host)

	return <-ender.Ends()
}

type wsClientCommand struct {
	url string
}

func (cl *wsClientCommand) Synopsis() string     { return "wsclient" }
func (cl *wsClientCommand) Flags() *cmdy.FlagSet { return nil }

func (cl *wsClientCommand) Args() *args.ArgSet {
	as := args.NewArgSet()
	as.StringOptional(&cl.url, "url", "ws://localhost:9632/", "host")
	return as
}

func (cl *wsClientCommand) Run(ctx cmdy.Context) error {
	// defer profile.Start().Stop()

	config := socketsrv.ConnectorConfig{}
	connector := socketsrv.NewConnector(config, proto)
	handler := &ServerHandler{}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	sock, _, err := dialer.Dial(cl.url, nil)
	if err != nil {
		return err
	}
	client, err := connector.Client(ctx, wsocketsrv.NewCommunicator(sock), handler)
	if err != nil {
		return err
	}

	s := time.Now()
	for i := 0; i < 10000; i++ {
		(client.Request(ctx, &TestRequest{}))
	}
	spew.Dump(time.Since(s))

	return client.Close()
}

type wsServerCommand struct {
	host string
}

func (sc *wsServerCommand) Synopsis() string   { return "wsserver" }
func (sc *wsServerCommand) Args() *args.ArgSet { return nil }

func (sc *wsServerCommand) Flags() *cmdy.FlagSet {
	fs := cmdy.NewFlagSet()
	fs.StringVar(&sc.host, "host", ":9632", "host")
	return fs
}

func (sc *wsServerCommand) Run(ctx cmdy.Context) error {
	defer profile.Start().Stop()

	var config socketsrv.ServerConfig
	ln := wsocketsrv.NewListener(websocket.Upgrader{})

	websrv := &http.Server{
		Addr:    sc.host,
		Handler: ln,
	}

	ender := service.NewEndListener(1)
	websvc := service.New("", serviceutil.NewHTTP(websrv))
	if err := service.StartTimeout(5*time.Second, services.Runner(), websvc); err != nil {
		return err
	}

	handler := &ServerHandler{}

	srv := socketsrv.NewServer(config, ln, proto, handler)
	svc := service.New(service.Name(sc.host), srv).WithEndListener(ender)
	if err := service.StartTimeout(5*time.Second, services.Runner(), svc); err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", sc.host)

	return <-ender.Ends()
}

type Proto struct{}

func (p Proto) Version() int         { return 1 }
func (p Proto) MessageLimit() uint32 { return 65536 }

func (p Proto) Message(kind int) (socketsrv.Message, error) {
	switch kind {
	case 1:
		return &TestRequest{}, nil
	case 2:
		return &TestResponse{}, nil
	case 3:
		return &TestCommand{}, nil
	case 4:
		return &OK{}, nil
	default:
		return nil, fmt.Errorf("unknown kind %d", kind)
	}
}

func (p Proto) MessageKind(msg socketsrv.Message) (int, error) {
	switch msg.(type) {
	case *TestRequest:
		return 1, nil
	case *TestResponse:
		return 2, nil
	case *TestCommand:
		return 3, nil
	case *OK:
		return 4, nil
	default:
		return 0, fmt.Errorf("unknown msg %T", msg)
	}
}

func (p Proto) Decode(in []byte) (env socketsrv.Envelope, rerr error) {
	var je JSONEnvelope
	if err := json.Unmarshal(in, &je); err != nil {
		return env, err
	}

	msg, err := p.Message(je.Kind)
	if err != nil {
		return env, err
	}

	if err := json.Unmarshal(je.Message, &msg); err != nil {
		return env, err
	}

	env.ID = je.ID
	env.ReplyTo = je.ReplyTo
	env.Kind = je.Kind
	env.Message = msg
	return env, nil
}

func (p Proto) Encode(env socketsrv.Envelope, into []byte) (extended []byte, rerr error) {
	var bw bytewriter.Writer
	bw.Give(into[:0])
	enc := json.NewEncoder(&bw)
	if err := enc.Encode(env.Message); err != nil {
		return nil, err
	}

	raw, n := bw.Take()
	je := JSONEnvelope{
		ID:      env.ID,
		ReplyTo: env.ReplyTo,
		Kind:    env.Kind,
		Message: raw[:n],
	}

	bw.Give(raw)
	enc = json.NewEncoder(&bw)
	if err := enc.Encode(je); err != nil {
		return nil, err
	}

	out, ilen := bw.Take()
	copy(out, out[n:])
	out = out[:ilen-n]

	return out, nil
}

type JSONEnvelope struct {
	ID      socketsrv.MessageID
	ReplyTo socketsrv.MessageID
	Kind    int
	Message json.RawMessage
}

type ServerHandler struct{}

func (h ServerHandler) HandleIncoming(id socketsrv.ConnID, msg socketsrv.Message) (rs socketsrv.Message, rerr error) {
	switch msg := msg.(type) {
	case *TestRequest:
		return &TestResponse{Bar: msg.Foo + ": yep!"}, nil
	default:
		return &OK{}, nil
	}
}

type OK struct{}

type TestRequest struct {
	Foo string
}

type TestResponse struct {
	Bar string
}

type TestCommand struct{}

type AltProto struct{}

func (p AltProto) Version() int         { return 1 }
func (p AltProto) MessageLimit() uint32 { return 65536 }

func (p AltProto) Message(kind int) (socketsrv.Message, error) {
	switch kind {
	case 1:
		return &TestRequest{}, nil
	case 2:
		return &TestResponse{}, nil
	case 3:
		return &TestCommand{}, nil
	case 4:
		return &OK{}, nil
	default:
		return nil, fmt.Errorf("unknown kind %d", kind)
	}
}

func (p AltProto) MessageKind(msg socketsrv.Message) (int, error) {
	switch msg.(type) {
	case *TestRequest:
		return 1, nil
	case *TestResponse:
		return 2, nil
	case *TestCommand:
		return 3, nil
	case *OK:
		return 4, nil
	default:
		return 0, fmt.Errorf("unknown msg %T", msg)
	}
}

func (p AltProto) Decode(in []byte) (env socketsrv.Envelope, rerr error) {
	env.ID = socketsrv.MessageID(binary.BigEndian.Uint32(in))
	env.ReplyTo = socketsrv.MessageID(binary.BigEndian.Uint32(in[4:]))
	env.Kind = int(binary.BigEndian.Uint32(in[8:]))
	env.Message, rerr = p.Message(env.Kind)
	return env, rerr
}

func (p AltProto) Encode(env socketsrv.Envelope, into []byte) (extended []byte, rerr error) {
	if len(into) < 12 {
		into = make([]byte, 12)
	}
	binary.BigEndian.PutUint32(into, uint32(env.ID))
	binary.BigEndian.PutUint32(into[4:], uint32(env.ReplyTo))
	binary.BigEndian.PutUint32(into[8:], uint32(env.Kind))
	return into[:12], nil
}
