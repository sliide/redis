package redis

import (
	"errors"
)

var client Client = nil

func SetClient(redisClient Client) {
	client = redisClient
}

func Init(connectionString string) error {
	client = NewClient(connectionString)
	return nil
}

func Close() {
	client.Close()
}

func Get(key string) (string, error) {
	return client.Get(key)
}

func MGet(keys []string) ([]string, error) {
	return client.MGet(keys...)
}

func Set(key string, value interface{}) error {
	return client.Set(key, value)
}

func SetEx(key string, expire int64, value interface{}) error {
	return client.SetEx(key, expire, value)
}

func Expire(key string, seconds int64) error {
	_, err := client.Expire(key, seconds)
	return err
}

func Del(key string) error {
	_, err := client.Del(key)
	return err
}

func LPush(key string, value string) error {
	_, err := client.LPush(key, value)
	return err
}

func RPush(key string, value string) error {
	_, err := client.RPush(key, value)
	return err
}

func LRange(key string) ([]string, error) {
	return client.LRange(key)
}

func LPop(key string) (string, error) {
	return client.LPop(key)
}

func Pop(key string) (string, error) {
	return client.LPop(key)
}

func Incr(key string) error {
	_, err := client.Incr(key)
	return err
}

func IncrBy(key string, inc interface{}) (interface{}, error) {
	increment, ok := NumberToInt64(inc)
	if !ok {
		return nil, errors.New("Increment must be convertible to int64")
	}
	return client.IncrBy(key, increment)
}

func ZAdd(key string, score float64, value interface{}) (int64, error) {
	return client.ZAdd(key, score, value)
}

func ZCount(key string, min interface{}, max interface{}) (int64, error) {
	return client.ZCount(key, min, max)
}

func SetNxEx(key string, value interface{}, expire int64) (int64, error) {
	return client.SetNxEx(key, value, expire)
}
