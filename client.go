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

	LPush(key string, value string) (int64, error)
	RPush(key string, value string) (int64, error)
	LRange(key string) ([]string, error)
	LPop(key string) (string, error)

	Incr(key string) (int64, error)
	IncrBy(key string, inc int64) (int64, error)

	ZAdd(key string, score float64, value interface{}) (int64, error)
	ZCount(key string, min interface{}, max interface{}) (int64, error)
}
