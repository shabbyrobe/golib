package jsonwriter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	initialNodeDepth = 8
	defaultBufsize   = 2048
)

type Writer struct {
	printer printer

	indent  string
	nodes   []node
	current int

	EscapeHTML bool
	Enforce    bool

	// Defaults to \n.
	NewlineString string

	// Determines how much memory the internal buffer will use. Set to 0 to use
	// the default.
	InitialBufSize int
}

type Option func(w *Writer)

func EscapeHTML(on bool) Option    { return func(w *Writer) { w.EscapeHTML = on } }
func WithInitialSize(i int) Option { return func(w *Writer) { w.InitialBufSize = i } }

func NewWriter(w io.Writer, options ...Option) *Writer {
	return newWriter(w, nil, options...)
}

func NewWriterBuffer(w io.Writer, buf *bufio.Writer, options ...Option) *Writer {
	buf.Reset(w)
	return newWriter(w, buf, options...)
}

func newWriter(w io.Writer, buf *bufio.Writer, options ...Option) *Writer {
	jw := &Writer{}
	jw.Enforce = true
	jw.EscapeHTML = false
	jw.NewlineString = "\n"
	jw.nodes = make([]node, initialNodeDepth)
	jw.nodes[0].kind = rootNode | valueState
	for _, o := range options {
		o(jw)
	}
	if jw.InitialBufSize <= 0 {
		jw.InitialBufSize = defaultBufsize
	}
	if buf == nil {
		buf = bufio.NewWriterSize(w, jw.InitialBufSize)
	}
	jw.printer = printer{
		Writer:     buf,
		escapeHTML: jw.EscapeHTML,
	}
	return jw
}

// Close ends every node on the stack and calls Flush()
func (w *Writer) Close() error {
	if err := w.EndAll(); err != nil {
		return err
	}
	return w.Flush()
}

// Flush ensures the output buffer accumuated inside the Writer
// is fully written to the underlying io.Writer.
func (w *Writer) Flush() error {
	return w.printer.Flush()
}

func (w *Writer) StartObject() error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.push(objectNode | keyState); err != nil {
		return err
	}
	return w.printer.WriteByte('{')
}

func (w *Writer) EndObject() error {
	if w.Enforce {
		if err := w.checkParent(objectNode, keyState); err != nil {
			return err
		}
	}
	if err := w.pop(); err != nil {
		return err
	}
	return w.printer.WriteByte('}')
}

func (w *Writer) StartList() error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.push(listNode | valueState); err != nil {
		return err
	}
	return w.printer.WriteByte('[')
}

func (w *Writer) EndList() error {
	if w.Enforce {
		if err := w.checkParent(listNode, valueState); err != nil {
			return err
		}
	}
	if err := w.pop(); err != nil {
		return err
	}
	return w.printer.WriteByte(']')
}

// EndAll ends every node on the stack
func (w *Writer) EndAll() error {
	for w.current > 0 {
		switch w.nodes[w.current].kind {
		case listNode | valueState:
			if err := w.EndList(); err != nil {
				return err
			}
		case objectNode | keyState:
			if err := w.EndObject(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("uncloseable node %q", w.nodes[w.current].kind)
		}
	}
	return nil
}

// WriteRaw lets you write any old string you like into the output regardless of
// whether or not it is valid.
func (w *Writer) WriteRaw(raw string) error {
	_, err := w.printer.WriteString(raw)
	return err
}

func (w *Writer) WriteNull() error {
	if w.Enforce {
		if err := w.checkParent(rootNode|objectNode|listNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	_, err := w.printer.WriteString("null")
	return err
}

func (w *Writer) WriteKey(key string) error {
	if w.Enforce {
		if err := w.checkParent(objectNode, keyState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	w.printer.string(key)
	return w.printer.cachedWriteError()
}

// WriteKeyUnescaped lets you write any old string you like into the output
// regardless of whether or not it is valid. This can be faster than WriteKey,
// but at the expense of a lot of safety.
//
// It will be quoted but not escaped in any way other than that.
// Writer.WriteKeyUnescaped(`"`) == `"""`, which is not valid JSON. Exercise
// caution.
func (w *Writer) WriteKeyUnescaped(raw string) error {
	if w.Enforce {
		if err := w.checkParent(objectNode, keyState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	w.printer.WriteByte('"')
	w.printer.WriteString(raw)
	return w.printer.WriteByte('"')
}

func (w *Writer) WriteString(str string) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	w.printer.string(str)
	return w.printer.cachedWriteError()
}

// WriteStringUnescaped lets you write any old string you like into the output
// regardless of whether or not it is valid. This can be faster than WriteString,
// but at the expense of a lot of safety.
//
// It will be quoted but not escaped in any way other than that.
// Writer.WriteStringKeyUnescaped(`"`) == `"""`, which is not valid JSON. Exercise
// caution.
func (w *Writer) WriteStringUnescaped(str string) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	w.printer.WriteByte('"')
	w.printer.WriteString(str)
	return w.printer.WriteByte('"')
}

func (w *Writer) WriteBool(v bool) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	if v {
		w.printer.WriteString("true")
	} else {
		w.printer.WriteString("false")
	}
	return w.printer.cachedWriteError()
}

func (w *Writer) WriteInt64(v int64) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	_, err := w.printer.WriteString(strconv.FormatInt(v, 10))
	return err
}

func (w *Writer) WriteUint64(v uint64) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	_, err := w.printer.WriteString(strconv.FormatUint(v, 10))
	return err
}

func (w *Writer) WriteFloat64(f float64) error {
	if w.Enforce {
		if err := w.checkParent(rootNode|listNode|objectNode, valueState); err != nil {
			return err
		}
	}
	if err := w.next(); err != nil {
		return err
	}
	return float64Encoder(&w.printer, f, w.printer.escapeHTML)
}

// {{{ Type madness:

func (w *Writer) WriteInt32(v int32) error { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteInt16(v int16) error { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteInt8(v int8) error   { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteInt(v int) error     { return w.WriteInt64(int64(v)) }

func (w *Writer) WriteUint32(v uint32) error { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteUint16(v uint16) error { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteUint8(v uint8) error   { return w.WriteInt64(int64(v)) }
func (w *Writer) WriteUint(v uint) error     { return w.WriteInt64(int64(v)) }

func (w *Writer) WriteFloat32(v float32) error { return w.WriteFloat64(float64(v)) }

// }}}

// {{{ Optional type madness:

func (w *Writer) WriteOptionalString(str *string) error {
	if str == nil {
		return w.WriteNull()
	}
	return w.WriteString(*str)
}

func (w *Writer) WriteOptionalBool(v *bool) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteBool(*v)
}

func (w *Writer) WriteOptionalInt64(v *int64) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteInt64(*v)
}

func (w *Writer) WriteOptionalInt32(v *int32) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteInt32(*v)
}

func (w *Writer) WriteOptionalInt16(v *int16) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteInt16(*v)
}

func (w *Writer) WriteOptionalInt8(v *int8) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteInt8(*v)
}

func (w *Writer) WriteOptionalInt(v *int) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteInt(*v)
}

func (w *Writer) WriteOptionalUint64(v *uint64) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteUint64(*v)
}

