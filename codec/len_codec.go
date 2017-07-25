/**
length based codec

|4byte length|body|
*/
package codec

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

const LEN_FIELD = 4

var (
	LEN_ERR = errors.New("error length")
)

type LenCodec struct {
	r *bufio.Reader
}

func NewLenCodec(r io.Reader) *LenCodec {
	return &LenCodec{r: bufio.NewReader(r)}
}

func (d *LenCodec) Read() ([]byte, error) {
	lb := make([]byte, LEN_FIELD)
	l, e := d.r.Read(lb)

	if e != nil {
		return nil, e
	}
	if l < LEN_FIELD {
		return nil, LEN_ERR
	}

	ln := binary.BigEndian.Uint32(lb)
	log.Println(ln)
	body := make([]byte, ln)
	n, err := d.r.Read(body)

	if err != nil {
		return nil, err
	}

	if n != int(ln) {
		return nil, LEN_ERR
	}

	return body, nil
}

func (d *LenCodec) Encode(pkt Packet) ([]byte, error) {
	str := pkt.(string)
	body := []byte(str)
	ln := len(body)

	lb := make([]byte, LEN_FIELD)

	binary.BigEndian.PutUint32(lb, uint32(ln))

	bytes := append(lb, body...)
	return bytes, nil
}

func (d *LenCodec) Decode(b []byte) (Packet, error) {
	return string(b[:]), nil
}
