package iotools

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
)

func TestLineChunker(t *testing.T) {
	for idx, tc := range []struct {
		sz     int
		input  string
		chunks []string
	}{
		{},
		{5, "foo", []string{"foo"}},
		{4, "foo", []string{"foo"}},
		{3, "foo", []string{"foo"}},
		{4, "foo\nbar\n", []string{"foo\n", "bar\n"}},
		{5, "foo\nbar\n", []string{"foo\n", "bar\n"}},
		{5, "foo\r\nbar\r\n", []string{"foo\r\n", "bar\r\n"}},
		{5, "foo\nbar", []string{"foo\n", "bar"}},

		{10, "foo\nbar\nbaz\nqux", []string{"foo\nbar\n", "baz\nqux"}},
	} {
		for _, makeReader := range []func(s string) io.Reader{
			func(s string) io.Reader { return strings.NewReader(s) },
			func(s string) io.Reader { return iotest.OneByteReader(strings.NewReader(s)) },
			func(s string) io.Reader { return iotest.DataErrReader(strings.NewReader(s)) },
		} {
			t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
				lc := NewLineChunker(makeReader(tc.input), 0)
				into := make([]byte, tc.sz)
				var result []string
				for {
					n, err := lc.NextChunk(into)
					if err == io.EOF {
						break
					} else if err != nil {
						t.Fatal(err)
					}
					result = append(result, string(into[:n]))
				}

				if !reflect.DeepEqual(tc.chunks, result) {
					t.Fatal("expected", tc.chunks, "got", result)
				}
			})
		}
	}
}

func TestLineChunkerFailures(t *testing.T) {
	for idx, tc := range []struct {
		sz     int
		input  string
		chunks []string
	}{
		{5, "foo bar\n", nil},
		{5, "foo\nbar baz\n", []string{"foo\n"}},
		{5, "food\nbar baz\n", []string{"food\n"}},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			lc := NewLineChunker(strings.NewReader(tc.input), 0)
			into := make([]byte, tc.sz)
			failed := false
			var result []string
			for {
				n, err := lc.NextChunk(into)
				if err == io.EOF {
					break
				} else if err != nil {
					failed = true
					break
				}
				result = append(result, string(into[:n]))
			}

			if !reflect.DeepEqual(tc.chunks, result) {
				t.Fatal("expected", tc.chunks, "got", result)
			}
			if !failed {
				t.Fatal()
			}
		})
	}
}