func (w *Writer) WriteOptionalUint32(v *uint32) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteUint32(*v)
}

func (w *Writer) WriteOptionalUint16(v *uint16) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteUint16(*v)
}

func (w *Writer) WriteOptionalUint8(v *uint8) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteUint8(*v)
}

func (w *Writer) WriteOptionalUint(v *uint) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteUint(*v)
}

func (w *Writer) WriteOptionalFloat64(v *float64) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteFloat64(*v)
}

func (w *Writer) WriteOptionalFloat32(v *float32) error {
	if v == nil {
		return w.WriteNull()
	}
	return w.WriteFloat32(*v)
}

// }}}

// {{{ Key/value helpers:

func (w *Writer) WriteKeyValueString(k string, v string) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteString(v)
}

func (w *Writer) WriteKeyValueBool(k string, v bool) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteBool(v)
}

func (w *Writer) WriteKeyValueNull(k string) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteNull()
}

func (w *Writer) WriteKeyValueFloat64(k string, v float64) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteFloat64(v)
}

func (w *Writer) WriteKeyValueFloat32(k string, v float32) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteFloat32(v)
}

func (w *Writer) WriteKeyValueInt64(k string, v int64) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteInt64(v)
}

func (w *Writer) WriteKeyValueInt32(k string, v int32) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteInt32(v)
}

func (w *Writer) WriteKeyValueInt16(k string, v int16) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteInt16(v)
}

func (w *Writer) WriteKeyValueInt8(k string, v int8) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteInt8(v)
}

func (w *Writer) WriteKeyValueInt(k string, v int) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteInt(v)
}

func (w *Writer) WriteKeyValueUint64(k string, v uint64) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteUint64(v)
}

func (w *Writer) WriteKeyValueUint32(k string, v uint32) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteUint32(v)
}

func (w *Writer) WriteKeyValueUint16(k string, v uint16) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteUint16(v)
}

func (w *Writer) WriteKeyValueUint8(k string, v uint8) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteUint8(v)
}

func (w *Writer) WriteKeyValueUint(k string, v uint) error {
	if err := w.WriteKey(k); err != nil {
		return err
	}
	return w.WriteUint(v)
}

// }}}

func (w *Writer) checkParent(node nodeKind, state nodeKind) error {
	k := w.nodes[w.current].kind
	if k&node == 0 {
		return fmt.Errorf("jsonwriter: unexpected node kind %q, expected %q", k.String(), node.String())
	}
	if k&state == 0 {
		return fmt.Errorf("jsonwriter: unexpected node state %q, expected %q", k.String(), state.StateString())
	}
	return nil
}

func (w *Writer) next() error {
	n := &w.nodes[w.current]

	switch n.kind {
	case objectNode | valueState:
		n.kind &= ^valueState
		n.kind |= keyState
		n.children++
		w.printer.WriteByte(':')

	case objectNode | keyState:
		n.kind &= ^keyState
		n.kind |= valueState
		if n.children > 0 {
			w.printer.WriteByte(',')
		}

	case listNode | valueState:
		if n.children > 0 {
			w.printer.WriteByte(',')
		}
		n.children++

	case rootNode | valueState:
		if n.children != 0 {
			return fmt.Errorf("can only write one bare value outside of list or object")
		}
		n.children++
	default:
		return fmt.Errorf("unexpected state")
	}

	return w.printer.cachedWriteError()
}

func (w *Writer) push(kind nodeKind) error {
	if err := w.next(); err != nil {
		return err
	}
	if len(w.nodes) <= w.current+1 {
		w.nodes = append(w.nodes, node{})
	}
	w.current++
	w.nodes[w.current].kind = kind
	w.nodes[w.current].children = 0
	return nil
}

func (w *Writer) pop() error {
	w.current--
	return nil
}
