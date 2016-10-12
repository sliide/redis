package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var pool *redis.Pool = nil

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func Init(redisURL string) (err error) {
	pool = newPool(redisURL)
	return nil
}

func Close() {
	pool.Close()
}

func Get(key string) (val string, err error) {
	c := pool.Get()
	defer c.Close()

	val, err = redis.String(c.Do("GET", key))
	return
}

func Set(key string, value interface{}) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("SET", key, value)

	return
}

func LPush(key string, value string) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("LPUSH", key, value)
	return
}

func RPush(key string, value string) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("RPUSH", key, value)
	return
}

func LRange(key string) (vals []string, err error) {
	c := pool.Get()
	defer c.Close()

	vals, err = redis.Strings(c.Do("LRANGE", key, 0, -1))
	return vals, err
}

func Pop(key string) (val string, err error) {
	c := pool.Get()
	defer c.Close()

	val, err = redis.String(c.Do("LPOP", key))
	return
}

func Incr(key string) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("INCR", key)
	return
}

func IncrBy(key string, inc interface{}) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("INCRBY", key, inc)
	return
}

func Expire(key string, seconds int) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = c.Do("EXPIRE", key, seconds)

	return err
}

func Del(key string) (err error) {
	c := pool.Get()
	defer c.Close()

	_, err = redis.Bool(c.Do("DEL", key))
	return err
}

func MGet(keys string) ([]string, error) {
	c := pool.Get()
	defer c.Close()

	var args []interface{}
	for _, key := range keys {
		args = append(args, key)
	}

	return redis.Strings(c.Do("MGET", args...))
}
