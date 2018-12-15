package talk

import (
	"github.com/gorilla/websocket"
	"fmt"
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

func (u *User) String() string {
	return fmt.Sprintf("address:%s, name:%v", u.Con.RemoteAddr().String(), u.name)
}

func (u *User) Close() {
	u.Con.Close()
	//FIXME:error happens here
	//close(u.MsgChan)
}
