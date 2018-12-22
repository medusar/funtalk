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
	Ret    Type = "ret"
)

type Message struct {
	Id         string      `json:"id,omitempty"`
	Type       Type        `json:"type,omitempty"`
	RoomId     string      `json:"roomId,omitempty"`
	Sender     string      `json:"sender,omitempty"`
	SenderName string      `json:"senderName,omitempty"`
	Content    interface{} `json:"content,omitempty"`
}

var PONG = &Message{Type: Pong}
var KICK = &Message{Type: Kick}
var OK = &Message{Type: Ret, Content: "ok"}

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
