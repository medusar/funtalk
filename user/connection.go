package user

import (
	"sync"
	"github.com/pkg/errors"
	"github.com/gorilla/websocket"
	"log"
)

type Connection struct {
	wsCon           *websocket.Conn
	inboundMsgChan  chan []byte
	outboundMsgChan chan []byte
	closeChan       chan byte
	closed          bool
	mutex           sync.Mutex
}

func InitConn(wsCon *websocket.Conn) (*Connection, error) {
	con := &Connection{
		wsCon:           wsCon,
		inboundMsgChan:  make(chan []byte, 1024),
		outboundMsgChan: make(chan []byte, 1024),
		closeChan:       make(chan byte, 1),
		closed:          false,
	}
	go con.readLoop()
	go con.writeLoop()
	return con, nil
}

func (c *Connection) ReadMessage() ([]byte, error) {
	select {
	case msg := <-c.inboundMsgChan:
		return msg, nil
	case <-c.closeChan:
		return nil, errors.New("connection is closed")
	}
}

func (c *Connection) WriteMessage(b []byte) error {
	select {
	case c.outboundMsgChan <- b:
		return nil
	case <-c.closeChan:
		return errors.New("connection is closed")
	}
}

func (c *Connection) Close() {
	if c.closed {
		return
	}
	//close a closeChan
	c.mutex.Lock()
	if !c.closed {
		c.wsCon.Close()
		close(c.closeChan)
		c.closed = true
	}
	c.mutex.Unlock()
}

func (c *Connection) RemoteAddr() string {
	return c.wsCon.RemoteAddr().String()
}

func (c *Connection) readLoop() {
	for {
		_, b, err := c.wsCon.ReadMessage()
		if err != nil {
			log.Println("error ws ReadMessage", err)
			c.Close()
			break
		}

		select {
		case c.inboundMsgChan <- b:
		case <-c.closeChan:
			c.Close()
			break
		}
	}
}

func (c *Connection) writeLoop() {
	for {
		select {
		case msg := <-c.outboundMsgChan:
			if err := c.wsCon.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("error ws WriteMessage", err)
				c.Close()
				break
			}
		case <-c.closeChan:
			c.Close()
			break
		}
	}
}
