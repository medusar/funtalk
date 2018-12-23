package storage

import (
	"github.com/gomodule/redigo/redis"
	"time"
	"github.com/pkg/errors"
	"strconv"
)

var (
	pool        *redis.Pool
	redisServer = ":6379"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

type UserRedisStore struct {
	pool *redis.Pool
}

func NewUserRedisStore(redisAddr string) *UserRedisStore {
	return &UserRedisStore{pool: newPool(redisAddr)}
}

func (urs *UserRedisStore) GetRooms(id string) ([]string, error) {
	conn := urs.pool.Get()
	defer conn.Close()
	param := "user:" + id + ":rooms"
	return redis.Strings(conn.Do("SMEMBERS", param))
}

func (urs *UserRedisStore) AddRoom(uid string, rids ...string) error {
	conn := urs.pool.Get()
	defer conn.Close()
	params := make([]interface{}, len(rids)+1)
	params[0] = "user:" + uid + ":rooms"
	for i, v := range rids {
		params[i+1] = v
	}
	_, err := redis.Int(conn.Do("SADD", params...))
	return err
}

func (urs *UserRedisStore) RemoveRoom(uid string, rids ...string) error {
	conn := urs.pool.Get()
	defer conn.Close()
	params := make([]interface{}, len(rids)+1)
	params[0] = "user:" + uid + ":rooms"
	for i, v := range rids {
		params[i+1] = v
	}
	_, err := redis.Int(conn.Do("SREM", params...))
	return err
}

func (urs *UserRedisStore) Get(uid string) (*UserInfo, error) {
	conn := urs.pool.Get()
	defer conn.Close()

	hash, err := redis.StringMap(conn.Do("HGETALL", "user:"+uid+":info"))
	if err != nil {
		return nil, err
	}

	if len(hash) == 0 {
		return nil, errors.New("user not exists")
	}

	//TODO: use reflect
	u := &UserInfo{
		Uid:      uid,
		Name:     hash["name"],
		Password: hash["password"],
	}

	return u, nil
}

func (urs *UserRedisStore) Set(uid string, u *UserInfo) error {
	conn := urs.pool.Get()
	defer conn.Close()

	//TODO: use reflect
	params := make([]interface{}, 5)
	params[0] = "user:" + uid + ":info"
	params[1] = "name"
	params[2] = u.Name
	params[3] = "password"
	params[4] = u.Password

	_, err := redis.String(conn.Do("HMSET", params...))
	return err
}

func (urs *UserRedisStore) GenUid() (string, error) {
	conn := urs.pool.Get()
	defer conn.Close()

	id, err := redis.Int(conn.Do("INCR", "user_id:generator"))
	if err != nil {
		return "", err
	}
	return strconv.Itoa(id), nil
}