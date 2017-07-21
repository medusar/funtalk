/**
server of the chat
*/
package server

import (
	"errors"
	"github.com/medusar/funtalk/codec"
	"io"
	"log"
	"net"
)

var (
	DecodeErr = errors.New("decode error")
)

type Chan struct {
	c     net.Conn
	alive bool
	codec codec.Codec
	s     *FunServer
}

func (c *Chan) Start() {
	for c.alive {
		bytes, err := c.codec.Read()
		if err != nil {
			log.Println("error read", err)
			c.s.OnErr(c, err)
			continue
		}

		data, error := c.codec.Decode(bytes)
		if error != nil {
			log.Println("error decode ", error)
			c.s.OnErr(c, error)
			continue
		}
		log.Println("data received:", data)

		go c.s.OnMsg(c, data)
	}
}

func (c *Chan) Close() {
	c.alive = false
	c.c.Close()
	c.s.OnClose(c)
}

type Listener interface {
	OnCon(c *Chan)
	OnMsg(c *Chan, p codec.Packet)
	OnClose(c *Chan)
	OnErr(c *Chan, e error)
}

type cdcFunc func(r io.Reader) codec.Codec

type FunServer struct {
	l      *net.TCPListener
	addr   *net.TCPAddr
	closed bool

	lns     []Listener
	cdcFunc cdcFunc
}

func NewServer(addr *net.TCPAddr, lns []Listener, cf cdcFunc) *FunServer {
	return &FunServer{
		addr:    addr,
		closed:  true,
		lns:     lns,
		cdcFunc: cf,
	}
}

func (s *FunServer) Start() {
	if !s.closed {
		panic("FunServer already started")
	}

	tcp, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		log.Fatal("err listen tcp", err)
	}

	s.closed = false
	s.l = tcp

	for !s.closed {
		conn, error := s.l.Accept()
		if error != nil {
			log.Println("error acccept", error)
			continue
		}
		go s.serve(conn)
	}
}

func (s *FunServer) Shutdown() {
	s.closed = true
	s.l.Close()
}

func (s *FunServer) Closed() bool {
	return s.closed
}

func (s *FunServer) OnCon(c *Chan) {
	for _, s := range s.lns {
		s.OnCon(c)
	}
}

func (s *FunServer) OnMsg(c *Chan, p codec.Packet) {
	for _, s := range s.lns {
		s.OnMsg(c, p)
	}
}

func (s *FunServer) OnClose(c *Chan) {
	for _, s := range s.lns {
		s.OnClose(c)
	}
}

func (s *FunServer) OnErr(c *Chan, e error) {
	for _, s := range s.lns {
		s.OnErr(c, e)
	}
}

func (s *FunServer) serve(conn net.Conn) {
	log.Println("new con accepted")

	ch := &Chan{
		c:     conn,
		alive: true,
		s:     s,
		codec: s.cdcFunc(conn),
	}

	s.OnCon(ch)

	ch.Start()
}
