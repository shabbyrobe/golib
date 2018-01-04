package jsonwriter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

var valueCases = []struct {
	out string
	do  func(w *Writer) error
}{
	{`"bar"`, func(w *Writer) error { return w.WriteString("bar") }},
	{`"bar"`, func(w *Writer) error { return w.WriteStringUnescaped("bar") }},
	{`-1`, func(w *Writer) error { return w.WriteInt(-1) }},
	{`1`, func(w *Writer) error { return w.WriteUint(1) }},
	{`-1.1`, func(w *Writer) error { return w.WriteFloat64(-1.1) }},
	{`true`, func(w *Writer) error { return w.WriteBool(true) }},
	{`null`, func(w *Writer) error { return w.WriteNull() }},
	{`[]`, func(w *Writer) error {
		if err := w.StartList(); err != nil {
			return err
		}
		return w.EndList()
	}},
	{`{}`, func(w *Writer) error {
		if err := w.StartObject(); err != nil {
			return err
		}
		return w.EndObject()
	}},
}

func TestWriterObject(t *testing.T) {
	tt := assert.WrapTB(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)

	tt.MustOK(w.StartObject())
	tt.MustOK(w.EndObject())
	tt.MustOK(w.Close())

	tt.MustEqual("{}", buf.String())
}

func TestWriterObjectFirstKeyValue(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)

			tt.MustOK(w.StartObject())
			tt.MustOK(w.WriteKey("foo"))
			tt.MustOK(c.do(w))
			tt.MustOK(w.EndObject())
			tt.MustOK(w.Close())

			tt.MustEqual(fmt.Sprintf(`{"foo":%s}`, c.out), buf.String())
		})
	}
}

func TestWriterObjectSecondKeyValue(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)

			tt.MustOK(w.StartObject())
			tt.MustOK(w.WriteKey("foo"))
			tt.MustOK(w.WriteString("yep"))

			tt.MustOK(w.WriteKey("bar"))
			tt.MustOK(c.do(w))
			tt.MustOK(w.Close())

			tt.MustEqual(fmt.Sprintf(`{"foo":"yep","bar":%s}`, c.out), buf.String())
		})
	}
}

func TestWriterObjectManyKeyValues(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)

			tt.MustOK(w.StartObject())
			tt.MustOK(w.WriteKey("foo"))
			tt.MustOK(w.WriteString("yep"))

			tt.MustOK(w.WriteKey("bar"))
			tt.MustOK(c.do(w))

			tt.MustOK(w.WriteKey("baz"))
			tt.MustOK(c.do(w))

			tt.MustOK(w.Close())

			tt.MustEqual(fmt.Sprintf(`{"foo":"yep","bar":%[1]s,"baz":%[1]s}`, c.out), buf.String())
		})
	}
}

func TestWriterList(t *testing.T) {
	tt := assert.WrapTB(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)

	tt.MustOK(w.StartList())
	tt.MustOK(w.EndList())
	tt.MustOK(w.Close())

	tt.MustEqual(`[]`, buf.String())
}

func TestWriterListFirstValue(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.MustOK(w.StartList())
			tt.MustOK(c.do(w))
			tt.MustOK(w.Close())
			tt.MustEqual(fmt.Sprintf(`[%s]`, c.out), buf.String())
		})
	}
}

func TestWriterListValues(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.MustOK(w.StartList())
			tt.MustOK(w.WriteString("foo"))
			tt.MustOK(c.do(w))
			tt.MustOK(w.Close())
			tt.MustEqual(fmt.Sprintf(`["foo",%s]`, c.out), buf.String())
		})
	}
}

func TestWriterListMultipleValues(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.MustOK(w.StartList())
			tt.MustOK(w.WriteString("foo"))
			tt.MustOK(c.do(w))
			tt.MustOK(c.do(w))
			tt.MustOK(w.Close())
			tt.MustEqual(fmt.Sprintf(`["foo",%[1]s,%[1]s]`, c.out), buf.String())
		})
	}
}

func TestWriteStringUnescaped(t *testing.T) {
	tt := assert.WrapTB(t)

	{ // happy path
		var buf bytes.Buffer
		w := NewWriter(&buf)
		tt.MustOK(w.WriteStringUnescaped("test"))
		tt.MustOK(w.Close())
		tt.MustEqual(`"test"`, buf.String())
	}

	{ // dodgy strings can be written no problem
		var buf bytes.Buffer
		w := NewWriter(&buf)
		tt.MustOK(w.WriteStringUnescaped(`"oops"`))
		tt.MustOK(w.Close())
		tt.MustEqual(`""oops""`, buf.String())
	}
}

