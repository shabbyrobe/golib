package bytestream

import (
	"encoding/binary"
	"fmt"
)

var (
	leOrder = binary.LittleEndian
	beOrder = binary.BigEndian
)

type Binary struct {
	ByteStream
}

func NewBinary(stream ByteStream) *Binary {
	return &Binary{stream}
}

func (bin *Binary) Uint8Within(min, max uint8) (uint8, error) {
	v, err := bin.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint16Within(order binary.ByteOrder, min, max uint16) (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := order.Uint16(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint16LEWithin(min, max uint16) (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint16(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint16BEWithin(min, max uint16) (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint16(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint32Within(order binary.ByteOrder, min, max uint32) (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := order.Uint32(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint32LEWithin(min, max uint32) (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint32(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint32BEWithin(min, max uint32) (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint32(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint64Within(order binary.ByteOrder, min, max uint64) (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := order.Uint64(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint64LEWithin(min, max uint64) (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := leOrder.Uint64(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint64BEWithin(min, max uint64) (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := beOrder.Uint64(b)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: uint64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int8Within(min, max int8) (int8, error) {
	uv, err := bin.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	v := int8(uv)
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int8 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int16Within(order binary.ByteOrder, min, max int16) (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := int16(order.Uint16(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int16LEWithin(min, max int16) (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := int16(leOrder.Uint16(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int16BEWithin(min, max int16) (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	v := int16(beOrder.Uint16(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int16 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int32Within(order binary.ByteOrder, min, max int32) (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := int32(order.Uint32(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int32LEWithin(min, max int32) (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := int32(leOrder.Uint32(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int32BEWithin(min, max int32) (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	v := int32(beOrder.Uint32(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int32 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int64Within(order binary.ByteOrder, min, max int64) (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := int64(order.Uint64(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int64LEWithin(min, max int64) (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := int64(leOrder.Uint64(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Int64BEWithin(min, max int64) (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	v := int64(beOrder.Uint64(b))
	if v < min || v > max {
		return v, fmt.Errorf("bytestream: int64 not in range %d <= %d <= %d", min, v, max)
	}
	return v, nil
}

func (bin *Binary) Uint8() (uint8, error) {
	return bin.ByteStream.ReadByte()
}

func (bin *Binary) Uint16(order binary.ByteOrder) (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return order.Uint16(b), nil
}

func (bin *Binary) Uint16LE() (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint16(b), nil
}

func (bin *Binary) Uint16BE() (uint16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint16(b), nil
}

func (bin *Binary) Uint32(order binary.ByteOrder) (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return order.Uint32(b), nil
}

func (bin *Binary) Uint32LE() (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint32(b), nil
}

func (bin *Binary) Uint32BE() (uint32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint32(b), nil
}

func (bin *Binary) Uint64(order binary.ByteOrder) (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return order.Uint64(b), nil
}

func (bin *Binary) Uint64LE() (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return leOrder.Uint64(b), nil
}

func (bin *Binary) Uint64BE() (uint64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return beOrder.Uint64(b), nil
}

func (bin *Binary) Int8() (int8, error) {
	uv, err := bin.ByteStream.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(uv), nil
}

func (bin *Binary) Int16(order binary.ByteOrder) (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return int16(order.Uint16(b)), nil
}

func (bin *Binary) Int16LE() (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return int16(leOrder.Uint16(b)), nil
}

func (bin *Binary) Int16BE() (int16, error) {
	b, err := bin.ByteStream.TakeExactly(2)
	if err != nil {
		return 0, err
	}
	return int16(beOrder.Uint16(b)), nil
}

func (bin *Binary) Int32(order binary.ByteOrder) (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return int32(order.Uint32(b)), nil
}

func (bin *Binary) Int32LE() (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return int32(leOrder.Uint32(b)), nil
}

func (bin *Binary) Int32BE() (int32, error) {
	b, err := bin.ByteStream.TakeExactly(4)
	if err != nil {
		return 0, err
	}
	return int32(beOrder.Uint32(b)), nil
}

func (bin *Binary) Int64(order binary.ByteOrder) (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return int64(order.Uint64(b)), nil
}

func (bin *Binary) Int64LE() (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return int64(leOrder.Uint64(b)), nil
}

func (bin *Binary) Int64BE() (int64, error) {
	b, err := bin.ByteStream.TakeExactly(8)
	if err != nil {
		return 0, err
	}
	return int64(beOrder.Uint64(b)), nil
}
