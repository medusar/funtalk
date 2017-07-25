package codec

import (
	"bufio"
	"io"
)

type Packet interface{}

type Encoder interface {
	Encode(pkt Packet) ([]byte, error)
}
type Decoder interface {
	Decode([]byte) (Packet, error)
	Read() ([]byte, error)
}

type Codec interface {
	Decoder
	Encoder
}

type delimCodec struct {
	delim byte
	r     *bufio.Reader
}

func NewDelimCodec(r io.Reader, delim byte) *delimCodec {
	return &delimCodec{
		delim: delim,
		r:     bufio.NewReader(r),
	}
}

func (d *delimCodec) Read() ([]byte, error) {
	return d.r.ReadSlice(d.delim)
}

func (d *delimCodec) Encode(pkt Packet) ([]byte, error) {
	ok := pkt.(string)
	return []byte(ok), nil
}

func (d *delimCodec) Decode(b []byte) (Packet, error) {
	return string(b[:]), nil
}

type PacketType byte

const (
	Msg = 1
)

// header
// |1|4|1|
// |version|opaque|type|
type FunPacketHeader struct {
	Version byte
	Opaque  int32
	Type    PacketType
}

type FunPacket struct {
	Header *FunPacketHeader
	Body   []byte
}
