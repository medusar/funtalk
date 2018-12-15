package main

import (
	"github.com/gorilla/websocket"
	"github.com/medusar/funtalk/talk"
	"html"
	"html/template"
	"log"
	"net/http"
	"time"
)

const LISTEN_ADDR = ":8080"

var templates = template.Must(template.ParseFiles("html/chat.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	return true
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn.SetPongHandler(pongHandler)
	user := &talk.User{Con: conn, MsgChan: make(chan *talk.Message, 100)}
	go serve(user)
}

func pongHandler(appData string) error {
	log.Println("pong received:", appData)
	return nil
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "chat.html", LISTEN_ADDR)
	if err != nil {
		log.Println(err)
	}
}

func serve(user *talk.User) {
	log.Println("user:", user)
	talk.UserOpenChan <- user
	go read(user)
	go write(user)
}

func write(user *talk.User) {
	for {
		var err error

		select {
		case msg := <-user.MsgChan:
			err = user.Con.WriteMessage(websocket.TextMessage, []byte(msg.ToJson()))
		case <-time.After(15 * time.Second):
			err = user.Con.WriteMessage(websocket.PingMessage, nil)
		}

		if err != nil {
			if err != websocket.ErrCloseSent {
				log.Printf("error: %v, user: %v \n", err, user.Name())
			}
			talk.UserCloseChan <- user
			return
		}
	}
}

func read(user *talk.User) {
	for {
		_, p, err := user.Con.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v, user-agent: %v \n", err, user.Name())
			}
			talk.UserCloseChan <- user
			return
		}

		message, err := talk.FromJson(p)
		if err != nil {
			log.Println("error unmarshalling json", err)
			continue
		}

		switch message.Type {
		case talk.Chat:
			talk.MsgChan <- &talk.Message{Type: talk.Chat, RoomId: "1", Sender: user.Name(), Content: html.EscapeString(message.Content.(string))}
		case talk.Ping:
			user.MsgChan <- &talk.Message{Type: talk.Pong}
		default:
			log.Println("unsupported message type:", message.Type)
		}
	}
}

func startServe() {
	users := talk.UserMap
	for {
		select {
		case u := <-talk.UserOpenChan:
			users[u.Name()] = u
			talk.MsgChan <- &talk.Message{Type: talk.Chat, RoomId: "1", Sender: "admin", Content: u.Name() + " joined room"}
			talk.MsgChan <- &talk.Message{Type: talk.Online, RoomId: "1", Sender: "admin", Content: talk.OnlineList(users)}
		case du := <-talk.UserCloseChan:
			du.Close()
			delete(users, du.Name())
			talk.MsgChan <- &talk.Message{Type: talk.Chat, RoomId: "1", Sender: "admin", Content: du.Name() + " left room"}
			talk.MsgChan <- &talk.Message{Type: talk.Online, RoomId: "1", Sender: "admin", Content: talk.OnlineList(users)}
		}
	}
}

func startMsgRouter() {
	for msg := range talk.MsgChan {
		for _, u := range talk.UserMap {
			if u.Name() == msg.Sender {
				continue
			}

			//in case it blocks, we use select to set a time limit
			select {
			case u.MsgChan <- msg:
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
}

func main() {
	go startServe()
	go startMsgRouter()
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/im", wsHandler)
	log.Fatal(http.ListenAndServe(LISTEN_ADDR, nil))
}
