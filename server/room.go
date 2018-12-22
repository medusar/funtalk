package server

import (
	"github.com/medusar/funtalk/message"
	"log"
	"github.com/medusar/funtalk/user"
	"time"
)

type Room struct {
	// room id
	id string
	// online users of this room
	users map[string]bool
	// messages to be sent to all the users
	msgChan chan *message.Message
	// close
	closeChan chan byte
	// how many times the users has been empty
	usersEmptyTimes uint8
}

func InitRoom(rid string) *Room {
	room := &Room{id: rid,
		users: make(map[string]bool),
		msgChan: make(chan *message.Message, 1024),
		closeChan: make(chan byte),
	}
	go room.loopRoomMsg()
	go room.tickTick()
	return room
}

func (r *Room) AddUser(uid string) {
	r.users[uid] = true
	r.updateOnlineList()
}

func (r *Room) DelUser(uid string) {
	delete(r.users, uid)
	r.updateOnlineList()
}

func (r *Room) Id() string {
	return r.id
}

func (r *Room) Send(msg *message.Message) {
	select {
	case r.msgChan <- msg:
		log.Printf("msg sent to room:%v", msg)
	case <-time.After(100 * time.Millisecond):
		log.Printf("msg discard because of timeout, %+v", msg)
	}
}

func (r *Room) loopRoomMsg() {
	for {
		select {
		case msg := <-r.msgChan:
			log.Printf("msg send:%+v", msg)
			for uid := range r.users {
				if u, ok := userMap[uid]; ok {
					if u.Uid() == msg.Sender {
						continue
					}
					if err := u.Write(msg); err != nil {
						log.Println("error write msg", err)
						userEventChan <- &user.Event{Type: user.Closed, User: u}
					}
				}
			}
		case <-r.closeChan:
			break
		}
	}
}

// trigger every n seconds
func (r *Room) tickTick() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
		close(r.closeChan)
	}()

	for range ticker.C {
		//check if the room is empty
		if len(r.users) > 0 {
			r.usersEmptyTimes = uint8(0)
		} else {
			r.usersEmptyTimes++
		}

		if r.usersEmptyTimes >= 5 {
			log.Printf("room empty for 5 ticks, close, rid:%s", r.id)
			break
		}
	}
}

func (r *Room) updateOnlineList() {
	uids := make([]string, 0, len(r.users))
	for uid := range r.users {
		uids = append(uids, uid)
	}
	r.Send(&message.Message{Type: message.Online, RoomId: r.id, Sender: "admin", Content: uids})
}
