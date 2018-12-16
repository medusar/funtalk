package message

import (
	"encoding/json"
	"log"
)

type Type string

const (
	Chat   Type = "chat"
	Online Type = "online"
	Ping   Type = "ping"
	Pong   Type = "pong"
	Auth   Type = "auth"
	Kick   Type = "kick"
)

type Message struct {
	Type    Type
	RoomId  string
	Sender  string
	Content interface{}
}

var PONG = &Message{Type: Pong}
var KICK = &Message{Type: Kick}

func FromJson(b []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(b, &m)
	return &m, err
}

func (m *Message) ToJson() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		log.Println("error marshalling Message", err)
		return []byte("")
	}
	return bytes
}
