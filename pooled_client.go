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

func (pc *PooledClient) SetEx(key string, expire int64, value interface{}) error {
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

func (pc *PooledClient) Expire(key string, seconds int64) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	result, err := redis.Int(c.Do("EXPIRE", key, seconds))
	return result == 1, err
}

func (pc *PooledClient) Del(keys ...string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("DEL", interfaceSlice(keys)...))
}

func (pc *PooledClient) MGet(keys ...string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("MGET", interfaceSlice(keys)...))
}

func (pc *PooledClient) ZAdd(key string, score float64, value interface{}) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("ZADD", key, score, value))
}

func (pc *PooledClient) ZCount(key string, min interface{}, max interface{}) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("ZCOUNT", key, min, max))
}
