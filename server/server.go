package server

import (
	"log"
	"github.com/medusar/funtalk/message"
	"github.com/medusar/funtalk/user"
	"html"
)

var (
	userEventChan = make(chan *user.Event, 1024)
	//key: uid, value: User
	userMap = make(map[string]*user.User)
	roomMap = make(map[string]*Room)
	//Message channel used to send messages to all the users
	outboundMsgChan = make(chan *message.Message, 1024)
)

// Get all the online user names
func OnlineList(users map[string]*user.User) []string {
	names := make([]string, 0, len(users))
	for n := range users {
		names = append(names, n)
	}
	return names
}

func CloseUser(u *user.User) {
	users := userMap
	delete(users, u.Name())
	outboundMsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " left room"}
	//Update online user list
	outboundMsgChan <- &message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
	removeFromRoom(u.Uid())
	u.Close()
}

func AddUser(u *user.User) {
	users := userMap

	oldUser := users[u.Name()]
	if oldUser != nil {
		//close old channel
		oldUser.Write(message.KICK)
		oldUser.Close()

		users[u.Uid()] = u
		return
	}

	users[u.Uid()] = u
}

func addToRoom(rid string, u *user.User) {
	room, ok := roomMap[rid]
	if !ok {
		room = InitRoom(rid)
		roomMap[rid] = room
	}

	if _, ok := room.Users[u.Uid()]; ok {
		return
	}

	room.Users[u.Uid()] = true
	outboundMsgChan <- &message.Message{Type: message.Chat, RoomId: rid, Sender: "admin", Content: u.Name() + " joined room"}
	outboundMsgChan <- &message.Message{Type: message.Online, RoomId: rid, Sender: "admin", Content: OnlineList(userMap)}
}

func removeFromRoom(uid string) {
	for _, room := range roomMap {
		delete(room.Users, uid)
	}
}

func StartWsService() {
	go handleUser()
	go startMsgRouter()
}

func handleUser() {
	for ue := range userEventChan {
		switch ue.Type {
		case user.Authed:
			AddUser(ue.User)
		case user.Closed:
			CloseUser(ue.User)
		}
	}
}

func startMsgRouter() {
	for msg := range outboundMsgChan {
		if room, ok := roomMap[msg.RoomId]; ok {
			room.MsgChan <- msg
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
			authParams := msg.Content.(map[string]interface{})
			if authParams != nil {
				uid := authParams["uid"].(string)
				if uid != "" {
					u.SetUid(uid)
					u.SetName(authParams["name"].(string))
					userEventChan <- &user.Event{Type: user.Authed, User: u}
				} else {
					log.Println("error auth, close connection")
					break
				}
			}
		case message.Chat:
			sendToRoom(msg, u)
			//TODO:return to notify client a success
			//if err := u.Write(&message.Message{Type: message.Ret, Content: msg.Id}); err != nil {
			//	log.Println("error write msg", err)
			//	break
			//}
		case message.Ping:
			if err := u.Write(message.PONG); err != nil {
				log.Println("error write msg", err)
				break
			}
		default:
			log.Println("unsupported message type:", msgType)
		}
	}

	userEventChan <- &user.Event{Type: user.Closed, User: u}
}

func sendToRoom(msg *message.Message, u *user.User) {
	if msg.RoomId == "" {
		//for test
		msg.RoomId = "1"
	}
	//if user not in room, add to room
	//TODO: add room when login
	addToRoom(msg.RoomId, u)
	outboundMsgChan <- &message.Message{Type: message.Chat, RoomId: msg.RoomId, Sender: u.Name(), Content: html.EscapeString(msg.Content.(string))}
}
