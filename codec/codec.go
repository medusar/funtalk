package codec

type Msg interface{}

type Encoder interface {
	Encode(msg Msg) ([]byte, error)
}
type Decoder interface {
	Decode([]byte) (Msg, error)
}
