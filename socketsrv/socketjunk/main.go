package main

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/binary"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
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
	"github.com/shabbyrobe/golib/socketsrv/jsonsrv"
	"github.com/shabbyrobe/golib/socketsrv/packetsrv"
	"github.com/shabbyrobe/golib/socketsrv/wsocketsrv"
)

var encoding = binary.BigEndian

func main() {
	if err := run(); err != nil {
		cmdy.Fatal(err)
	}
}

func run() error {
	bld := func() (cmdy.Command, cmdy.Init) {
		return cmdy.NewGroup("socketjunk", cmdy.Builders{
			"tcpclient": func() (cmdy.Command, cmdy.Init) { return &tcpClientCommand{}, nil },
			"tcpserver": func() (cmdy.Command, cmdy.Init) { return &tcpServerCommand{}, nil },
			"pktclient": func() (cmdy.Command, cmdy.Init) { return &pktClientCommand{}, nil },
			"pktserver": func() (cmdy.Command, cmdy.Init) { return &pktServerCommand{}, nil },
			"wsclient":  func() (cmdy.Command, cmdy.Init) { return &wsClientCommand{}, nil },
			"wsserver":  func() (cmdy.Command, cmdy.Init) { return &wsServerCommand{}, nil },
		}), nil
	}

	return cmdy.Run(context.Background(), os.Args[1:], bld)
}

func negotiatorBuild() socketsrv.Negotiator {
	negotiator := socketsrv.NewVersionNegotiator(
		Proto{version: 1, messageLimit: 1000000, mapper: Mapper{}, codec: SimpleCodec{}},
		Proto{version: 2, messageLimit: 1000000, mapper: Mapper{}, codec: jsonsrv.Codec},
		// CompressingProto{},
	)
	return negotiator
}

type tcpClientCommand struct {
	host    string
	spammer spammer
}

func (cl *tcpClientCommand) Synopsis() string { return "tcpclient" }
func (cl *tcpClientCommand) Flags() *cmdy.FlagSet {
	fs := cmdy.NewFlagSet()
	cl.spammer.Flags(fs)
	return fs
}

func (cl *tcpClientCommand) Args() *args.ArgSet {
	as := args.NewArgSet()
	as.StringOptional(&cl.host, "host", "localhost:9631", "host")
	return as
}

