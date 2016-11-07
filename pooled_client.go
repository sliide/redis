package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type PooledClient struct {
	pool redis.Pool
}

func NewClient(server string) Client {
	return &PooledClient{
		*newPool(server),
	}
}

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

func (pc *PooledClient) Close() {
	pc.pool.Close()
}

func (pc *PooledClient) Get(key string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("GET", key))
}

func (pc *PooledClient) Set(key string, value interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, value)
	return err
}

func (pc *PooledClient) SetEx(key string, expire int, value interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("SETEX", key, expire, value)
	return err
}

func (pc *PooledClient) LPush(key string, value string) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("LPUSH", key, value)
	return err
}

func (pc *PooledClient) RPush(key string, value string) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("RPUSH", key, value)
	return err
}

func (pc *PooledClient) LRange(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("LRANGE", key, 0, -1))
}

func (pc *PooledClient) Pop(key string) (string, error) {
	return pc.LPop(key)
}

func (pc *PooledClient) LPop(key string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("LPOP", key))
}

func (pc *PooledClient) Incr(key string) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("INCR", key)
	return err
}

func (pc *PooledClient) IncrBy(key string, inc interface{}) (interface{}, error) {
	c := pc.pool.Get()
	defer c.Close()

	return c.Do("INCRBY", key, inc)
}

func (pc *PooledClient) Expire(key string, seconds int) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("EXPIRE", key, seconds)
	return err
}

func (pc *PooledClient) Del(key string) error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := redis.Bool(c.Do("DEL", key))
	return err
}

func (pc *PooledClient) MGet(keys []string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	var args []interface{}
	for _, key := range keys {
		args = append(args, key)
	}

	return redis.Strings(c.Do("MGET", args...))
}

func (pc *PooledClient) ZAdd(key string, score float64, value interface{}) (int, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int(c.Do("ZADD", key, score, value))
}

func (pc *PooledClient) ZCount(key string, min interface{}, max interface{}) (int, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int(c.Do("ZCOUNT", key, min, max))
}
