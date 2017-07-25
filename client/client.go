package client

import (
	"github.com/medusar/funtalk/codec"
	"io"
	"log"
	"net"
)

type Listener interface {
	OnMsg(client *FunClient, msg codec.Packet)
}

type logListener struct {
}

func (l *logListener) OnMsg(client *FunClient, msg codec.Packet) {
	log.Println("msg received:", msg)
}

var dftListener *logListener = &logListener{}

type codecFun func(reader io.Reader) codec.Codec

type FunClient struct {
	addr   *net.TCPAddr
	c      *net.TCPConn
	cntd   bool
	cdc    codec.Codec
	cdcFun codecFun
	ls     []Listener
}

func NewFunClient(addr string, cfun codecFun, ls []Listener) *FunClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic("error addr:" + err.Error())
	}
	if cfun == nil {
		panic("error cfun:nil")
	}
	if ls == nil || len(ls) == 0 {
		ls = []Listener{dftListener}
	}
	return &FunClient{addr: tcpAddr, cdcFun: cfun, ls: ls}
}

func (c *FunClient) Connect() error {
	con, e := net.DialTCP("tcp", nil, c.addr)
	if e != nil {
		return e
	}
	c.c = con
	c.cdc = c.cdcFun(c.c)
	c.cntd = true

	go c.startRead()

	return nil
}

//FIXME：start read
func (c *FunClient) startRead() {
	for c.cntd {
		data, e := c.cdc.Read()
		if e != nil {
			log.Fatal("error read", e)
		}
		pkt, e := c.cdc.Decode(data)
		if e != nil {
			log.Println("error decode", e)
			continue
		}

		go func() {
			for _, l := range c.ls {
				l.OnMsg(c, pkt)
			}
		}()
	}
}

func (c *FunClient) IsCntd() bool {
	return c.cntd
}

func (c *FunClient) Send(pkt codec.Packet) (codec.Packet, error) {
	if !c.cntd {
		panic("not connected yet")
	}

	data, e := c.cdc.Encode(pkt)
	if e != nil {
		return nil, e
	}

	_, err := c.c.Write(data)
	if err != nil {
		return nil, err
	}

	//FIXME:
	bytes, err := c.cdc.Read()
	if err != nil {
		return nil, err
	}

	resp, err := c.cdc.Decode(bytes)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
