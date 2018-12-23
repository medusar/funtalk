package service

import (
	"github.com/medusar/funtalk/storage"
	"log"
	"math/rand"
	"time"
)

var userStore = storage.NewUserRedisStore(":6379")

func init() {
	rand.Seed(time.Now().Unix())
}

type UserService struct {
}

func (us *UserService) CheckPassword(uid, password string) bool {
	userInfo, err := userStore.Get(uid)
	if err != nil {
		log.Println("error Get by uid", err)
		return false
	}
	return userInfo.Password == password
}

func (us *UserService) GetName(uid string) (string, error) {
	userInfo, err := userStore.Get(uid)
	if err != nil {
		log.Println("error Get by uid", err)
		return "", err
	}
	return userInfo.Name, nil
}

func (us *UserService) GetRooms(uid string) []string {
	rooms, err := userStore.GetRooms(uid)
	if err != nil {
		log.Printf("error GetRooms, uid: %s, %v", uid, err)
		return make([]string, 0)
	}
	return rooms
}