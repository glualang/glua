package utils

import (
	"errors"
	"io"
)

/**
 * 可回溯的ByteReader
 */

type RepeatableByteReader interface {
	io.ByteReader
	Position() int
	Reset(position int) error
}

type simpleRepeatableByteReader struct {
	source   io.ByteReader // 原始数据流
	position int // 当前读取到的偏移量
	buf      []byte // 累计读取到的数据
}

func (reader *simpleRepeatableByteReader) ReadByte() (data byte, err error) {
	if reader.position < len(reader.buf) {
		data = reader.buf[reader.position]
		reader.position++
		return
	}
	data, err = reader.source.ReadByte()
	if err != nil {
		return
	}
	reader.buf = append(reader.buf, data)
	reader.position++
	return
}

func (reader *simpleRepeatableByteReader) Position() int {
	return reader.position
}

func (reader *simpleRepeatableByteReader) Reset(position int) (err error) {
	if position > reader.position {
		err = errors.New("simpleRepeatableByteReader can't reset to position out of range")
		return
	}
	reader.position = position
	return
}

func ByteReaderToRepeatable(source io.ByteReader) RepeatableByteReader {
	return &simpleRepeatableByteReader{
		source:   source,
		position: 0,
		buf:      make([]byte, 0),
	}
}
