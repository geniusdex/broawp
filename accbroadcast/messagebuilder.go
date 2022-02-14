package accbroadcast

import "encoding/binary"

type messageBuilder struct {
	buf []byte
}

func newMessageBuilder() *messageBuilder {
	return &messageBuilder{
		buf: make([]byte, 0, 1024),
	}
}

func (mb *messageBuilder) Bytes() []byte {
	return mb.buf
}

func (mb *messageBuilder) WriteByte(b byte) {
	mb.buf = append(mb.buf, b)
}

func (mb *messageBuilder) WriteUint16(val uint16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, val)
	mb.buf = append(mb.buf, buf...)
}

func (mb *messageBuilder) WriteUint32(val uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, val)
	mb.buf = append(mb.buf, buf...)
}

func (mb *messageBuilder) WriteString(str string) {
	mb.WriteUint16(uint16(len(str)))
	mb.buf = append(mb.buf, str...)
}
