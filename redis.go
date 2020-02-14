package redis

import "errors"

var client Client = NewMemoryClient()

// DefaultClient returns the default client
func DefaultClient() Client {
	return client
}

// SetDefaultClient sets the default client to the given one
func SetDefaultClient(redisClient Client) {
	client = redisClient
}

// Init the Redis client which connects to the given endpoint and set to default client
func Init(connectionString string) error {
	client = NewClient(connectionString)
	return nil
}

// Close the client connection
// NOTE: Please do not use this function due to complex coupling
func Close() {
	client.Close()
}

// Get return a value from the storage by a given key
// see https://redis.io/commands/SetEx for the details
// NOTE: Please do not use this function due to complex coupling
func Get(key string) (string, error) {
	return client.Get(key)
}

// MGet returns values from the storage by given keys
// see https://redis.io/commands/MGet for the details
// NOTE: Please do not use this function due to complex coupling
func MGet(keys []string) ([]string, error) {
	return client.MGet(keys...)
}

// Set key to hold the value in the storage
// see https://redis.io/commands/Set for the details
// NOTE: Please do not use this function due to complex coupling
func Set(key string, value interface{}) error {
	return client.Set(key, value)
}

// SetEx which set key to hold the value and
// set key to timeout after a given number of seconds in the storage
// see https://redis.io/commands/SetEx for the details
// NOTE: Please do not use this function due to complex coupling
func SetEx(key string, expire int64, value interface{}) error {
	return client.SetEx(key, expire, value)
}

// Expire sets key to timeout after a given number of seconds
// see https://redis.io/commands/Expire for the details
// NOTE: Please do not use this function due to complex coupling
func Expire(key string, seconds int64) error {
	_, err := client.Expire(key, seconds)
	return err
}

// Del deleted the key in the storage
// see https://redis.io/commands/Del for the details
// NOTE: Please do not use this function due to complex coupling
func Del(key string) error {
	_, err := client.Del(key)
	return err
}

// LPush which insert the given value at the head of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operations.
// see https://redis.io/commands/LPush for the details
// NOTE: Please do not use this function due to complex coupling
func LPush(key string, value string) error {
	_, err := client.LPush(key, value)
	return err
}

// RPush which insert the given value at the trail of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operations.
// see https://redis.io/commands/RPush for the details
// NOTE: Please do not use this function due to complex coupling
func RPush(key string, value string) error {
	_, err := client.RPush(key, value)
	return err
}

// LRange returns the specified elements of the list stored at key.
// see https://redis.io/commands/LRange for the details
// NOTE: Please do not use this function due to complex coupling
func LRange(key string) ([]string, error) {
	return client.LRange(key)
}

// LPop removes and returns the first element of the list stored at key.
// see https://redis.io/commands/LPop for the details
// NOTE: Please do not use this function due to complex coupling
func LPop(key string) (string, error) {
	return client.LPop(key)
}

// Pop removes and returns the first element of the list stored at key.
// NOTE: Please do not use this function due to complex coupling
func Pop(key string) (string, error) {
	return client.LPop(key)
}

// Incr the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// see https://redis.io/commands/Incr for the details
// NOTE: Please do not use this function due to complex coupling
func Incr(key string) error {
	_, err := client.Incr(key)
	return err
}

// IncrBy the number stored at key by a given value `inc`.
// If the key does not exist, it is set to 0 before performing the operation.
// NOTE: Please do not use this function due to complex coupling
func IncrBy(key string, inc interface{}) (interface{}, error) {
	increment, ok := numberToInt64(inc)
	if !ok {
		return nil, errors.New("Increment must be convertible to int64")
	}
	return client.IncrBy(key, increment)
}

// ZAdd
// see https://redis.io/commands/ZAdd for the details
// NOTE: Please do not use this function due to complex coupling
func ZAdd(key string, score float64, value interface{}) (int64, error) {
	return client.ZAdd(key, score, value)
}

// ZCount
// see https://redis.io/commands/ZCount for the details
// NOTE: Please do not use this function due to complex coupling
func ZCount(key string, min interface{}, max interface{}) (int64, error) {
	return client.ZCount(key, min, max)
}

