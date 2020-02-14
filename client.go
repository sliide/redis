package redis

// Client defines functions that access memory storeage (e.g. Redis)
type Client interface {
	// Close the client connection
	Close()

	// Ping the service to check connection
	Ping() error

	// Get return a value from the storage by a given key
	// see https://redis.io/commands/Get for the details
	Get(key string) (string, error)
	// MGet returns values from the storage by given keys
	// see https://redis.io/commands/MGet for the details
	MGet(keys ...string) ([]string, error)
	// Set key to hold the value in the storage
	// see https://redis.io/commands/Set for the details
	Set(key string, value interface{}) error
	// SetEx which set key to hold the value and
	// set key to timeout after a given number of seconds in the storage
	// see https://redis.io/commands/SetEx for the details
	SetEx(key string, expire int64, value interface{}) error
	// Set key to hold string value if key does not exist.
	// In that case, it is equal to SET. When key already holds a value, no operation is performed.
	// SETNX is short for "SET if Not eXists".
	// see https://redis.io/commands/SetNX for the details
	SetNxEx(key string, value interface{}, expire int64) (int64, error)
	// Expire sets key to timeout after a given number of seconds
	// see https://redis.io/commands/Expire for the details
	Expire(key string, seconds int64) (bool, error)
	// Del deleted the key in the storage
	// see https://redis.io/commands/Del for the details
	Del(keys ...string) (int64, error)
	// Incr the number stored at key by one.
	// If the key does not exist, it is set to 0 before performing the operation.
	// see https://redis.io/commands/Incr for the details
	Incr(key string) (int64, error)
	// IncrBy the number stored at key by a given value `inc`.
	// If the key does not exist, it is set to 0 before performing the operation.
	IncrBy(key string, inc int64) (int64, error)
	// IncrByFloat the number stored at key by a given value `inc`.
	// If the key does not exist, it is set to 0 before performing the operation.
	IncrByFloat(key string, inc float64) (float64, error)

	// Eval evaluates scripts using the Lua interpreter built into Redis
	// see https://redis.io/commands/Eval for the details
	Eval(string, int) (interface{}, error)

	// LPush which insert the given value at the head of the list stored at key.
	// If key does not exist, it is created as empty list before performing the push operations.
	// see https://redis.io/commands/LPush for the details
	LPush(key string, value string) (int64, error)
	// RPush which insert the given value at the trail of the list stored at key.
	// If key does not exist, it is created as empty list before performing the push operations.
	// see https://redis.io/commands/RPush for the details
	RPush(key string, value string) (int64, error)
	// LRange returns the specified elements of the list stored at key.
	// see https://redis.io/commands/LRange for the details
	LRange(key string) ([]string, error)
	// LPop removes and returns the first element of the list stored at key.
	// see https://redis.io/commands/LPop for the details
	LPop(key string) (string, error)

	// ZAdd
	// see https://redis.io/commands/ZAdd for the details
	ZAdd(key string, score float64, value interface{}) (int64, error)
	// ZCount
	// see https://redis.io/commands/ZCount for the details
	ZCount(key string, min interface{}, max interface{}) (int64, error)

	// SAdd
	// see https://redis.io/commands/SAdd for the details
	SAdd(key string, members ...string) (int64, error)
	// SMembers
	// see https://redis.io/commands/SMembers for the details
	SMembers(key string) ([]string, error)

	// HDel removes the given fields from the hash stored at key.
	// see https://redis.io/commands/HDel for the details
	HDel(key string, fields ...string) (int64, error)
	// HExists returns if field is an existing field in the hash stored at key.
	// see https://redis.io/commands/HExists for the details
	HExists(key string, field string) (bool, error)
	// HGet returns the value associated with field in the hash stored at key.
	// see https://redis.io/commands/HGet for the details
	HGet(key string, field string) (string, error)
	// HGetAll returns all fields and values of the hash stored at key.
	// see https://redis.io/commands/HGetAll for the details
	HGetAll(key string) (map[string]string, error)
	// HMGet returns the values associated with the given fields in the hash stored at key.
	// see https://redis.io/commands/HMGet for the details
	HMGet(key string, fields ...string) (map[string]string, error)
	// HLen returns the number of fields contained in the hash stored at key.
	// see https://redis.io/commands/HLen for the details
	HLen(key string) (int64, error)
	// HKeys returns all field names in the hash stored at key.
	// see https://redis.io/commands/HKeys for the details
	HKeys(key string) ([]string, error)
	// HVals returns all values in the hash stored at key.
	// see https://redis.io/commands/HVals for the details
	HVals(key string) ([]string, error)
	// HScan iterates fields of Hash types and their associated values and returns the matching values.
	// see https://redis.io/commands/HScan for the details
	HScan(key string, pattern string) (map[string]string, error)
	// HSet sets field in the hash stored at key to value.
	// If key does not exist, a new key holding a hash is created.
	// see https://redis.io/commands/HSet for the details
	HSet(key string, field string, value interface{}) (bool, error)
	// HMSet sets the specified fields to their respective values in the hash stored at key.
	// see https://redis.io/commands/HMSet for the details
	HMSet(key string, fields map[string]interface{}) error
	// HIncrBy increments the number stored at field in the hash stored at key by increment.
	// If key does not exist, a new key holding a hash is created.
	// see https://redis.io/commands/HIncrBy for the details
	HIncrBy(key string, field string, inc int64) (int64, error)
	// HIncrBy increments the number stored at field in the hash stored at key by increment.
	// If key does not exist, a new key holding a hash is created.
	// see https://redis.io/commands/HIncrBy for the details
	HIncrByFloat(key string, field string, inc float64) (float64, error)
}
