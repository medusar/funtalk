package main

import (
	"github.com/medusar/funtalk/client"
	"github.com/medusar/funtalk/codec"
	"io"
	"log"
	"strconv"
	"time"
)

func main() {
	funClient := client.NewFunClient(":8888", func(reader io.Reader) codec.Codec {
		return codec.NewLenCodec(reader)
	}, []client.Listener{})

	e := funClient.Connect()
	if e != nil {
		log.Fatal("failed", e)
	}

	i := 1
	for i < 10 {
		resp, error := funClient.Send("how are you, data :" + strconv.Itoa(i))
		if error != nil {
			log.Println("error send", error)
			continue
		}

		log.Println("resp:", resp)
		i++
		time.Sleep(1 * time.Second)
	}
}
