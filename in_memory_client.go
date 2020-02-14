package redis

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/gomodule/redigo/redis"
)

// NewMemoryClient returns a redis client that running in the memory
// NOTE: Use for the debug purpose
func NewMemoryClient() Client {
	return &memoryClient{
		Keys:    map[string]interface{}{},
		Expires: map[string]time.Time{},
		mu:      sync.Mutex{},
	}
}

type memoryClient struct {
	Keys    map[string]interface{}
	Expires map[string]time.Time
	mu      sync.Mutex
}

func (dc *memoryClient) Close() {}

func (dc *memoryClient) Ping() error {
	return nil
}

func (dc *memoryClient) Get(key string) (val string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		return "", errors.New("redigo: nil returned")
	}

	return valueToString(value), nil
}

func (dc *memoryClient) Set(key string, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Keys[key] = value
	delete(dc.Expires, key)
	return nil
}

func (dc *memoryClient) SetEx(key string, expire int64, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	dc.Keys[key] = value
	return nil
}

func (dc *memoryClient) LPush(key string, value string) (length int64, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values, ok := dc.Keys[key]
	if !ok {
		dc.Keys[key] = []string{value}
		return 1, nil
	}

	array, ok := values.([]string)
	if !ok {
		return 0, errors.New("Can not push into a non list")
	}

	dc.Keys[key] = append([]string{value}, array...)
	delete(dc.Expires, key)
	return int64(len(array) + 1), nil
}

func (dc *memoryClient) RPush(key string, value string) (length int64, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, ok := dc.Keys[key]; !ok {
		dc.Keys[key] = []string{value}
		return 1, nil
	}
	array, ok := dc.Keys[key].([]string)
	if !ok {
		return 0, errors.New("Can not push into a non list")
	}

	dc.Keys[key] = append(array, value)
	delete(dc.Expires, key)
	return int64(len(array) + 1), nil
}

func (dc *memoryClient) LRange(key string) (vals []string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if value, ok := dc.Keys[key]; ok {
		return value.([]string), nil
	}
	return []string{}, nil
}

func (dc *memoryClient) LPop(key string) (val string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		return "", errors.New("Key not found")
	}

	stringArray := values.([]string)
	returnValue := stringArray[0]

	if len(stringArray) == 1 {
		delete(dc.Keys, key)
	} else {
		dc.Keys[key] = stringArray[1:]
	}

	return returnValue, nil
}

func (dc *memoryClient) Incr(key string) (val int64, err error) {
	return dc.IncrBy(key, 1)
}

func (dc *memoryClient) IncrBy(key string, inc int64) (val int64, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if hasExpire && time.Now().After(expire) {
		ok = false
		delete(dc.Expires, key)
	}

	var numericValue int64 = 0
	if ok {
		numericValue, ok = numberToInt64(value)
		if !ok {
			return 0, errors.New("Stored value can not be converted to int64")
		}
	}

	numericValue += inc
	dc.Keys[key] = numericValue
	return numericValue, nil
}

func (dc *memoryClient) IncrByFloat(key string, inc float64) (float64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if hasExpire && time.Now().After(expire) {
		ok = false
		delete(dc.Expires, key)
	}

	var numericValue float64 = 0
	if ok {
		numericValue, ok = numberToFloat64(value)
		if !ok {
			return 0, errors.New("Stored value can not be converted to float64")
		}
	}

	numericValue += inc
	dc.Keys[key] = numericValue
	return numericValue, nil
}

func (dc *memoryClient) Expire(key string, expire int64) (bool, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, exists := dc.Keys[key]; !exists {
		return false, nil
	}

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	return true, nil
}

func (dc *memoryClient) Del(keys ...string) (count int64, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	count = 0
	for _, key := range keys {
		if _, exists := dc.Keys[key]; exists {
			delete(dc.Keys, key)
			count++
		}
	}
	return count, nil
}

func (dc *memoryClient) MGet(keys ...string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values := []string{}

	for _, key := range keys {
		value, ok := dc.Keys[key]
		expire, hasExpire := dc.Expires[key]

		if !ok || (hasExpire && time.Now().After(expire)) {
			values = append(values, "")
		} else {
			values = append(values, valueToString(value))
		}
	}

	return values, nil
}

