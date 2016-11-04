package redis

type Client interface {
	Close()

	Get(key string) (string, error)
	MGet(keys []string) ([]string, error)
	Set(key string, value interface{}) error
	SetEx(key string, expire int, value interface{}) error
	Expire(key string, seconds int) error
	Del(key string) error

	LPush(key string, value string) error
	RPush(key string, value string) error
	LRange(key string) ([]string, error)
	LPop(key string) (string, error)
	Pop(key string) (string, error) // legacy LPop

	Incr(key string) error
	IncrBy(key string, inc interface{}) (interface{}, error)

	ZAdd(key string, score float64, value interface{}) (int, error)
	ZCount(key string, min interface{}, max interface{}) (int, error)
}