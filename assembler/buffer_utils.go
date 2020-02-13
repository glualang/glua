package assembler

import (
	"bytes"
	"encoding/binary"
	"math"
)

func BufferWriteBytes(buffer *bytes.Buffer, data []byte, size int) (err error) {
	for i := int(0); i < size; i++ {
		err = buffer.WriteByte(data[i])
	}
	return err
}

func BufferWriteCharArray(buffer *bytes.Buffer, data string) error {
	dataBytes := []byte(data)
	return BufferWriteBytes(buffer, dataBytes, len(dataBytes))
}

func BufferWriteString(buffer *bytes.Buffer, data string) error {
	dataBytes := []byte(data)
	strLen := len(data)
	if strLen < 0xFE {
		res := BufferWriteInt8(buffer, uint8(strLen+1))
		if res != nil {
			return res
		}
	} else {
		res := BufferWriteInt8(buffer, 0xFF)
		if res != nil {
			return res
		}
		res = BufferWriteUInt64(buffer, uint64(strLen+1))
		if res != nil {
			return res
		}
	}
	return BufferWriteBytes(buffer, dataBytes, len(dataBytes))
}

func BufferWriteInt8(buffer *bytes.Buffer, data uint8) error {
	return buffer.WriteByte(byte(data))
}

func BufferWriteUInt32(buffer *bytes.Buffer, data uint32) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, data)
	_, err := buffer.Write(bs)
	return err
}

func BufferWriteUInt64(buffer *bytes.Buffer, data uint64) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, data)
	_, err := buffer.Write(bs)
	return err
}

func BufferWriteInt64(buffer *bytes.Buffer, data int64) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(data))
	_, err := buffer.Write(bs)
	return err
}

func BufferWriteFloat32(buffer *bytes.Buffer, data float32) error {
	return BufferWriteUInt32(buffer, math.Float32bits(data))
}

func BufferWriteFloat64(buffer *bytes.Buffer, data float64) error {
	return BufferWriteUInt64(buffer, math.Float64bits(data))
}

func BufferWriteBool(buffer *bytes.Buffer, data bool) error {
	var n uint8
	if data {
		n = 1
	} else {
		n = 0
	}
	return BufferWriteInt8(buffer, n)
}