func (dc *memoryClient) ZAdd(key string, score float64, value interface{}) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	scoreAndValue := []interface{}{score, value}

	value, ok := dc.Keys[key].([][]interface{})
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		dc.Keys[key] = [][]interface{}{
			scoreAndValue,
		}
		return 1, nil
	}

	currentSet, ok := value.([][]interface{})
	if !ok {
		return 0, errors.New("Couldn't convert to type")
	}

	// don't append if value exists on set.
	for i := 0; i < len(currentSet); i++ {
		if scoreAndValue[0] == currentSet[i][0] && scoreAndValue[1] == currentSet[i][1] {
			return 0, nil
		}
	}

	// bubble sort
	for i := 0; i < len(currentSet); i++ {
		for j := i; j < len(currentSet); j++ {
			if currentSet[i][0].(float64) < currentSet[j][0].(float64) {
				currentSet[i], currentSet[j] = currentSet[j], currentSet[i]
			}
		}
	}

	dc.Keys[key] = append(currentSet, scoreAndValue)
	return 1, nil
}

func (dc *memoryClient) ZCount(key string, min interface{}, max interface{}) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		return 0, nil
	}

	currentSet, ok := value.([][]interface{})
	var negativeInfinite bool
	var positiveInfinite bool

	var minScore float64
	var maxScore float64

	switch min.(type) {
	case string:
		negativeInfinite = true
	default:
		minScore, ok = numberToFloat64(min)
		if !ok {
			return 0, errors.New("minimum score is not a number")
		}
	}

	switch max.(type) {
	case string:
		positiveInfinite = true
	default:
		maxScore, ok = numberToFloat64(max)
		if !ok {
			return 0, errors.New("maximum score is not a number")
		}
	}

	if negativeInfinite && positiveInfinite {
		return int64(len(currentSet)), nil
	}

	count := int64(0)
	for i := 0; i < len(currentSet); i++ {
		currentScore := currentSet[i][0].(float64)

		if (negativeInfinite && currentScore < maxScore) ||
			(currentScore >= minScore && currentScore < maxScore) ||
			(currentScore >= minScore && positiveInfinite) {

			count = count + 1
		}
	}

	return count, nil
}

type set map[string]interface{}

func (s *set) Add(members ...string) {
	for _, member := range members {
		(*s)[member] = nil
	}
}

func (s set) Members() []string {
	retVal := make([]string, 0, len(s))
	for m := range s {
		retVal = append(retVal, m)
	}
	return retVal
}

func (dc *memoryClient) SAdd(key string, members ...string) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		newSet := make(set, len(members))
		newSet.Add(members...)
		dc.Keys[key] = newSet
		return int64(len(newSet)), nil
	}

	existingSet, ok := value.(set)
	if !ok {
		return 0, errors.New("Stored value is not a set")
	}
	oldSize := len(existingSet)
	existingSet.Add(members...)

	return int64(len(existingSet) - oldSize), nil
}

func (dc *memoryClient) SMembers(key string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		return []string{}, nil
	}
	existingSet, ok := value.(set)
	if !ok {
		return nil, errors.New("Stored value is not a set")
	}

	return existingSet.Members(), nil
}

func (dc *memoryClient) SetNxEx(key string, value interface{}, timeout int64) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	_, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if (ok && (hasExpire && !time.Now().After(expire))) || ok && !hasExpire {
		return 0, nil
	}

	dc.Expires[key] = time.Now().Add(time.Duration(timeout) * time.Second)
	dc.Keys[key] = value

	return 1, nil
}

func (dc *memoryClient) Eval(script string, keyCount int) (interface{}, error) {
	// not implemented
	return nil, nil
}

func (dc *memoryClient) getHash(key string) (map[string]string, bool) {
	value := dc.Keys[key]
	if value == nil {
		return nil, true
	}

	hash, ok := value.(map[string]string)
	return hash, ok
}

func (dc *memoryClient) getHashAndCreateIfNotExists(key string) (map[string]string, bool) {
	value := dc.Keys[key]
	if value == nil {
		value = make(map[string]string)
		dc.Keys[key] = value
	}

	hash, ok := value.(map[string]string)
	return hash, ok
}

