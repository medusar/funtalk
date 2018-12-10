package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"html/template"
	"github.com/medusar/funtalk/talk"
	"fmt"
)

const LISTEN_ADDR = "localhost:8080"

var templates = template.Must(template.ParseFiles("html/ws.html"))

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
	err := templates.ExecuteTemplate(w, "ws.html", LISTEN_ADDR)
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

		talk.MsgChan <- &talk.Msg{UserName: user.Name(), Text: string(p[:])}
	}
}

func startServe() {
	users := talk.UserMap
	for {
		select {
		case u := <-talk.UserOpenChan:
			users[u.Name()] = u
			talk.MsgChan <- &talk.Msg{UserName: "admin", Text: u.Name() + " joined room, welcome!"}
		case du := <-talk.UserCloseChan:
			du.Con.Close()
			talk.MsgChan <- &talk.Msg{UserName: du.Name(), Text: "left room"}
			delete(users, du.Name())
		}
	}
}

func startMsgRouter() {
	for msg := range talk.MsgChan {
		for _, u := range talk.UserMap {
			if u.Name() == msg.UserName {
				continue
			}

			message := fmt.Sprintln("[" + msg.UserName + "]: " + msg.Text)
			u.MsgChan <- message
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
