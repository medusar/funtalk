package talk

import (
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/medusar/funtalk/connection"
	"github.com/medusar/funtalk/message"
)

//User connected
type User struct {
	con  *connection.Connection
	name string
}

func InitUser(wsCon *websocket.Conn) (*User, error) {
	con, err := connection.Init(wsCon)
	if err != nil {
		return nil, err
	}
	user := &User{con: con,}
	return user, nil
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
	return fmt.Sprintf("address:%s, name:%v", u.con.RemoteAddr(), u.name)
}

func (u *User) Close() {
	u.con.Close()
}

func (u *User) Read() (*message.Message, error) {
	bytes, err := u.con.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message.FromJson(bytes)
}

func (u *User) Write(m *message.Message) error {
	return u.con.WriteMessage([]byte(m.ToJson()))
}
