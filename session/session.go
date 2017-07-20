package session

import (
	"github.com/medusar/funtalk/user"
	"net"
	"time"
)

type UserSession struct {
	u *user.FunUser
	c *net.Conn
	t time.Time
}

func (s *UserSession) User() *user.FunUser {
	return s.u
}

func (s *UserSession) Conn() *net.Conn {
	return s.c
}

func (s *UserSession) Time() time.Time {
	return s.t
}

type SessionContainer struct {
	sessions map[string]*UserSession
	rooms    map[string]*Room
}

func NewSessionContainer() *SessionContainer {
	return &SessionContainer{
		sessions: make(map[string]*UserSession),
		rooms:    make(map[string]*Room),
	}
}

func (c *SessionContainer) Sessions() map[string]*UserSession {
	return c.sessions
}

func (c *SessionContainer) Rooms() map[string]*Room {
	return c.rooms
}

func (c *SessionContainer) AddSession(s *UserSession) {
	id := s.u.Id
	if _, ok := c.sessions[id]; ok {

	} else {
		c.sessions[id] = s
	}
}
