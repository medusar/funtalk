package talk

import (
	"log"
	"github.com/medusar/funtalk/message"
	"html"
	"github.com/medusar/funtalk/user"
)

var (
	userEventChan = make(chan *user.Event, 1024)
	//key: user name, value: User
	userMap = make(map[string]*user.User)
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
	u.Close()
}

func AddUser(u *user.User) {
	users := userMap

	oldUser := users[u.Name()]
	if oldUser != nil {
		//close old channel
		oldUser.Write(message.KICK)
		oldUser.Close()

		users[u.Name()] = u
		// Send welcome message to current user only
		u.Write(&message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"})
		// Update online user list to current user only
		u.Write(&message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)})
		return
	}

	users[u.Name()] = u
	outboundMsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"}
	// Update online user list
	outboundMsgChan <- &message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
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
		for _, u := range userMap {
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
				username := authParams["username"].(string)
				if username != "" {
					u.SetName(username)
					userEventChan <- &user.Event{Type: user.Authed, User: u}
				}
			}
		case message.Chat:
			outboundMsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: u.Name(), Content: html.EscapeString(msg.Content.(string))}
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
