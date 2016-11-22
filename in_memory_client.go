package redis

import (
	"errors"
	"sync"
	"time"
)

func NewInMemoryClient() Client {
	return &InMemoryClient{
		Keys:    map[string]interface{}{},
		Expires: map[string]time.Time{},
		mu:      sync.Mutex{},
	}
}

type InMemoryClient struct {
	Keys    map[string]interface{}
	Expires map[string]time.Time
	mu      sync.Mutex
}

func (dc *InMemoryClient) Close() {}

func (dc *InMemoryClient) Get(key string) (val string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]
	expire, hasExpire := dc.Expires[key]

	if !ok || (hasExpire && time.Now().After(expire)) {
		return "", errors.New("redigo: nil returned")
	}

	return ValueToString(value), nil
}

func (dc *InMemoryClient) Set(key string, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Keys[key] = value
	delete(dc.Expires, key)
	return nil
}

func (dc *InMemoryClient) SetEx(key string, expire int64, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	dc.Keys[key] = value
	return nil
}

func (dc *InMemoryClient) LPush(key string, value string) (length int64, err error) {
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

func (dc *InMemoryClient) RPush(key string, value string) (length int64, err error) {
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

func (dc *InMemoryClient) LRange(key string) (vals []string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if value, ok := dc.Keys[key]; ok {
		return value.([]string), nil
	}
	return []string{}, nil
}

func (dc *InMemoryClient) LPop(key string) (val string, err error) {
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

func (dc *InMemoryClient) Incr(key string) (val int64, err error) {
	return dc.IncrBy(key, 1)
}

func (dc *InMemoryClient) IncrBy(key string, inc int64) (val int64, err error) {
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
		numericValue, ok = NumberToInt64(value)
		if !ok {
			return 0, errors.New("Stored value can not be converted to int64")
		}
	}

	numericValue += inc
	dc.Keys[key] = numericValue
	return numericValue, nil
}

func (dc *InMemoryClient) Expire(key string, expire int64) (bool, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, exists := dc.Keys[key]; !exists {
		return false, nil
	}

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	return true, nil
}

func (dc *InMemoryClient) Del(keys ...string) (count int64, err error) {
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

func (dc *InMemoryClient) MGet(keys ...string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values := []string{}

	for _, key := range keys {
		value, ok := dc.Keys[key]
		expire, hasExpire := dc.Expires[key]

		if !ok || (hasExpire && time.Now().After(expire)) {
			values = append(values, "")
		} else {
			values = append(values, ValueToString(value))
		}
	}

	return values, nil
}

func (dc *InMemoryClient) ZAdd(key string, score float64, value interface{}) (int64, error) {
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

func (dc *InMemoryClient) ZCount(key string, min interface{}, max interface{}) (int64, error) {
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
		minScore = NumberToFloat64(min)
	}

	switch max.(type) {
	case string:
		positiveInfinite = true
	default:
		maxScore = NumberToFloat64(max)
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

			count += 1
		}
	}

	return count, nil
}

func (dc *InMemoryClient) SetNxEx(key string, value interface{}, timeout int64) (int64, error) {
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