func TestWriteKeyUnescaped(t *testing.T) {
	tt := assert.WrapTB(t)

	{ // happy path
		var buf bytes.Buffer
		w := NewWriter(&buf)
		tt.MustOK(w.StartObject())
		tt.MustOK(w.WriteKeyUnescaped("test"))
		tt.MustOK(w.WriteString("test"))
		tt.MustOK(w.Close())
		tt.MustEqual(`{"test":"test"}`, buf.String())
	}

	{ // dodgy strings can be written no problem
		var buf bytes.Buffer
		w := NewWriter(&buf)
		tt.MustOK(w.StartObject())
		tt.MustOK(w.WriteKeyUnescaped(`"oops"`))
		tt.MustOK(w.WriteString("test"))
		tt.MustOK(w.Close())
		tt.MustEqual(`{""oops"":"test"}`, buf.String())
	}
}

func TestWriteOptionalString(t *testing.T) {
	tt := assert.WrapTB(t)

	{ // null
		var buf bytes.Buffer
		w := NewWriter(&buf)
		tt.MustOK(w.WriteOptionalString(nil))
		tt.MustOK(w.Close())
		tt.MustEqual(`null`, buf.String())
	}

	{ // value
		var buf bytes.Buffer
		w := NewWriter(&buf)
		str := "yep"
		tt.MustOK(w.WriteOptionalString(&str))
		tt.MustOK(w.Close())
		tt.MustEqual(`"yep"`, buf.String())
	}
}

func TestWriteValueInvalid(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.MustOK(w.WriteString("yep"))
			err := c.do(w)
			tt.MustAssert(err != nil)
			tt.MustAssert(strings.Contains(err.Error(), "can only write one bare value"), err.Error())
		})
	}
}

func TestWriteValueInsteadOfKey(t *testing.T) {
	for _, c := range valueCases {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.MustOK(w.StartObject())
			err := c.do(w)
			tt.MustAssert(err != nil)
			tt.MustAssert(strings.Contains(err.Error(), `unexpected node state "key", expected "value"`), err.Error())
		})
	}
}

func TestWriterPopTooManyNodes(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	tt := assert.WrapTB(t)
	tt.MustOK(w.StartList())
	tt.MustOK(w.StartList())
	tt.MustOK(w.EndList())
	tt.MustOK(w.EndList())

	err := w.EndList()
	tt.MustAssert(err != nil)
	tt.MustAssert(strings.Contains(err.Error(), `unexpected node kind "root", expected "list"`), err.Error())
}

func TestWriterReusableBuffer(t *testing.T) {
	tt := assert.WrapTB(t)
	var buf bytes.Buffer
	wbuf := bufio.NewWriter(&buf)

	w := NewWriterBuffer(&buf, wbuf)
	tt.MustOK(w.StartObject())
	tt.MustOK(w.EndObject())
	tt.MustOK(w.Flush())
	tt.MustEqual(`{}`, buf.String())
	buf.Reset()

	w = NewWriterBuffer(&buf, wbuf)
	tt.MustOK(w.StartObject())
	tt.MustOK(w.EndObject())
	tt.MustOK(w.Flush())
	tt.MustEqual(`{}`, buf.String())
	buf.Reset()
}

var (
	benchStr   string
	benchBytes []byte
)

func BenchmarkPrinterString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		p := &printer{Writer: bufio.NewWriterSize(&buf, 20)}
		p.string("1234578")
		p.Flush()
		benchStr = buf.String()
	}
}

func BenchmarkPrinterStringDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		p := &printer{Writer: bufio.NewWriterSize(&buf, 20)}
		p.WriteString("1234567890")
		p.Flush()
		benchStr = buf.String()
	}
}

type TestStruct struct {
	Str   string
	Int   int
	Float float64
	List  []string
	Child struct {
		Str   string
		Int   int
		Float float64
	}
	Dict map[string]string
}

func BenchmarkWriterComplex(b *testing.B) {
	var buf bytes.Buffer
	wbuf := bufio.NewWriter(&buf)

	for i := 0; i < b.N; i++ {
		w := NewWriterBuffer(&buf, wbuf)
		w.StartObject()
		w.WriteKeyValueString("Str", "yep")
		w.WriteKeyValueInt("Int", 1)
		w.WriteKeyValueFloat64("Float", 1.1)

		w.WriteKey("List")
		w.StartList()
		for _, v := range []string{"a", "b", "c"} {
			w.WriteString(v)
		}
		w.EndList()

		w.WriteKey("Child")
		w.StartObject()
		w.WriteKeyValueString("Str", "yep")
		w.WriteKeyValueInt("Int", 1)
		w.WriteKeyValueFloat64("Float", 1.1)
		w.EndObject()

		w.WriteKey("Dict")
		w.StartObject()
		for _, v := range []string{"a", "b", "c"} {
			w.WriteKeyValueString(v, v)
		}
		w.EndObject()
		w.EndObject()

		w.Flush()
		benchBytes = buf.Bytes()
		buf.Reset()
	}
}

func BenchmarkWriterStdlibComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := &TestStruct{
			Str:   "yep",
			Int:   1,
			Float: 1.1,
			List:  []string{"a", "b", "c"},
			Dict: map[string]string{
				"a": "a", "b": "b", "c": "c",
			},
		}
		c.Child.Str = "yep"
		c.Child.Int = 1
		c.Child.Float = 1.1

		benchBytes, _ = json.Marshal(c)
	}
}
