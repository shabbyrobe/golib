package bytestream

import (
	"encoding/binary"
	"fmt"
)

type Binary interface {
	Uint8Within(min, max uint8) (uint8, error)
	Uint16Within(min, max uint16) (uint16, error)
	Uint32Within(min, max uint32) (uint32, error)
	Uint64Within(min, max uint64) (uint64, error)
	Int8Within(min, max int8) (int8, error)
	Int16Within(min, max int16) (int16, error)
	Int32Within(min, max int32) (int32, error)
	Int64Within(min, max int64) (int64, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)
}

var leOrder = binary.LittleEndian

type LittleEndian struct {
	ByteStream
}

var _ Binary = &LittleEndian{}

func NewLittleEndian(stream ByteStream) *LittleEndian {
	return &LittleEndian{stream}
}

func (le *LittleEndian) Uint8Within(min, max uint8) (uint8, error) {
	v, err := le.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Uint16Within(min, max uint16) (uint16, error) {
	b, err := le.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint16(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Uint32Within(min, max uint32) (uint32, error) {
	b, err := le.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint32(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Uint64Within(min, max uint64) (uint64, error) {
	b, err := le.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint64(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Int8Within(min, max int8) (int8, error) {
	uv, err := le.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	v := int8(uv)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Int16Within(min, max int16) (int16, error) {
	b, err := le.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := int16(leOrder.Uint16(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Int32Within(min, max int32) (int32, error) {
	b, err := le.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := int32(leOrder.Uint32(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Int64Within(min, max int64) (int64, error) {
	b, err := le.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := int64(leOrder.Uint64(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (le *LittleEndian) Uint8() (uint8, error) {
	return le.ByteStream.ReadByte()
}

func (le *LittleEndian) Uint16() (uint16, error) {
	b, err := le.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint16(b), nil
}

func (le *LittleEndian) Uint32() (uint32, error) {
	b, err := le.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint32(b), nil
}

func (le *LittleEndian) Uint64() (uint64, error) {
	b, err := le.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint64(b), nil
}

func (le *LittleEndian) Int8() (int8, error) {
	uv, err := le.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(uv), nil
}

func (le *LittleEndian) Int16() (int16, error) {
	b, err := le.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return int16(leOrder.Uint16(b)), nil
}

func (le *LittleEndian) Int32() (int32, error) {
	b, err := le.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return int32(leOrder.Uint32(b)), nil
}

func (le *LittleEndian) Int64() (int64, error) {
	b, err := le.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return int64(leOrder.Uint64(b)), nil
}

var beOrder = binary.BigEndian

type BigEndian struct {
	ByteStream
}

var _ Binary = &BigEndian{}

func NewBigEndian(stream ByteStream) *BigEndian {
	return &BigEndian{stream}
}

func (be *BigEndian) Uint8Within(min, max uint8) (uint8, error) {
	v, err := be.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Uint16Within(min, max uint16) (uint16, error) {
	b, err := be.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint16(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Uint32Within(min, max uint32) (uint32, error) {
	b, err := be.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint32(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Uint64Within(min, max uint64) (uint64, error) {
	b, err := be.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint64(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Int8Within(min, max int8) (int8, error) {
	uv, err := be.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	v := int8(uv)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Int16Within(min, max int16) (int16, error) {
	b, err := be.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := int16(beOrder.Uint16(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Int32Within(min, max int32) (int32, error) {
	b, err := be.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := int32(beOrder.Uint32(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Int64Within(min, max int64) (int64, error) {
	b, err := be.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := int64(beOrder.Uint64(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (be *BigEndian) Uint8() (uint8, error) {
	return be.ByteStream.ReadByte()
}

func (be *BigEndian) Uint16() (uint16, error) {
	b, err := be.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint16(b), nil
}

func (be *BigEndian) Uint32() (uint32, error) {
	b, err := be.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint32(b), nil
}

func (be *BigEndian) Uint64() (uint64, error) {
	b, err := be.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint64(b), nil
}

func (be *BigEndian) Int8() (int8, error) {
	uv, err := be.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(uv), nil
}

func (be *BigEndian) Int16() (int16, error) {
	b, err := be.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return int16(beOrder.Uint16(b)), nil
}

func (be *BigEndian) Int32() (int32, error) {
	b, err := be.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return int32(beOrder.Uint32(b)), nil
}

func (be *BigEndian) Int64() (int64, error) {
	b, err := be.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return int64(beOrder.Uint64(b)), nil
}
