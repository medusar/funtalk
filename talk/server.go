package talk

import (
	"time"
	"github.com/gorilla/websocket"
	"log"
	"html"
	"errors"
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
	MsgChan = make(chan *Message, 100)
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
	MsgChan <- &Message{Type: Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " left room"}
	//Update online user list
	MsgChan <- &Message{Type: Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
	u.Close()
}

func AddUser(u *User) {
	users := UserMap

	oldUser := users[u.Name()]
	if oldUser != nil {
		//close old channel
		oldUser.MsgChan <- &Message{Type: Kick}
		users[u.Name()] = u
		// Send welcome message to current user only
		u.MsgChan <- &Message{Type: Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"}
		// Update online user list to current user only
		u.MsgChan <- &Message{Type: Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
		return
	}
	users[u.Name()] = u
	MsgChan <- &Message{Type: Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"}
	// Update online user list
	MsgChan <- &Message{Type: Online, RoomId: "1", Sender: "admin", Content: OnlineList(users)}
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
			//in case it blocks, we use select to set a time limit
			select {
			case u.MsgChan <- msg:
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
}

func Serve(user *User) {
	go read(user)
	go write(user)

	//if no error occurs, it will block here
	select {
	case e := <-user.ErrChan:
		log.Println("error chat", e)
		UserCloseChan <- user
	}

}

func write(user *User) {
	for {
		var err error

		select {
		case msg := <-user.MsgChan:
			err = user.Con.WriteMessage(websocket.TextMessage, []byte(msg.ToJson()))
		case <-time.After(15 * time.Second):
			err = user.Con.WriteMessage(websocket.PingMessage, nil)
		}

		if err != nil {
			if err != websocket.ErrCloseSent {
				log.Printf("error: %v, user: %v \n", err, user.Name())
			}
			user.ErrChan <- err
			return
		}
	}
}

func read(user *User) {
	for {
		//Can also use `user.Con.ReadJSON()`
		_, p, err := user.Con.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v, user-agent: %v \n", err, user.Name())
			}
			user.ErrChan <- err
			return
		}

		message, err := FromJson(p)
		if err != nil {
			log.Println("error unmarshalling json", err)
			continue
		}

		if message.Type != Auth && !user.HasAuthed() {
			log.Printf("unauthed user, message type:%v, user:%s", message.Type, user.String())
			user.ErrChan <- errors.New("should auth first!")
			return
		}

		switch message.Type {
		case Auth:
			//TODO: optimize
			authParams := message.Content.(map[string]interface{})
			if authParams != nil {
				username := authParams["username"].(string)
				if username != "" {
					user.SetName(username)
					UserOpenChan <- user
				}
			}
		case Chat:
			MsgChan <- &Message{Type: Chat, RoomId: "1", Sender: user.Name(), Content: html.EscapeString(message.Content.(string))}
		case Ping:
			user.MsgChan <- &Message{Type: Pong}
		default:
			log.Println("unsupported message type:", message.Type)
		}
	}
}
