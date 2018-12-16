package talk

import (
	"log"
	"github.com/medusar/funtalk/message"
	"html"
)

var (
	//Event fired when a user connected
	UserOpenChan = make(chan *User, 100)
	//Event fired when a user should be closed
	UserCloseChan = make(chan *User, 100)
	//Event fired when a user login twice, the first will ke kicked out
	UserKickChan = make(chan *User, 10)

	//Contain all the users connected
	//key: user name, value: User
	UserMap = make(map[string]*User)
	//Message channel used to send messages to all the users
	MsgChan = make(chan *message.Message, 100)
)

// Get all the online user names
func OnlineList(users map[string]*User) []string {
	names := make([]string, 0, len(users))
	for n := range users {
		names = append(names, n)
	}
	return names
}

func CloseUser(u *User) {
	users := UserMap
	delete(users, u.Name())
	MsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " left room"}
	//Update online user list
	MsgChan <- &message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
	u.Close()
}

func AddUser(u *User) {
	users := UserMap

	oldUser := users[u.Name()]
	if oldUser != nil {
		//close old channel
		err := oldUser.Write(&message.Message{Type: message.Kick})
		if err != nil {
			//write failed
			//TODO:
			return
		}

		users[u.Name()] = u
		// Send welcome message to current user only
		u.Write(&message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"})
		// Update online user list to current user only
		u.Write(&message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)})
		return
	}

	users[u.Name()] = u
	MsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"}
	// Update online user list
	MsgChan <- &message.Message{Type: message.Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
}

func StartWsService() {
	go handleUser()
	go startMsgRouter()
}

func handleUser() {
	for {
		select {
		case u := <-UserOpenChan:
			AddUser(u)
		case du := <-UserCloseChan:
			CloseUser(du)
		}
	}
}

func startMsgRouter() {
	for msg := range MsgChan {
		for _, u := range UserMap {
			if u.Name() == msg.Sender {
				continue
			}
			if err := u.Write(msg); err != nil {
				log.Println("error write msg", err)
				//TODO:
			}
		}
	}
}

func Serve(user *User) {
	for {
		msg, err := user.Read()
		if err != nil {
			log.Println("error read message from user", err, user.Name())
			return
		}

		messageType := msg.Type
		if messageType != message.Auth && !user.HasAuthed() {
			log.Printf("unauthed msg:%s \n", messageType)
			return
		}

		switch messageType {
		case message.Auth:
			authParams := msg.Content.(map[string]interface{})
			if authParams != nil {
				username := authParams["username"].(string)
				if username != "" {
					user.SetName(username)
					//TODO
					UserOpenChan <- user
				}
			}
		case message.Chat:
			MsgChan <- &message.Message{Type: message.Chat, RoomId: "1", Sender: user.Name(), Content: html.EscapeString(msg.Content.(string))}
		case message.Ping:
			if err := user.Write(&message.Message{Type: message.Pong}); err != nil {
				log.Println("error write msg", err)
				//TODO:
				break
			}
		default:
			log.Println("unsupported message type:", messageType)
		}
	}
}
