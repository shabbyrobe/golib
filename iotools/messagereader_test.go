package iotools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestMessageReaderBytePrefix(t *testing.T) {
	buf := []byte{}
	expected := [][]byte{}

	for i := 0; i < 256; i++ {
		buf = append(buf, byte(i))
		var cur []byte
		for j := 0; j < i; j++ {
			buf = append(buf, byte(j))
			cur = append(cur, byte(j))
		}
		if i > 0 {
			expected = append(expected, cur)
		}
	}

	// This attempts to chop the last message up. Probably worth making a separate test.
	maxScratch := len(buf) + 2
	scratch := make([]byte, maxScratch)
	for scratchSize := maxScratch - 256 - 10; scratchSize <= maxScratch; scratchSize++ {
		pr := NewMessageReaderBytePrefix(bytes.NewReader(buf), scratch[:scratchSize])
		i := 0
		for {
			out, n, err := pr.ReadNext()
			if err == io.EOF {
				break
			}
			if i < len(expected) {
				if err != nil {
					t.Fatal(err)
				}
				if n != len(out) {
					t.Fatalf("failed at index %d", i)
				}
				if !reflect.DeepEqual(expected[i], out) {
					t.Fatalf("failed at index %d", i)
				}
			}
			i++
		}
		if i != 255 { // 255, not 256: the '0' case doesn't yield a message.
			t.Fatal()
		}
	}
}

func TestMessageReaderBytePrefixReadEmpty(t *testing.T) {
	pr := NewMessageReaderBytePrefix(bytes.NewReader([]byte{}), nil)
	i := 0
	for {
		out, n, err := pr.ReadNext()
		if err == io.EOF {
			break
		}
		if len(out) != 0 {
			t.Fatal()
		}
		if n != 0 {
			t.Fatal()
		}
		i++
	}
	if i != 0 {
		t.Fatal()
	}
}

func TestMessageReaderBytePrefixSplitRead(t *testing.T) {
	// Creates a buffer of messages, all 255 bytes long, except
	// for the first one. We test all possible combinations of the
	// first message's length. This ensures that the split between
	// the two ReadFull calls will land on every possible byte index
	// in the middle of a single message.
	for i := 1; i < 256; i++ {
		t.Run(fmt.Sprintf("sz=%db", i), func(t *testing.T) {
			msgs := make([]byte, 1024)
			msgs[0] = byte(i)
			lens := []int{i}

			for j := i + 1; j < 1024; j += 256 {
				cur := 255
				if 1024-j < 255 {
					cur = 1024 - j - 1
				}
				msgs[j] = byte(cur)
				if cur > 0 {
					lens = append(lens, cur)
				}
			}

			// 512 is the shortest allowable scratch:
			pr := NewMessageReaderBytePrefix(bytes.NewReader(msgs), make([]byte, 512))
			var result []int
			for {
				_, n, err := pr.ReadNext()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				result = append(result, n)
			}

			if !reflect.DeepEqual(lens, result) {
				t.Fatal()
			}
		})
	}
}

var BenchMessageReaderBytePrefixResult int

func BenchmarkMessageReaderBytePrefix(b *testing.B) {
	buf := []byte{}

	for k := 0; k < 16; k++ {
		for i := 0; i < 256; i++ {
			buf = append(buf, byte(i))
			for j := 0; j < i; j++ {
				buf = append(buf, byte(j))
			}
		}
	}

	b.SetBytes(int64(len(buf)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pr := NewMessageReaderBytePrefix(bytes.NewReader(buf), nil)
		for {
			_, n, err := pr.ReadNext()
			if err == io.EOF {
				break
			}
			BenchMessageReaderBytePrefixResult += n
		}
	}
}

func TestMessageReaderShortPrefix(t *testing.T) {
	scratch := make([]byte, 65536)
	in := make([]byte, 65538)
	for i := 1; i < 65536; i += 128 {
		buf := in[:i+2]

		binary.LittleEndian.PutUint16(buf, uint16(i))
		pr := NewMessageReaderShortPrefix(bytes.NewReader(buf), scratch)
		if i > 2 {
			buf[3] = 1
		}

		out, n, err := pr.ReadNext()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(buf[2:], out) {
			t.Fatal()
		}
		if i != n {
			t.Fatal()
		}

		out, n, err = pr.ReadNext()
		if err != io.EOF {
			t.Fatal()
		}
		if len(out) != 0 {
			t.Fatal()
		}
		if n != 0 {
			t.Fatal()
		}
	}
}

func TestMessageReaderShortPrefixReadEmpty(t *testing.T) {
	pr := NewMessageReaderShortPrefix(bytes.NewReader([]byte{}), nil)
	i := 0
	for {
		out, n, err := pr.ReadNext()
		if len(out) != 0 {
			t.Fatal()
		}
		if n != 0 {
			t.Fatal()
		}
		if err == io.EOF {
			break
		}
		i++
	}
	if i != 0 {
		t.Fatal()
	}
}
