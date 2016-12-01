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

	// _ = "OK", nil is possible only if NX or XX is used
	// Those should be a separate interface. It should return a boolean
	// if the operation took place or not.
	_, err := c.Do("SET", key, value)
	return err
}

func (pc *PooledClient) SetEx(key string, expire int64, value interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	// _ = "OK"
	_, err := c.Do("SETEX", key, expire, value)
	return err
}

func (pc *PooledClient) LPush(key string, value string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("LPUSH", key, value))
}

func (pc *PooledClient) RPush(key string, value string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("RPUSH", key, value))
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

func (pc *PooledClient) Incr(key string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("INCR", key))
}

func (pc *PooledClient) IncrBy(key string, inc int64) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("INCRBY", key, inc))
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

func (pc *PooledClient) SetNxEx(key string, value interface{}, expire int64) (int64, error) {
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

func (pc *PooledClient) Eval(script string, keyCount int) (interface{}, error) {
	c := pc.pool.Get()
	defer c.Close()

	redisScript := redis.NewScript(keyCount, script)

	return redisScript.Do(c)
}

func (pc *PooledClient) HDel(key string, fields ...string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HDEL", interfaceSlice(append([]string{key}, fields...))...))
}

func (pc *PooledClient) HExists(key string, field string) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do("HEXISTS", key, field))
}

func (pc *PooledClient) HGet(key string, field string) (string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.String(c.Do("HGET", key, field))
}

func (pc *PooledClient) HGetAll(key string) (map[string]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	keysValues, err := redis.Strings(c.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	return stringMap(keysValues), nil
}

func (pc *PooledClient) HLen(key string) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HLEN", key))
}

func (pc *PooledClient) HMGet(key string, fields ...string) (map[string]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	values, err := redis.Strings(c.Do("HMGET", interfaceSlice(append([]string{key}, fields...))...))
	if err != nil {
		return nil, err
	}

	return zipMap(fields, values), nil
}

func (pc *PooledClient) HKeys(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("HKEYS", key))
}

func (pc *PooledClient) HMSet(key string, fields map[string]interface{}) error {
	c := pc.pool.Get()
	defer c.Close()

	args := append(make([]interface{}, 0, len(fields)*2+1), key)
	for key, value := range fields {
		args = append(args, key, value)
	}

	_, err := redis.String(c.Do("HMSET", args...))
	return err
}

func (pc *PooledClient) HSet(key string, field string, value interface{}) (bool, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do("HSET", key, field, value))
}

func (pc *PooledClient) HVals(key string) ([]string, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do("HVALS", key))
}

func (pc *PooledClient) HIncrBy(key string, field string, inc int64) (int64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Int64(c.Do("HINCRBY", key, field, inc))
}

func (pc *PooledClient) HIncrByFloat(key string, field string, inc float64) (float64, error) {
	c := pc.pool.Get()
	defer c.Close()

	return redis.Float64(c.Do("HINCRBYFLOAT", key, field, inc))
}

func stringMap(keysValues []string) map[string]string {
	if len(keysValues)%2 == 1 {
		return nil
	}

	hash := make(map[string]string, len(keysValues)/2)
	for i := 0; i < len(keysValues); i += 2 {
		hash[keysValues[i]] = keysValues[i+1]
	}
	return hash
}

func zipMap(keys, values []string) map[string]string {
	if len(keys) != len(values) {
		return nil
	}

	hash := make(map[string]string, len(keys))
	for i := 0; i < len(keys); i++ {
		hash[keys[i]] = values[i]
	}
	return hash
}
