package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// NewClient returns a new redis
func NewClient(server string) Client {
	return &pooledClient{
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

type pooledClient struct {
	pool redis.Pool
}

func (pc *pooledClient) Close() {
	pc.pool.Close()
}

func (pc *pooledClient) Ping() error {
	c := pc.pool.Get()
	defer c.Close()

	_, err := c.Do("PING")
	return err
}

func (pc *pooledClient) Get(key string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("GET", key))
}

func (pc *pooledClient) Set(key string, value interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	// _ = "OK", nil is possible only if NX or XX is used
	// Those should be a separate interface. It should return a boolean
	// if the operation took place or not.
	_, err := c.Do("SET", key, value)
	return err
}

func (pc *pooledClient) SetEx(key string, expire int64, value interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	// _ = "OK"
	_, err := c.Do("SETEX", key, expire, value)
	return err
}

func (pc *pooledClient) LPush(key string, value string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("LPUSH", key, value))
}

func (pc *pooledClient) RPush(key string, value string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("RPUSH", key, value))
}

func (pc *pooledClient) LRange(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("LRANGE", key, 0, -1))
}

func (pc *pooledClient) LPop(key string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("LPOP", key))
}

func (pc *pooledClient) Incr(key string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("INCR", key))
}

func (pc *pooledClient) IncrBy(key string, inc int64) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("INCRBY", key, inc))
}

func (pc *pooledClient) IncrByFloat(key string, inc float64) (float64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Float64(c.Do("INCRBYFLOAT", key, inc))
}

func (pc *pooledClient) Expire(key string, seconds int64) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	result, err := redis.Int(c.Do("EXPIRE", key, seconds))
	return result == 1, err
}

func (pc *pooledClient) Del(keys ...string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("DEL", redis.Args{}.AddFlat(keys)...))
}

func (pc *pooledClient) MGet(keys ...string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("MGET", redis.Args{}.AddFlat(keys)...))
}

func (pc *pooledClient) ZAdd(key string, score float64, value interface{}) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("ZADD", key, score, value))
}

func (pc *pooledClient) ZCount(key string, min interface{}, max interface{}) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("ZCOUNT", key, min, max))
}

func (pc *pooledClient) SAdd(key string, members ...string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("SADD", redis.Args{key}.AddFlat(members)...))
}

func (pc *pooledClient) SMembers(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("SMEMBERS", key))
}

func (pc *pooledClient) SetNxEx(key string, value interface{}, expire int64) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	// The normal redis SETNX command does not accept timeouts, For this reason
	// our implementation simply uses the normal SET command with NX(Not exists) and
	// EX(expire) options and mimics what would be returned by SETNX.
	val, err := c.Do("SET", key, value, "NX", "EX", expire)

	switch val {
	case "OK":
		return redis.Int64(int64(1), err)
	default:
		return redis.Int64(int64(0), err)
	}
}

func (pc *pooledClient) Eval(script string, keyCount int) (interface{}, error) {
	c := pc.pool.Get()
	defer c.Close()

	redisScript := redis.NewScript(keyCount, script)

	return redisScript.Do(c)
}

func (pc *pooledClient) HDel(key string, fields ...string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HDEL", redis.Args{key}.AddFlat(fields)...))
}

func (pc *pooledClient) HExists(key string, field string) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do("HEXISTS", key, field))
}

func (pc *pooledClient) HGet(key string, field string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("HGET", key, field))
}

func (pc *pooledClient) HGetAll(key string) (map[string]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.StringMap(c.Do("HGETALL", key))
}

func (pc *pooledClient) HLen(key string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HLEN", key))
}

func (pc *pooledClient) HMGet(key string, fields ...string) (map[string]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	values, err := redis.Strings(c.Do("HMGET", redis.Args{key}.AddFlat(fields)...))
	if err != nil {
		return nil, err
	}

	return zipMap(fields, values), nil
}

func (pc *pooledClient) HKeys(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("HKEYS", key))
}

func (pc *pooledClient) HMSet(key string, fields map[string]interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	args := append(make([]interface{}, 0, len(fields)*2+1), key)
	for key, value := range fields {
		args = append(args, key, value)
	}

	_, err := redis.String(c.Do("HMSET", args...))
	return err
}

func (pc *pooledClient) HSet(key string, field string, value interface{}) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do("HSET", key, field, value))
}

func (pc *pooledClient) HVals(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("HVALS", key))
}

func (pc *pooledClient) HScan(key string, pattern string) (map[string]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	if pattern == "" {
		pattern = "*"
	}

	cursor := "0"
	keysValues := make([]interface{}, 0)
	for {
		values, err := redis.Values(c.Do("HSCAN", key, cursor, "MATCH", pattern))
		if err != nil {
			return nil, err
		}
		keysValues = append(keysValues, values[1].([]interface{})...)
		cursor = string(values[0].([]byte))
		if cursor == "0" {
			break
		}
	}

	return redis.StringMap(keysValues, nil)
}

func (pc *pooledClient) HIncrBy(key string, field string, inc int64) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HINCRBY", key, field, inc))
}

func (pc *pooledClient) HIncrByFloat(key string, field string, inc float64) (float64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Float64(c.Do("HINCRBYFLOAT", key, field, inc))
}

func zipMap(keys, values []string) map[string]string {
	if len(keys) != len(values) {
		return nil
	}

	hash := make(map[string]string, len(keys))
	for i := range keys {
		hash[keys[i]] = values[i]
	}
	return hash
}
