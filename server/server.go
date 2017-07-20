/**
	server of the chat
 */
package server

import (
	"net"
	"log"
	"bufio"
	"github.com/medusar/funtalk/codec"
)

type Listener interface {
	OnMsg(msg codec.Msg)
}

type Server struct {
	addr *net.TCPAddr
	enc  codec.Encoder
	dec  codec.Decoder
	lns  []Listener
}

func (s *Server) Start() {
	tcp, lerr := net.ListenTCP("tcp", s.addr)
	if lerr != nil {
		log.Fatal("err listen tcp", lerr)
	}

	for {
		conn, error := tcp.Accept()
		if error != nil {
			log.Println("error acccept", error)
			continue
		}
		go s.serve(conn)
	}
}

func (s *Server) serve(conn net.Conn) {
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		bytes := make([]byte, 0)
		n, err := r.Read(bytes)
		if err != nil {
			log.Println("error reading", err)
			break
		}
		msg, dErr := s.dec.Decode(bytes)
		if dErr != nil {
			log.Println("error decode", dErr)
			continue
		}

		for _, ls := range s.lns {
			ls.OnMsg(msg)
		}
	}
}
