package user

import "github.com/medusar/funtalk/common"

type Sex uint8

const (
	Unknown = Sex(0)
	Female
	Male
)

type FunUser struct {
	Id   string //uniq id
	Nick string //nick name
	Sex  Sex    //0
}

func NewUser(nick string, sex Sex) *FunUser {
	//TODO: check sex
	return &FunUser{
		Id:   common.NewIdStr(),
		Nick: nick,
		Sex:  sex,
	}
}
