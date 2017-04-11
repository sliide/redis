package redis

type Client interface {
	Close()

	Get(key string) (string, error)
	MGet(keys ...string) ([]string, error)
	Set(key string, value interface{}) error
	SetEx(key string, expire int64, value interface{}) error
	SetNxEx(key string, value interface{}, expire int64) (int64, error)
	Expire(key string, seconds int64) (bool, error)
	Del(keys ...string) (int64, error)
	Incr(key string) (int64, error)
	IncrBy(key string, inc int64) (int64, error)
	IncrByFloat(key string, inc float64) (float64, error)

	Eval(string, int) (interface{}, error)

	LPush(key string, value string) (int64, error)
	RPush(key string, value string) (int64, error)
	LRange(key string) ([]string, error)
	LPop(key string) (string, error)

	ZAdd(key string, score float64, value interface{}) (int64, error)
	ZCount(key string, min interface{}, max interface{}) (int64, error)

	SAdd(key string, members ...string) (int64, error)
	SMembers(key string) ([]string, error)

	HDel(key string, fields ...string) (int64, error)
	HExists(key string, field string) (bool, error)
	HGet(key string, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	HMGet(key string, fields ...string) (map[string]string, error)
	HLen(key string) (int64, error)
	HKeys(key string) ([]string, error)
	HVals(key string) ([]string, error)
	HScan(key string, pattern string) (map[string]string, error)
	HSet(key string, field string, value interface{}) (bool, error)
	HMSet(key string, fields map[string]interface{}) error
	HIncrBy(key string, field string, inc int64) (int64, error)
	HIncrByFloat(key string, field string, inc float64) (float64, error)
}
