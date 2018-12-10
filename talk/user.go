package talk

import (
	"github.com/gorilla/websocket"
)

type User struct {
	Con     *websocket.Conn
	MsgChan chan string
}

func (u *User) Name() string {
	return u.Con.RemoteAddr().String()
}
