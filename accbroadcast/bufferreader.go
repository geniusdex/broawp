package accbroadcast

import (
	"encoding/binary"
	"math"
)

type bufferReader struct {
	buf []byte
	pos int
}

func newBufferReader(buf []byte) *bufferReader {
	return &bufferReader{
		buf: buf,
		pos: 0,
	}
}

func (br *bufferReader) ReadByte() byte {
	val := br.buf[br.pos]
	br.pos++
	return val
}

func (br *bufferReader) ReadBool() bool {
	return br.ReadByte() > 0
}

func (br *bufferReader) ReadUint16() uint16 {
	val := binary.LittleEndian.Uint16(br.buf[br.pos : br.pos+2])
	br.pos += 2
	return val
}

func (br *bufferReader) ReadUint32() uint32 {
	val := binary.LittleEndian.Uint32(br.buf[br.pos : br.pos+4])
	br.pos += 4
	return val
}

func (br *bufferReader) ReadFloat32() float32 {
	return math.Float32frombits(br.ReadUint32())
}

func (br *bufferReader) ReadString() string {
	length := int(br.ReadUint16())
	val := br.buf[br.pos : br.pos+length]
	br.pos += length
	return string(val)
}