func (dc *memoryClient) HDel(key string, fields ...string) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return 0, errors.New("try to delete fields from non hash value")
	}

	if hash == nil {
		return 0, nil
	}

	deleteCounts := int64(0)
	for _, field := range fields {
		_, ok = hash[field]
		if ok {
			deleteCounts++
			delete(hash, field)
		}
	}

	return deleteCounts, nil
}

func (dc *memoryClient) HExists(key string, field string) (bool, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return false, errors.New("try to delete fields from non hash value")
	}

	if hash == nil {
		return false, nil
	}

	_, ok = hash[field]
	return ok, nil
}

func (dc *memoryClient) HGet(key string, field string) (string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return "", errors.New("try to get field from non hash value")
	}

	if hash == nil {
		return "", redis.ErrNil
	}

	fieldValue, ok := hash[field]
	if !ok {
		return "", redis.ErrNil
	}

	return fieldValue, nil
}

func (dc *memoryClient) HGetAll(key string) (map[string]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return nil, errors.New("try to get all fields from non hash value")
	}

	return hash, nil
}

func (dc *memoryClient) HLen(key string) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return 0, errors.New("try to get length of non hash value")
	}

	return int64(len(hash)), nil
}

func (dc *memoryClient) HMGet(key string, fields ...string) (map[string]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return nil, errors.New("try to get fields from non hash value")
	}

	if hash == nil {
		return nil, nil
	}

	filteredHash := make(map[string]string, len(fields))
	for _, field := range fields {
		v, ok := hash[field]
		if ok {
			filteredHash[field] = v
		}
	}

	return filteredHash, nil
}

func (dc *memoryClient) HKeys(key string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return nil, errors.New("try to get keys from non hash value")
	}

	if hash == nil {
		return nil, nil
	}

	keys := make([]string, 0, len(hash))
	for k := range hash {
		keys = append(keys, k)
	}

	return keys, nil
}

func (dc *memoryClient) HMSet(key string, fields map[string]interface{}) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHashAndCreateIfNotExists(key)
	if !ok {
		return errors.New("try to set keys on non hash value")
	}

	for field, value := range fields {
		hash[field] = valueToString(value)
	}

	return nil
}

func (dc *memoryClient) HSet(key string, field string, value interface{}) (bool, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHashAndCreateIfNotExists(key)
	if !ok {
		return false, errors.New("try to set keys on non hash value")
	}

	_, exists := hash[field]
	hash[field] = valueToString(value)

	return !exists, nil
}

func (dc *memoryClient) HVals(key string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return nil, errors.New("try to get values from non hash value")
	}

	if hash == nil {
		return nil, nil
	}

	values := make([]string, 0, len(hash))
	for _, v := range hash {
		values = append(values, v)
	}

	return values, nil
}

func (dc *memoryClient) HScan(key string, pattern string) (map[string]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHash(key)
	if !ok {
		return nil, errors.New("try to get values from non hash value")
	}

	if hash == nil {
		return nil, nil
	}

	matcher, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}
	matchedHash := make(map[string]string, 0)
	for k, v := range hash {
		if matcher.Match(k) {
			matchedHash[k] = v
		}
	}

	return matchedHash, nil
}

func (dc *memoryClient) HIncrBy(key string, field string, inc int64) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHashAndCreateIfNotExists(key)
	if !ok {
		return 0, errors.New("try to increase field value on non hash value")
	}

	value := hash[field]
	if len(value) == 0 {
		value = "0"
	}

	number, ok := numberToInt64(value)
	if !ok {
		return 0, errors.New("value to be increased can not be converted to integer")
	}

	number += inc

	hash[field] = valueToString(number)
	return number, nil
}

func (dc *memoryClient) HIncrByFloat(key string, field string, inc float64) (float64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	hash, ok := dc.getHashAndCreateIfNotExists(key)
	if !ok {
		return 0, errors.New("try to increase field value on non hash value")
	}

	value := hash[field]
	if len(value) == 0 {
		value = "0.0"
	}

	number, ok := numberToFloat64(value)
	if !ok {
		return 0, errors.New("value to be increased can not be converted to float")
	}

	number += inc

	hash[field] = fmt.Sprintf("%f", number)
	return number, nil
}
