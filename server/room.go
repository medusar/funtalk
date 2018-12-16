package server

import (
	"github.com/medusar/funtalk/message"
	"log"
	"github.com/medusar/funtalk/user"
)

type Room struct {
	Id      string
	Users   map[string]bool
	MsgChan chan *message.Message
}

func InitRoom(rid string) *Room {
	room := &Room{Id: rid, Users: make(map[string]bool), MsgChan: make(chan *message.Message, 1024)}
	go roomMsgLoop(room)
	return room
}

func roomMsgLoop(room *Room) {
	//FIXME:stop loop when room is empty
	for msg := range room.MsgChan {
		for uid := range room.Users {
			if u, ok := userMap[uid]; ok {
				if u.Name() == msg.Sender {
					continue
				}
				if err := u.Write(msg); err != nil {
					log.Println("error write msg", err)
					userEventChan <- &user.Event{Type: user.Closed, User: u}
				}
			}
		}
	}
}
