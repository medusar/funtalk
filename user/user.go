package user

import (
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/medusar/funtalk/message"
)

type EventType string

const (
	Authed EventType = "authed"
	Closed EventType = "closed"
)

type Event struct {
	User *User
	Type EventType
}

//User connected
type User struct {
	uid     string
	con     *Connection
	name    string
	roomIds map[string]bool
}

func InitUser(wsCon *websocket.Conn) (*User, error) {
	con, err := InitConn(wsCon)
	if err != nil {
		return nil, err
	}
	u := &User{con: con, roomIds: make(map[string]bool)}
	return u, nil
}

func (u *User) Name() string {
	return u.name
}

func (u *User) SetName(name string) {
	u.name = name
}

func (u *User) Uid() string {
	return u.uid
}

func (u *User) SetUid(uid string) {
	u.uid = uid
}

func (u *User) SetRoomIds(roomIds []string) {
	for _, rid := range roomIds {
		u.roomIds[rid] = true
	}
}

func (u *User) RoomIds() map[string]bool {
	return u.roomIds
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
