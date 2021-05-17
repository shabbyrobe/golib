package bytestream

import (
	"io"
	"testing"
)

func TestLittleEndian(t *testing.T) {
	t.Run("uint8", func(t *testing.T) {
		bin := NewLittleEndian(NewBuffer([]byte{1, 2, 3}))
		assertUint8(t, bin, 1)
		assertUint8(t, bin, 2)
		assertUint8(t, bin, 3)
		assertUint8NotFound(t, bin)
	})

	t.Run("uint16", func(t *testing.T) {
		bin := NewLittleEndian(NewBuffer([]byte{1, 2, 3, 4}))
		assertUint16(t, bin, 1|2<<8)
		assertUint16(t, bin, 3|4<<8)
		assertUint16NotFound(t, bin)

		bin = NewLittleEndian(NewBuffer([]byte{1}))
		assertUint16NotFound(t, bin)
		assertUint8(t, bin, 1)

		bin = NewLittleEndian(NewBuffer([]byte{}))
		assertUint16NotFound(t, bin)
	})

	t.Run("uint32", func(t *testing.T) {
		bin := NewLittleEndian(NewBuffer([]byte{1, 2, 3, 4, 1, 2, 3, 5}))
		assertUint32(t, bin, 1|2<<8|3<<16|4<<24)
		assertUint32(t, bin, 1|2<<8|3<<16|5<<24)
		assertUint32NotFound(t, bin)
	})

	t.Run("uint64", func(t *testing.T) {
		bin := NewLittleEndian(NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}))
		assertUint64(t, bin, 1|2<<8|3<<16|4<<24|5<<32|6<<40|7<<48|8<<56)
		assertUint64NotFound(t, bin)
	})

	t.Run("int8", func(t *testing.T) {
		bin := NewLittleEndian(NewBuffer([]byte{0, 1, 127, 128, 255}))
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
		bin := NewBigEndian(NewBuffer([]byte{1, 2, 3}))
		assertUint8(t, bin, 1)
		assertUint8(t, bin, 2)
		assertUint8(t, bin, 3)
		assertUint8NotFound(t, bin)
	})

	t.Run("uint16", func(t *testing.T) {
		bin := NewBigEndian(NewBuffer([]byte{1, 2, 3, 4}))
		assertUint16(t, bin, 1<<8|2)
		assertUint16(t, bin, 3<<8|4)
		assertUint16NotFound(t, bin)

		bin = NewBigEndian(NewBuffer([]byte{1}))
		assertUint16NotFound(t, bin)
		assertUint8(t, bin, 1)

		bin = NewBigEndian(NewBuffer([]byte{}))
		assertUint16NotFound(t, bin)
	})

	t.Run("uint32", func(t *testing.T) {
		bin := NewBigEndian(NewBuffer([]byte{1, 2, 3, 4, 1, 2, 3, 5}))
		assertUint32(t, bin, 1<<24|2<<16|3<<8|4)
		assertUint32(t, bin, 1<<24|2<<16|3<<8|5)
		assertUint32NotFound(t, bin)
	})

	t.Run("uint64", func(t *testing.T) {
		bin := NewBigEndian(NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}))
		assertUint64(t, bin, 1<<56|2<<48|3<<40|4<<32|5<<24|6<<16|7<<8|8)
		assertUint64NotFound(t, bin)
	})

	t.Run("int8", func(t *testing.T) {
		bin := NewBigEndian(NewBuffer([]byte{0, 1, 127, 128, 255}))
		assertInt8(t, bin, 0)
		assertInt8(t, bin, 1)
		assertInt8(t, bin, 127)
		assertInt8(t, bin, -128)
		assertInt8(t, bin, -1)
		assertInt8NotFound(t, bin)
	})
}

func assertUint8NotFound(t *testing.T, bin Binary) {
	t.Helper()
	if _, err := bin.Uint8(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint8(t *testing.T, bin Binary, v uint8) {
	t.Helper()
	r, err := bin.Uint8()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint8", v, "!=", r)
	}
}

func assertUint16NotFound(t *testing.T, bin Binary) {
	t.Helper()
	if _, err := bin.Uint16(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint16(t *testing.T, bin Binary, v uint16) {
	t.Helper()
	r, err := bin.Uint16()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint16", v, "!=", r)
	}
}

func assertUint32NotFound(t *testing.T, bin Binary) {
	t.Helper()
	if _, err := bin.Uint32(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint32(t *testing.T, bin Binary, v uint32) {
	t.Helper()
	r, err := bin.Uint32()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint32", v, "!=", r)
	}
}

func assertUint64NotFound(t *testing.T, bin Binary) {
	t.Helper()
	if _, err := bin.Uint64(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertUint64(t *testing.T, bin Binary, v uint64) {
	t.Helper()
	r, err := bin.Uint64()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("uint64", v, "!=", r)
	}
}

func assertInt8NotFound(t *testing.T, bin Binary) {
	t.Helper()
	if _, err := bin.Int8(); err != io.ErrUnexpectedEOF {
		t.Fatal()
	}
}

func assertInt8(t *testing.T, bin Binary, v int8) {
	t.Helper()
	r, err := bin.Int8()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("int8", v, "!=", r)
	}
}

func assertInt16(t *testing.T, bin Binary, v int16) {
	t.Helper()
	r, err := bin.Int16()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("int16", v, "!=", r)
	}
}

func assertInt32(t *testing.T, bin Binary, v int32) {
	t.Helper()
	r, err := bin.Int32()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("int32", v, "!=", r)
	}
}

func assertInt64(t *testing.T, bin Binary, v int64) {
	t.Helper()
	r, err := bin.Int64()
	if err != nil {
		t.Fatal(err)
	}
	if v != r {
		t.Fatal("int64", v, "!=", r)
	}
}
