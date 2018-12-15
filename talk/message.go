package talk

import (
	"encoding/json"
	"log"
)

type MessageType string

const (
	Chat   MessageType = "chat"
	Online MessageType = "online"
	Ping   MessageType = "ping"
	Pong   MessageType = "pong"
	Auth   MessageType = "auth"
	Kick   MessageType = "kick"
)

type Message struct {
	Type    MessageType
	RoomId  string
	Sender  string
	Content interface{}
}

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
