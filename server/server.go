package server

import (
	"log"
	"github.com/medusar/funtalk/message"
	"github.com/medusar/funtalk/user"
	"github.com/medusar/funtalk/service"
	"github.com/pkg/errors"
)

var (
	userEventChan = make(chan *user.Event, 1024)
	//key: uid, value: User
	userMap = make(map[string]*user.User)
	roomMap = make(map[string]*Room)
)

func closeUser(u *user.User) {
	users := userMap
	delete(users, u.Uid())

	for rid := range u.RoomIds() {
		if room, ok := roomMap[rid]; ok {
			room.DelUser(u.Uid())
		}
	}

	u.Close()
}

func getOrCreateRoom(rid string) *Room {
	if room, ok := roomMap[rid]; ok {
		return room
	}
	room := InitRoom(rid)
	roomMap[rid] = room
	log.Printf("room created, rid:%s", rid)
	return room
}

func addUser(u *user.User) {
	users := userMap
	uid := u.Uid()

	if oldUser, ok := users[uid]; ok {
		//close old channel
		oldUser.Write(message.KICK)
		oldUser.Close()
	}
	users[uid] = u

	//add room info
	for roomId := range u.RoomIds() {
		room := getOrCreateRoom(roomId)
		room.AddUser(uid)
	}
}

func StartWsService() {
	go loopUserEvent()
}

func loopUserEvent() {
	for ue := range userEventChan {
		switch ue.Type {
		case user.Authed:
			addUser(ue.User)
		case user.Closed:
			closeUser(ue.User)
		}
	}
}

func Serve(u *user.User) {
	for {
		msg, err := u.Read()
		if err != nil {
			log.Println("error user Read", err, u.Name())
			break
		}

		msgType := msg.Type
		if msgType != message.Auth && !u.HasAuthed() {
			log.Printf("unauthed msg:%s \n", msgType)
			break
		}

		switch msgType {
		case message.Auth:
			if err := auth(msg, u); err != nil {
				log.Println("error auth", err)
				break
			}
			userEventChan <- &user.Event{Type: user.Authed, User: u}
			if err := u.Write(message.OK); err != nil {
				log.Println("error write msg", err)
				break
			}
		case message.Chat:
			if err := sendToRoom(msg, u); err != nil {
				log.Printf("error chat, uid:%s, err:%+v", u.Uid(), err)
				break
			}
			if err := u.Write(message.OK); err != nil {
				log.Println("error write msg", err)
				break
			}
		case message.Ping:
			if err := u.Write(message.PONG); err != nil {
				log.Println("error write msg", err)
				break
			}
		default:
			log.Println("unsupported message type:", msgType)
		}
	}

	// some error has occurred, close the channel
	userEventChan <- &user.Event{Type: user.Closed, User: u}
}

func auth(msg *message.Message, u *user.User) error {
	params, ok := msg.Content.(map[string]interface{})
	if !ok {
		return errors.New("error auth, illegal Content")
	}
	uid, ok := params["uid"].(string)
	if !ok || uid == "" {
		return errors.New("error auth, illegal params")
	}
	u.SetUid(uid)
	u.SetName(params["name"].(string))
	roomIds := service.GetRooms(uid)
	u.SetRoomIds(roomIds)
	return nil
}

func sendToRoom(msg *message.Message, u *user.User) error {
	rid := msg.RoomId
	if rid == "" {
		return errors.New("msg roomid is empty")
	}
	if _, ok := u.RoomIds()[rid]; !ok {
		return errors.New("user not in room:" + rid)
	}
	log.Printf("send msg to room: %v", msg)

	msg.Sender = u.Uid()
	msg.SenderName = u.Name()

	roomMap[rid].Send(msg)
	return nil
}
