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

func Init(redisURL string) (error) {
	pool = newPool(redisURL)
	return nil
}

func Close() {
	pool.Close()
}

func Get(key string) (string, error) {
	c := pool.Get()
	defer c.Close()

	return redis.String(c.Do("GET", key))
}

func Set(key string, value interface{}) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, value)

	return err
}

func SetEx(key string, expire int, value interface{}) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("SETEX", key, expire, value)

	return err
}

func LPush(key string, value string) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("LPUSH", key, value)
	return err
}

func RPush(key string, value string) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("RPUSH", key, value)
	return err
}

func LRange(key string) ([]string, error) {
	c := pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("LRANGE", key, 0, -1))
}

func Pop(key string) (string, error) {
	c := pool.Get()
	defer c.Close()

	return redis.String(c.Do("LPOP", key))
}

func Incr(key string) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("INCR", key)
	return err
}

func IncrBy(key string, inc interface{}) (interface{}, error) {
	c := pool.Get()
	defer c.Close()

	return c.Do("INCRBY", key, inc)
}

func Expire(key string, seconds int) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("EXPIRE", key, seconds)

	return err
}

func Del(key string) (error) {
	c := pool.Get()
	defer c.Close()

	_, err := redis.Bool(c.Do("DEL", key))
	return err
}

func MGet(keys []string) ([]string, error) {
	c := pool.Get()
	defer c.Close()

	var args []interface{}
	for _, key := range keys {
		args = append(args, key)
	}

	return redis.Strings(c.Do("MGET", args...))
}

func ZAdd(key string, score float64, value interface{})  (int, error) {
	c:= pool.Get()
	defer c.Close()

	return redis.Int(c.Do("ZADD", key, score, value))
}

func ZCount(key string, min interface{}, max interface{}) (int, error) {
	c:= pool.Get()
	defer c.Close()

	return redis.Int(c.Do("ZCOUNT", key, min, max))
}
