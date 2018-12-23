package api

import (
	"net/http"
	"log"
)

const CookieName = "sessionid"

type UserApi struct {
}

func (userApi *UserApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//cookie, err := r.Cookie(CookieName)
	//if err != nil {
	//	w.WriteHeader(http.StatusUnauthorized)
	//	return
	//}
	//TODO: encrypt cookie
	//sessionid := cookie.Value


	method := r.Method
	requestURI := r.RequestURI
	log.Println("request uri:", requestURI, ", method:", method)

	r.ParseForm()
	//form := r.Form

}

type RoomApi struct {
}

func (roomApi *RoomApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
