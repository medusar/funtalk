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

var templates = template.Must(template.ParseFiles("html/chat.html", "html/login.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(_ *http.Request) bool {
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

func pongHandler(_ string) error {
	return nil
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	username := cookie.Value
	err = templates.ExecuteTemplate(w, "chat.html", username)
	if err != nil {
		log.Println(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "illegal method:"+r.Method, http.StatusBadRequest)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Println("error parsing form", err)
		http.Error(w, "illegal form", http.StatusBadRequest)
		return
	}

	form := r.Form
	username := form["username"]
	if username == nil || len(username) == 0 || username[0] == "" {
		http.Error(w, "illegal arguments", http.StatusBadRequest)
		return
	}
	password := form["password"]
	if password == nil || len(password) == 0 || password[0] == "" {
		http.Error(w, "illegal arguments", http.StatusBadRequest)
		return
	}

	if checkLogin(username[0], password[0]) {
		http.SetCookie(w, &http.Cookie{Name: "username", Value: username[0], MaxAge: 24 * 60 * 60})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		//TODO: improve user experience
		http.Error(w, "illegal arguments", http.StatusBadRequest)
	}
}

func checkLogin(username, password string) bool {
	//TODO: check username and password
	return true
}

func serve(user *talk.User) {
	go read(user)
	go write(user)

	//if no error occurs, it will block here
	for e := <-user.ErrChan; e != nil; {
		log.Println("error chat", e)
	}

	talk.UserCloseChan <- user
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
			user.ErrChan <- err
			return
		}
	}
}

func read(user *talk.User) {
	for {
		//Can also use `user.Con.ReadJSON()`
		_, p, err := user.Con.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v, user-agent: %v \n", err, user.Name())
			}
			user.ErrChan <- err
			return
		}

		message, err := talk.FromJson(p)
		if err != nil {
			log.Println("error unmarshalling json", err)
			continue
		}

		switch message.Type {
		case talk.Auth:
			//TODO: optimize
			authParams := message.Content.(map[string]interface{})
			if authParams != nil {
				username := authParams["username"].(string)
				if username != "" {
					user.SetName(username)
					talk.UserOpenChan <- user
				}
			}
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
	for {
		select {
		case u := <-talk.UserOpenChan:
			talk.AddUser(u)
		case du := <-talk.UserCloseChan:
			talk.CloseUser(du)
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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/im", wsHandler)
	log.Fatal(http.ListenAndServe(LISTEN_ADDR, nil))
}
