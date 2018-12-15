package talk

import (
	"github.com/gorilla/websocket"
)

type User struct {
	Con     *websocket.Conn
	MsgChan chan *Message
}

func (u *User) Name() string {
	return u.Con.RemoteAddr().String()
}

func (u *User) Close() {
	u.Con.Close()
	close(MsgChan)
}
