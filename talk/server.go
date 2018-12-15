package talk

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
