package main

import (
	"github.com/medusar/funtalk/codec"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	c, error := net.Dial("tcp", ":8888")
	if error != nil {
		log.Fatal("error", error)
	}

	lc := codec.NewLenCodec(c)

	i := 1
	for i < 100 {
		data, err := lc.Encode("hello" + strconv.Itoa(i))
		if err != nil {
			log.Println("error encode", err)
			continue
		}

		c.Write(data)
		i++
		time.Sleep(1 * time.Second)
	}
}
