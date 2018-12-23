package api

import (
	"net/http"
	"log"
	"github.com/medusar/funtalk/service"
	"crypto/md5"
	"fmt"
	"encoding/base64"
	"github.com/pkg/errors"
	"strings"
)

const CookieName = "sessionid"

var userService service.UserService

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

// cookie=base64(uid+"__"+hex(md5(uid__remoteAddr)))
func SetCookie(uid string, w http.ResponseWriter, r *http.Request) error {
	encoded := fmt.Sprintf("%x", md5.Sum([]byte(uid+"__"+r.RemoteAddr)))
	beforeEncrypt := uid + "__" + encoded
	//FIXME: encrypt before base64
	ckValue := base64.StdEncoding.EncodeToString([]byte(beforeEncrypt))
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: ckValue, MaxAge: 60 * 60 * 24})
	return nil
}

func ValidateCookie(cookie, address string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(cookie)
	if err != nil {
		return "", err
	}

	afterBase64 := string(b[:])

	//FIXME: decrypt

	split := strings.Split(afterBase64, "__")
	if len(split) != 2 {
		return "", errors.New("illegal cookie")
	}

	uid := split[0]

	encoded := fmt.Sprintf("%x", md5.Sum([]byte(uid+"__"+address)))
	if split[1] != encoded {
		return "", errors.New("cookie does not match")
	}

	return uid, nil
}

func ValidateHttpCookie(r *http.Request) (string, error) {
	ck, err := r.Cookie(CookieName)
	if err != nil {
		return "", err
	}
	return ValidateCookie(ck.Value, r.RemoteAddr)
}