// SAdd
// see https://redis.io/commands/SAdd for the details
// NOTE: Please do not use this function due to complex coupling
func SAdd(key string, members ...string) (int64, error) {
	return client.SAdd(key, members...)
}

// SMembers
// see https://redis.io/commands/SMembers for the details
// NOTE: Please do not use this function due to complex coupling
func SMembers(key string) ([]string, error) {
	return client.SMembers(key)
}

// SetNxEx
// see https://redis.io/commands/SetNxEx for the details
// NOTE: Please do not use this function due to complex coupling
func SetNxEx(key string, value interface{}, expire int64) (int64, error) {
	return client.SetNxEx(key, value, expire)
}

// Eval evaluates scripts using the Lua interpreter built into Redis
// see https://redis.io/commands/Eval for the details
// NOTE: Please do not use this function due to complex coupling
func Eval(script string, keyCount int) (interface{}, error) {
	return client.Eval(script, keyCount)
}

// HDel removes the given fields from the hash stored at key.
// see https://redis.io/commands/HDel for the details
// NOTE: Please do not use this function due to complex coupling
func HDel(key string, fields ...string) (int64, error) {
	return client.HDel(key, fields...)
}

// HExists returns if field is an existing field in the hash stored at key.
// see https://redis.io/commands/HExists for the details
// NOTE: Please do not use this function due to complex coupling
func HExists(key string, field string) (bool, error) {
	return client.HExists(key, field)
}

// HGet returns the value associated with field in the hash stored at key.
// see https://redis.io/commands/HGet for the details
// NOTE: Please do not use this function due to complex coupling
func HGet(key string, field string) (string, error) {
	return client.HGet(key, field)
}

// HGetAll returns all fields and values of the hash stored at key.
// see https://redis.io/commands/HGetAll for the details
// NOTE: Please do not use this function due to complex coupling
func HGetAll(key string) (map[string]string, error) {
	return client.HGetAll(key)
}

// HLen returns the number of fields contained in the hash stored at key.
// see https://redis.io/commands/HLen for the details
// NOTE: Please do not use this function due to complex coupling
func HLen(key string) (int64, error) {
	return client.HLen(key)
}

// HMGet returns the values associated with the given fields in the hash stored at key.
// see https://redis.io/commands/HMGet for the details
// NOTE: Please do not use this function due to complex coupling
func HMGet(key string, fields ...string) (map[string]string, error) {
	return client.HMGet(key, fields...)
}

// HKeys returns all field names in the hash stored at key.
// see https://redis.io/commands/HKeys for the details
// NOTE: Please do not use this function due to complex coupling
func HKeys(key string) ([]string, error) {
	return client.HKeys(key)
}

// HMSet sets the specified fields to their respective values in the hash stored at key.
// see https://redis.io/commands/HMSet for the details
// NOTE: Please do not use this function due to complex coupling
func HMSet(key string, fields map[string]interface{}) error {
	return client.HMSet(key, fields)
}

// HSet sets field in the hash stored at key to value.
// If key does not exist, a new key holding a hash is created.
// see https://redis.io/commands/HSet for the details
// NOTE: Please do not use this function due to complex coupling
func HSet(key string, field string, value interface{}) (bool, error) {
	return client.HSet(key, field, value)
}

// HVals returns all values in the hash stored at key.
// see https://redis.io/commands/HVals for the details
// NOTE: Please do not use this function due to complex coupling
func HVals(key string) ([]string, error) {
	return client.HVals(key)
}

// HIncrBy increments the number stored at field in the hash stored at key by increment.
// If key does not exist, a new key holding a hash is created.
// see https://redis.io/commands/HIncrBy for the details
// NOTE: Please do not use this function due to complex coupling
func HIncrBy(key string, field string, inc int64) (int64, error) {
	return client.HIncrBy(key, field, inc)
}

// HIncrByFloat
// see https://redis.io/commands/HIncrByFloat for the details
// NOTE: Please do not use this function due to complex coupling
func HIncrByFloat(key string, field string, inc float64) (float64, error) {
	return client.HIncrByFloat(key, field, inc)
}
