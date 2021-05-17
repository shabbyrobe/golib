package bytestream

import (
	"encoding/binary"
	"io"
	"testing"
)

func TestLittleEndian(t *testing.T) {
	t.Run("uint8", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3}))
		assertUint8(t, bin, 1)
		assertUint8(t, bin, 2)
		assertUint8(t, bin, 3)
		assertUint8NotFound(t, bin)
	})

	t.Run("uint16le", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4}))
		assertUint16LE(t, bin, 1|2<<8)
		assertUint16LE(t, bin, 3|4<<8)
		assertUint16NotFound(t, bin)

		bin = NewBinary(NewBuffer([]byte{1}))
		assertUint16NotFound(t, bin)
		assertUint8(t, bin, 1)

		bin = NewBinary(NewBuffer([]byte{}))
		assertUint16NotFound(t, bin)
	})

	t.Run("uint32le", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4, 1, 2, 3, 5}))
		assertUint32LE(t, bin, 1|2<<8|3<<16|4<<24)
		assertUint32LE(t, bin, 1|2<<8|3<<16|5<<24)
		assertUint32NotFound(t, bin)
	})

	t.Run("uint64", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}))
		assertUint64LE(t, bin, 1|2<<8|3<<16|4<<24|5<<32|6<<40|7<<48|8<<56)
		assertUint64NotFound(t, bin)
	})

	t.Run("int8", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{0, 1, 127, 128, 255}))
		assertInt8(t, bin, 0)
		assertInt8(t, bin, 1)
		assertInt8(t, bin, 127)
		assertInt8(t, bin, -128)
		assertInt8(t, bin, -1)
		assertInt8NotFound(t, bin)
	})
}

func TestBigEndian(t *testing.T) {
	t.Run("uint8", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3}))
		assertUint8(t, bin, 1)
		assertUint8(t, bin, 2)
		assertUint8(t, bin, 3)
		assertUint8NotFound(t, bin)
	})

	t.Run("uint16", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4}))
		assertUint16BE(t, bin, 1<<8|2)
		assertUint16BE(t, bin, 3<<8|4)
		assertUint16NotFound(t, bin)

		bin = NewBinary(NewBuffer([]byte{1}))
		assertUint16NotFound(t, bin)
		assertUint8(t, bin, 1)

		bin = NewBinary(NewBuffer([]byte{}))
		assertUint16NotFound(t, bin)
	})

	t.Run("uint32", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4, 1, 2, 3, 5}))
		assertUint32BE(t, bin, 1<<24|2<<16|3<<8|4)
		assertUint32BE(t, bin, 1<<24|2<<16|3<<8|5)
		assertUint32NotFound(t, bin)
	})

	t.Run("uint64", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}))
		assertUint64BE(t, bin, 1<<56|2<<48|3<<40|4<<32|5<<24|6<<16|7<<8|8)
		assertUint64NotFound(t, bin)
	})

	t.Run("int8", func(t *testing.T) {
		bin := NewBinary(NewBuffer([]byte{0, 1, 127, 128, 255}))
		assertInt8(t, bin, 0)
		assertInt8(t, bin, 1)
		assertInt8(t, bin, 127)
		assertInt8(t, bin, -128)
		assertInt8(t, bin, -1)
		assertInt8NotFound(t, bin)
	})
}

func assertUint8NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Uint8(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint8(t *testing.T, bin *Binary, v uint8) {
	t.Helper()
	r, err := bin.Uint8()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint8", v, "!=", r)
	}
}

func assertInt8(t *testing.T, bin *Binary, v int8) {
	t.Helper()
	r, err := bin.Int8()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("int8", v, "!=", r)
	}
}

func assertInt8NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Int8(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint16NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Uint16(binary.LittleEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint16(binary.BigEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint16LE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint16BE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertInt16NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Int16(binary.LittleEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Int16(binary.BigEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Int16LE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Int16BE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint16LE(t *testing.T, bin *Binary, v uint16) {
	t.Helper()

	r, err := bin.Uint16(binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint16", v, "!=", r)
	}
}

func assertUint16BE(t *testing.T, bin *Binary, v uint16) {
	t.Helper()

	r, err := bin.Uint16(binary.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint16", v, "!=", r)
	}
}

func assertUint32NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Uint32(binary.LittleEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint32(binary.BigEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint32LE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint32BE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint32LE(t *testing.T, bin *Binary, v uint32) {
	t.Helper()

	r, err := bin.Uint32(binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint32", v, "!=", r)
	}
}

func assertUint32BE(t *testing.T, bin *Binary, v uint32) {
	t.Helper()

	r, err := bin.Uint32(binary.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint32", v, "!=", r)
	}
}

func assertUint64NotFound(t *testing.T, bin *Binary) {
	t.Helper()
	if _, err := bin.Uint64(binary.LittleEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint64(binary.BigEndian); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint64LE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
	if _, err := bin.Uint64BE(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint64LE(t *testing.T, bin *Binary, v uint64) {
	t.Helper()

	r, err := bin.Uint64(binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint64", v, "!=", r)
	}
}

func assertUint64BE(t *testing.T, bin *Binary, v uint64) {
	t.Helper()

	r, err := bin.Uint64(binary.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint64", v, "!=", r)
	}
}