func (cl *tcpClientCommand) Run(ctx cmdy.Context) error {
	// defer profile.Start().Stop()
	debugServer(":14440")

	dialer := cl.spammer.Dialer(negotiatorBuild())

	clientCb := func(handler socketsrv.Handler, opts ...socketsrv.ClientOption) (socketsrv.Client, error) {
		client, err := dialer.DialStream(ctx, "tcp", cl.host, handler, opts...)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return cl.spammer.Spam(ctx, nil, clientCb)
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
	// defer profile.Start(profile.BlockProfile).Stop()

	debugServer(":14441")

	handler := &ServerHandler{}

	var config = socketsrv.DefaultServerConfig()
	config.Conn.ResponseTimeout = 20 * time.Second
	config.Conn.ReadTimeout = 20 * time.Second
	config.Conn.WriteTimeout = 20 * time.Second

	ln, err := socketsrv.ListenStream("tcp", sc.host)
	if err != nil {
		return err
	}

	srv := socketsrv.NewServer(config, ln, negotiatorBuild(), handler,
		socketsrv.ServerConnect(func(srv *socketsrv.Server, id socketsrv.ConnID) {
			fmt.Println("connect", id)
		}),
		socketsrv.ServerDisconnect(func(srv *socketsrv.Server, id socketsrv.ConnID, err error) {
			fmt.Println("disconnect", id, err)
		}),
	)

	go srv.Serve()

	fmt.Printf("listening on %s\n", sc.host)

	select {}
}

type pktClientCommand struct {
	host string
}

func (cl *pktClientCommand) Synopsis() string     { return "pktclient" }
func (cl *pktClientCommand) Flags() *cmdy.FlagSet { return nil }

func (cl *pktClientCommand) Args() *args.ArgSet {
	as := args.NewArgSet()
	as.StringOptional(&cl.host, "host", "localhost:9633", "host")
	return as
}

func (cl *pktClientCommand) Run(ctx cmdy.Context) error {
	socketDialer := socketsrv.DefaultDialer(negotiatorBuild())

	dialer := net.Dialer{}
	handler := &ServerHandler{}

	in, err := ioutil.ReadFile("/Users/bl/Downloads/The Rust Programming Language.htm")
	if err != nil {
		return err
	}
	in = in[:50]
	_ = in

	s := time.Now()

	iter := 200
	threads := 1000
	var wg sync.WaitGroup
	wg.Add(threads)

	for thread := 0; thread < threads; thread++ {
		go func(thread int) {
			defer wg.Done()

			conn, err := dialer.DialContext(ctx, "udp", cl.host)
			if err != nil {
				panic(err)
			}
			client, err := socketDialer.Client(ctx, packetsrv.ClientCommunicator(conn), handler)
			if err != nil {
				panic(err)
			}

			rq := &TestRequest{
				Foo: fmt.Sprintf("%d", thread),
			}
			for i := 0; i < iter; i++ {
				rsp, err := (client.Request(ctx, rq))
				if err != nil {
					panic(err)
				}
				_ = rsp
				// fmt.Printf("%#v\n", rsp)
			}
		}(thread)
	}

	wg.Wait()
	spew.Dump(time.Since(s))

	return nil
}

type pktServerCommand struct {
	host string
}

func (sc *pktServerCommand) Synopsis() string   { return "pktserver" }
func (sc *pktServerCommand) Args() *args.ArgSet { return nil }

func (sc *pktServerCommand) Flags() *cmdy.FlagSet {
	fs := cmdy.NewFlagSet()
	fs.StringVar(&sc.host, "host", ":9633", "host")
	return fs
}

func (sc *pktServerCommand) Run(ctx cmdy.Context) error {
	defer profile.Start().Stop()

	debugServer(":14442")

	handler := &ServerHandler{}

	ln, err := packetsrv.Listen("udp", sc.host)
	if err != nil {
		return err
	}

	srv := socketsrv.NewServer(nil, ln, negotiatorBuild(), handler)
	ender := service.NewEndListener(1)
	svc := service.New(service.Name(sc.host), srv).WithEndListener(ender)
	if err := service.StartTimeout(5*time.Second, services.Runner(), svc); err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", sc.host)

	return <-ender.Ends()
}

type wsClientCommand struct {
	url     string
	spammer spammer
}

func (cl *wsClientCommand) Synopsis() string { return "wsclient" }
func (cl *wsClientCommand) Flags() *cmdy.FlagSet {
	fs := cmdy.NewFlagSet()
	cl.spammer.Flags(fs)
	return fs
}

func (cl *wsClientCommand) Args() *args.ArgSet {
	as := args.NewArgSet()
	as.StringOptional(&cl.url, "url", "ws://localhost:9632/", "host")
	return as
}

func (cl *wsClientCommand) Run(ctx cmdy.Context) error {
	// defer profile.Start().Stop()

	dialer := cl.spammer.Dialer(negotiatorBuild())

	clientCb := func(handler socketsrv.Handler, opts ...socketsrv.ClientOption) (socketsrv.Client, error) {
		wsDialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}
		sock, _, err := wsDialer.Dial(cl.url, nil)
		if err != nil {
			return nil, err
		}
		client, err := dialer.Client(ctx, wsocketsrv.NewCommunicator(sock), handler, opts...)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return cl.spammer.Spam(ctx, nil, clientCb)
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

	srv := socketsrv.NewServer(nil, ln, negotiatorBuild(), handler,
		socketsrv.ServerConnect(func(srv *socketsrv.Server, id socketsrv.ConnID) {
			fmt.Println("connect", id)
		}),
		socketsrv.ServerDisconnect(func(srv *socketsrv.Server, id socketsrv.ConnID, err error) {
			fmt.Println("disconnect", id, err)
		}),
	)
	svc := service.New(service.Name(sc.host), srv).WithEndListener(ender)
	if err := service.StartTimeout(5*time.Second, services.Runner(), svc); err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", sc.host)

	return <-ender.Ends()
}

type VersionNegotiator struct {
	Protocols map[uint32]socketsrv.Protocol
	Timeout   time.Duration
}

func (v *VersionNegotiator) Negotiate(side socketsrv.Side, c socketsrv.Communicator) (socketsrv.Protocol, error) {
	timeout := v.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var oursBuf = make([]byte, len(v.Protocols)*4)
	i := 0
	for v := range v.Protocols {
		encoding.PutUint32(oursBuf[i:], v)
		i += 4
	}
	if err := c.WriteMessage(oursBuf, timeout); err != nil {
		return nil, err
	}

	msg, err := c.ReadMessage(nil, 1024, timeout)
	if err != nil {
		return nil, err
	}
	if len(msg)%4 != 0 {
		return nil, fmt.Errorf("unexpected remote versions")
	}

	var max uint32
	var found bool
	for i := 0; i < len(msg); i += 4 {
		cur := encoding.Uint32(msg[i:])
		if _, ok := v.Protocols[cur]; ok {
			found = true
			if cur > max {
				max = cur
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("could not negotiate protocol")
	}

	return v.Protocols[max], nil
}

type SimpleCodec struct{}

var _ socketsrv.Codec = SimpleCodec{}

func (p SimpleCodec) Decode(in []byte, mapper socketsrv.Mapper, decdata *socketsrv.ProtoData) (env socketsrv.Envelope, rerr error) {
	var je JSONEnvelope
	if err := json.Unmarshal(in, &je); err != nil {
		return env, err
	}

	msg, err := mapper.Message(je.Kind)
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

func (p SimpleCodec) Encode(env socketsrv.Envelope, into []byte, encdata *socketsrv.ProtoData) (extended []byte, rerr error) {
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

func (h ServerHandler) HandleRequest(ctx context.Context, in socketsrv.IncomingRequest) (rs socketsrv.Message, rerr error) {
	switch msg := in.Message.(type) {
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

type Mapper struct{}

func (m Mapper) Message(kind int) (socketsrv.Message, error) {
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

func (m Mapper) MessageKind(msg socketsrv.Message) (int, error) {
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

type Proto struct {
	version      int
	name         string
	mapper       socketsrv.Mapper
	codec        socketsrv.Codec
	messageLimit int
}

func (p Proto) ProtocolName() string     { return p.name }
func (p Proto) Mapper() socketsrv.Mapper { return p.mapper }
func (p Proto) Codec() socketsrv.Codec   { return p.codec }
func (p Proto) Version() int             { return p.version }
func (p Proto) MessageLimit() int        { return p.messageLimit }

type CompressingCodec struct{}

var _ socketsrv.Codec = &CompressingCodec{}

func (p CompressingCodec) Decode(in []byte, mapper socketsrv.Mapper, decData *socketsrv.ProtoData) (env socketsrv.Envelope, rerr error) {
	var data *altProtoData
	if *decData == nil {
		data = &altProtoData{}
		*decData = data
	} else {
		data = (*decData).(*altProtoData)
	}

	if len(in) < 13 {
		return env, fmt.Errorf("short message")
	}
	env.ID = socketsrv.MessageID(encoding.Uint32(in))
	env.ReplyTo = socketsrv.MessageID(encoding.Uint32(in[4:]))
	env.Kind = int(encoding.Uint32(in[8:]))
	env.Message, rerr = mapper.Message(env.Kind)

	compressed := in[12]&0x1 == 0x1
	msg := in[13:]
	if compressed {
		dr := flate.NewReader(bytes.NewReader(msg))
		var bw bytewriter.Writer
		var n int
		bw.Give(data.scratch[:0])
		if _, err := io.Copy(&bw, dr); err != nil {
			return env, err
		}
		msg, n = bw.Take()
		msg = msg[:n]
	}

	if err := json.Unmarshal(msg, &env.Message); err != nil {
		return env, err
	}

	return env, nil
}

func (p CompressingCodec) Encode(env socketsrv.Envelope, into []byte, encData *socketsrv.ProtoData) (extended []byte, rerr error) {
	var data *altProtoData
	if *encData == nil {
		fw, err := flate.NewWriter(ioutil.Discard, 9)
		if err != nil {
			return into, err
		}
		data = &altProtoData{
			flateWriter: fw,
		}
		*encData = data
	} else {
		data = (*encData).(*altProtoData)
	}

	if len(into) < 13 {
		into = make([]byte, 13)
	}
	encoding.PutUint32(into, uint32(env.ID))
	encoding.PutUint32(into[4:], uint32(env.ReplyTo))
	encoding.PutUint32(into[8:], uint32(env.Kind))
	into[12] = 0

	var bw bytewriter.Writer
	bw.Give(into[:13])

	enc := json.NewEncoder(&bw)
	if err := enc.Encode(env.Message); err != nil {
		return into, err
	}

	out, ilen := bw.Take()

	if ilen >= 5000 {
		into[12] |= 0x01

		if len(data.scratch) < 13 {
			data.scratch = make([]byte, 13)
		}
		copy(data.scratch, into[:13])

		bw.Give(data.scratch[:13])
		cw := data.flateWriter
		cw.Reset(&bw)
		if w, err := cw.Write(out[13:ilen]); err != nil {
			_ = cw.Close()
			return into, err
		} else if w != ilen-13 {
			_ = cw.Close()
			return into, fmt.Errorf("short write")
		}
		if err := cw.Flush(); err != nil {
			_ = cw.Close()
			return into, err
		}
		if err := cw.Close(); err != nil {
			_ = cw.Close()
			return into, err
		}

		out, ilen = bw.Take()
	}
	return out[:ilen], nil
}

type altProtoData struct {
	scratch     []byte
	flateWriter *flate.Writer
}

func (a *altProtoData) Close() error {
	return nil
}

func debugServer(host string) {
	mux := http.NewServeMux()
	mux.Handle("/debug/vars", expvar.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	hsrv := &http.Server{Addr: host}
	hsrv.Handler = mux
	hsvc := &serviceutil.HTTP{Server: hsrv}
	svc := service.New("", hsvc)
	if err := service.StartTimeout(10*time.Second, services.Runner(), svc); err != nil {
		panic(err)
	}
	fmt.Println("debug server running on", hsvc.Port())
}
