package talk

var (
	UserOpenChan  = make(chan *User, 100)
	UserCloseChan = make(chan *User, 100)
	UserMap       = make(map[string]*User)
	MsgChan       = make(chan *Message, 100)
)

func OnlineList(users map[string]*User) []string {
	names := make([]string, 0, len(users))
	for n := range users {
		names = append(names, n)
	}
	return names
}
