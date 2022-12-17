package curl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type HTTPVersion string

const (
	HTTP10              HTTPVersion = "--http1.0"
	HTTP11              HTTPVersion = "--http1.1"
	HTTP2               HTTPVersion = "--http2"
	HTTP2PriorKnowledge HTTPVersion = "--http2-prior-knowledge"
	HTTP3               HTTPVersion = "--http3"
)

type Network string

const (
	NetIPv4 Network = "--ipv4"
	NetIPv6 Network = "--ipv6"
)

type Transport struct {
	CurlBin            string
	Version            HTTPVersion
	Insecure           bool
	Network            Network
	DisableCompression bool

	// --limit-rate
	// --path-as-is
}

var _ http.RoundTripper = &Transport{}

func (rt *Transport) RoundTrip(rq *http.Request) (*http.Response, error) {
	var rdr io.Reader = rq.Body
	if rq.Body != nil {
		defer rq.Body.Close()
	}

	var rs http.Response
	var err error

	bin := rt.CurlBin
	if bin == "" {
		bin, err = exec.LookPath("curl")
		if err != nil {
			return nil, err
		}
	}

	var args = []string{
		"--include",
		"--silent",
		"--raw",
		"--buffer",
		"--show-error",
		"--suppress-connect-headers",
		"--no-styled-output",
		"--output", "-",
	}

	if rt.Insecure {
		args = append(args, "--insecure")
	}

	switch rt.Version {
	case "":
	case HTTP10, HTTP11, HTTP2, HTTP2PriorKnowledge, HTTP3:
		args = append(args, string(rt.Version))
	default:
		return nil, fmt.Errorf("unknown version flag %q", rt.Version)
	}

	switch rt.Network {
	case "":
	case NetIPv4, NetIPv6:
		args = append(args, string(rt.Network))
	default:
		return nil, fmt.Errorf("unknown network flag %q", rt.Network)
	}

	switch rq.Method {
	case http.MethodHead:
		args = append(args, "--head")
	case http.MethodConnect, http.MethodTrace:
		return nil, fmt.Errorf("unsupported method %q", rq.Method)
	case http.MethodPost, http.MethodDelete, http.MethodPut, http.MethodPatch, http.MethodOptions:
		args = append(args, "-X", rq.Method)
	default:
		// FIXME: Allow?
		args = append(args, "-X", rq.Method)
	}

	if rq.Host != "" {
		if hv, ok := curlHeader("Host", rq.Host); ok {
			args = append(args, "-H", hv)
		}
	}

	requestedGzip := false
	if !rt.DisableCompression &&
		rq.Header.Get("Accept-Encoding") == "" &&
		rq.Header.Get("Range") == "" &&
		rq.Method != "HEAD" {
		// Note that we don't request this for HEAD requests,
		// due to a bug in nginx:
		//   https://trac.nginx.org/nginx/ticket/358
		//   https://golang.org/issue/5522
		//
		// We don't request gzip if the request is for a range, since
		// auto-decoding a portion of a gzipped document will just fail
		// anyway. See https://golang.org/issue/8923
		requestedGzip = true
		args = append(args, "-H", "Accept-Encoding: gzip")
	}

	// For client requests, certain headers such as Content-Length
	// and Connection are automatically written when needed and
	// values in Header may be ignored. See the documentation
	// for the Request.Write method.
	//
	// Cookies will appear in the header dict if set:
	for h, vs := range rq.Header {
		for _, v := range vs {
			if hv, ok := curlHeader(h, v); ok {
				args = append(args, "-H", hv)
			}
		}
	}

	args = append(args, rq.URL.String())

	if rq.ContentLength > 0 && rdr != nil {
		rdr = io.LimitReader(rdr, rq.ContentLength)
	}

	ctx := rq.Context()
	if dead, ok := ctx.Deadline(); ok {
		maxTime := dead.Sub(time.Now())
		args = append(args, "--max-time", fmt.Sprintf("%f", maxTime.Seconds()))
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = rdr

	// FIXME: capture somehow
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	br := bufio.NewReader(stdout)
	body := &responseBody{rdr: br, cls: stdout, cmd: cmd}

	rs.Request = rq
	if err := beginResponse(&rs, body, br, requestedGzip); err != nil {
		body.Close()
		return nil, err
	}

	return &rs, nil
}

func beginResponse(rs *http.Response, body *responseBody, br *bufio.Reader, requestedGzip bool) error {
	ln, isPrefix, err := br.ReadLine()
	if isPrefix {
		return fmt.Errorf("status line too long")
	}

	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	if err := parseFirstLine(string(ln), rs); err != nil {
		return err
	}

	tp := textproto.NewReader(br)
	hdr, err := tp.ReadMIMEHeader()
	if err != nil {
		return err
	}

	rs.Header = http.Header(hdr)
	rs.Body = body

	if requestedGzip {
		gz := &gzipReadCloser{body: body.rdr, cls: body.cls}
		body.rdr, body.cls = gz, gz
	}

	return nil
}

func parseFirstLine(line string, resp *http.Response) (err error) {
	if i := strings.IndexByte(line, ' '); i == -1 {
		return badStringError("malformed HTTP response", line)
	} else {
		resp.Proto = line[:i]
		resp.Status = strings.TrimLeft(line[i+1:], " ")
	}
	statusCode := resp.Status
	if i := strings.IndexByte(resp.Status, ' '); i != -1 {
		statusCode = resp.Status[:i]
	}
	if len(statusCode) != 3 {
		return badStringError("malformed HTTP status code", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return badStringError("malformed HTTP status code", statusCode)
	}
	var ok bool
	if resp.ProtoMajor, resp.ProtoMinor, ok = parseHTTPVersion(resp.Proto); !ok {
		return badStringError("malformed HTTP version", resp.Proto)
	}
	return nil
}

func parseHTTPVersion(ver string) (major, minor int, ok bool) {
	if ver == "HTTP/2" {
		return 2, 0, true
	}
	return http.ParseHTTPVersion(ver)
}

func badStringError(what, val string) error { return fmt.Errorf("%s %q", what, val) }

type responseBody struct {
	rdr io.Reader
	cls io.Closer
	cmd *exec.Cmd
}

func (bd *responseBody) Read(b []byte) (n int, err error) {
	n, err = bd.rdr.Read(b)
	if err == io.EOF {
		bd.Close()
	}
	return n, err
}

func (bd *responseBody) Close() (err error) {
	if bd.cls != nil {
		err = bd.cls.Close()
		bd.cmd.Process.Kill() // FIXME: is this right?
		if werr := bd.cmd.Wait(); err == nil && werr != nil {
			err = werr
		}
		bd.cls = nil
	}
	return err
}

func curlHeader(h string, v string) (hv string, ok bool) {
	trimmed := strings.TrimSpace(h)
	if len(trimmed) == 0 || trimmed[0] == '@' {
		// Curl interprets '@' as "take headers from this file"
		return "", false
	}

	if v == "" {
		// > If you send the custom header with no-value then its header must be
		// > terminated with a semicolon
		hv = fmt.Sprintf("%s;", h)
	} else {
		hv = fmt.Sprintf("%s: %s", h, v)
	}
	return hv, true
}
