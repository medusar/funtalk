package storage

type UserInfo struct {
	Uid      string
	Name     string
	Password string
}

type UserStore interface {
	GetRooms(uid string) ([]string, error)
	AddRoom(uid string, rids ...string) error
	RemoveRoom(uid string, rids ...string) error
	Get(uid string) (*UserInfo, error)
	Set(uid string, u *UserInfo) error
	GenUid() (string, error)
}
