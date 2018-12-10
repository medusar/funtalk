package talk

var (
	UserOpenChan  = make(chan *User, 100)
	UserCloseChan = make(chan *User, 100)
	UserMap       = make(map[string]*User)
	MsgChan       = make(chan *Msg, 100)
)

type Msg struct {
	UserName string
	Text     string
}
