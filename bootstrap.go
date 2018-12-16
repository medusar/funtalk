package main

import (
	"github.com/gorilla/websocket"
	"github.com/medusar/funtalk/talk"
	"html/template"
	"log"
	"net/http"
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

	user, err := talk.InitUser(conn)
	if err != nil {
		log.Println("error init user", err)
		return
	}

	go talk.Serve(user)
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

func main() {
	talk.StartWsService()
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/im", wsHandler)
	log.Fatal(http.ListenAndServe(LISTEN_ADDR, nil))
}
