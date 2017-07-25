package codec

import (
	"bufio"
	"bytes"
	"testing"
)

func TestLenCodec_Encode(t *testing.T) {

	codec := &LenCodec{}
	data, e := codec.Encode("hello")
	if e != nil {
		t.Fatal(e)
	}

	t.Log(data)
}

func TestDelimCodec_Decode(t *testing.T) {
	raw := "hello, what is your name ? i am 美杜莎"

	en := &LenCodec{}
	data, e := en.Encode(raw)
	if e != nil {
		t.Fatal(e)
	}

	t.Log(data)
	reader := bytes.NewReader(data)
	newReader := bufio.NewReader(reader)
	de := &LenCodec{
		r: newReader,
	}

	b, error := de.Read()
	if error != nil {
		t.Fatal("error", error)
	}

	pkt, e := de.Decode(b)
	if e != nil {
		t.Fatal("error decode", e)
	}

	d := pkt.(string)
	t.Log("data:", d)

	if d != raw {
		t.Fatal("encode and decode does not match")
	}
}

func TestLenCodec_Read(t *testing.T) {

}
