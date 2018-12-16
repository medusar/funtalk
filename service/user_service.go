package service

import (
	"github.com/medusar/funtalk/storage"
	"log"
)

var userStore = storage.NewUserRedisStore(":6379")

func CheckUserPassrod(uid, password string) bool {
	userInfo, err := userStore.Get(uid)
	if err != nil {
		log.Println("error Get by uid", err)
		return false
	}
	return userInfo.Password == password
}

func GetName(uid string) (string, error) {
	userInfo, err := userStore.Get(uid)
	if err != nil {
		log.Println("error Get by uid", err)
		return "", err
	}
	return userInfo.Name, nil
}
