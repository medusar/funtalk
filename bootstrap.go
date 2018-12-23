package main

import (
	"github.com/gorilla/websocket"
	"github.com/medusar/funtalk/api"
	"github.com/medusar/funtalk/server"
	"github.com/medusar/funtalk/service"
	"github.com/medusar/funtalk/user"
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

var userService service.UserService

func checkOrigin(_ *http.Request) bool {
	return true
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	u, err := user.InitUser(conn)
	if err != nil {
		log.Println("error init user", err)
		return
	}

	go server.Serve(u)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := api.ValidateHttpCookie(r)
	if err != nil {
		log.Printf("error cookie: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	item := make(map[string]string)
	item["Uid"] = uid
	name, err := userService.GetName(uid)
	if err != nil {
		log.Printf("error GetName, uid:%s, %v", uid, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	item["Name"] = name
	err = templates.ExecuteTemplate(w, "chat.html", item)
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
	uid := form["uid"]
	if uid == nil || len(uid) == 0 || uid[0] == "" {
		http.Error(w, "illegal arguments", http.StatusBadRequest)
		return
	}
	password := form["password"]
	if password == nil || len(password) == 0 || password[0] == "" {
		http.Error(w, "illegal arguments", http.StatusBadRequest)
		return
	}

	if userService.CheckPassword(uid[0], password[0]) {
		err := api.SetCookie(uid[0], w, r)
		if err != nil {
			http.Error(w, "illegal arguments", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		//TODO: improve user experience
		http.Error(w, "illegal arguments", http.StatusBadRequest)
	}
}

func main() {
	server.StartWsService()

	mux := http.NewServeMux()

	mux.HandleFunc("/", pageHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/im", wsHandler)

	mux.Handle("/api/user/", &api.UserApi{})
	//mux.Handle("/api/user", &api.UserApi{})
	mux.Handle("/api/room/", &api.RoomApi{})

	log.Fatal(http.ListenAndServe(LISTEN_ADDR, mux))
}
