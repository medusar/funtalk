package main

import (
	"github.com/medusar/funtalk/codec"
	"github.com/medusar/funtalk/server"
	"io"
	"log"
	"net"
)

type SimListener struct {
}

func (s *SimListener) OnCon(c *server.Chan) {
	log.Println("sl: oncon", c)
}
func (s *SimListener) OnMsg(c *server.Chan, p codec.Packet) {
	log.Println("sl: onmsg", c, p)
	c.Write(p)
}
func (s *SimListener) OnClose(c *server.Chan) {
	log.Println("sl: onclose", c)
}
func (s *SimListener) OnErr(c *server.Chan, e error) {
	log.Println("ls: onerr", c, e)
	c.Close()
}

func main() {
	addr, rerr := net.ResolveTCPAddr("tcp4", ":8888")
	if rerr != nil {
		log.Fatal("error resolve tcp addr", rerr)
	}

	funServer := server.NewServer(addr, []server.Listener{&SimListener{}}, func(r io.Reader) codec.Codec {
		//return codec.NewDelimCodec(r, '\n')
		return codec.NewLenCodec(r)
	})
	funServer.Start()
}
