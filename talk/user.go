package talk

import (
	"github.com/gorilla/websocket"
)

//User connected
type User struct {
	Con *websocket.Conn
	//Outbound essage channel
	MsgChan chan *Message
	ErrChan chan error
	name    string
}

func (u *User) Name() string {
	return u.name
}

func (u *User) SetName(name string) {
	u.name = name
}

// If user has authed, his name will be set
func (u *User) HasAuthed() bool {
	return u.name != ""
}

func (u *User) Close() {
	u.Con.Close()
	//FIXME:error happens here
	//close(u.MsgChan)
}
