package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"html/template"
	"github.com/medusar/funtalk/talk"
	"html"
)

const LISTEN_ADDR = "localhost:8080"

var templates = template.Must(template.ParseFiles("html/chat.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	user := &talk.User{Con: conn, MsgChan: make(chan string, 100)}
	go serve(user)
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
	for msg := range user.MsgChan {
		err := user.Con.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("error writting", err)
			talk.UserCloseChan <- user
		}
	}
}

func read(user *talk.User) {
	for {
		_, p, err := user.Con.ReadMessage()
		if err != nil {
			log.Println("error reading", err)
			talk.UserCloseChan <- user
			break
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
			// no op
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
			du.Con.Close()
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

			u.MsgChan <- string(msg.ToJson())
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
