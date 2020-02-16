package utils

import "io"

// simple memory byte stream implementation

type ByteStream interface {
	io.Writer
	WriteByte(b byte) error
	WriteString(str string) error
	ToBytes() []byte
}

type simpleByteStream struct {
	buffer []byte
}

func NewSimpleByteStream() *simpleByteStream {
	return &simpleByteStream{}
}

func (stream *simpleByteStream) WriteByte(b byte) (err error) {
	_, err = stream.Write([]byte{b})
	return
}

func (stream *simpleByteStream) Write(data []byte) (n int, err error) {
	oldBufferLen := len(stream.buffer)
	newBuf := make([]byte, oldBufferLen + len(data))
	for i, b := range stream.buffer {
		newBuf[i] = b
	}
	for i, b := range data {
		newBuf[i+oldBufferLen] = b
	}
	stream.buffer = newBuf
	n = len(data)
	return
}

func (stream *simpleByteStream) WriteString(str string) (err error) {
	_, err = stream.Write([]byte(str))
	return err
}

func (stream *simpleByteStream) ToBytes() []byte {
	return stream.buffer[:]
}


