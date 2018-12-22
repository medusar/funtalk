package api

import (
	"net/http"
	"log"
)

type UserApi struct {
}

func (userApi *UserApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	requestURI := r.RequestURI
	log.Println("request uri:", requestURI, ", method:", method)

	r.ParseForm()

	form := r.Form
	log.Printf("request form: %+v", form)

}

type RoomApi struct {
}

func (roomApi *RoomApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
