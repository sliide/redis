package redis

import (
	"errors"
	"fmt"
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

	if !ok {
		return "", nil
	}

	expire, ok := dc.Expires[key]

	if !ok {
		switch value.(type) {
		case string:
			return value.(string), nil
		case int:
			return fmt.Sprintf("%d", value.(int)), nil
		case int64:
			return fmt.Sprintf("%d", value.(int64)), nil
		}
	}

	if time.Now().After(expire) {
		return "", nil
	}

	return value.(string), nil
}

func (dc *InMemoryClient) Set(key string, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Keys[key] = value
	return nil
}

func (dc *InMemoryClient) SetEx(key string, expire int, value interface{}) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	dc.Keys[key] = value
	return nil
}

func (dc *InMemoryClient) LPush(key string, value string) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values, ok := dc.Keys[key]
	if !ok {
		dc.Keys[key] = []string{value}
		return nil
	}

	array := values.([]string)
	dc.Keys[key] = append([]string{value}, array...)
	return nil
}

func (dc *InMemoryClient) RPush(key string, value string) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, ok := dc.Keys[key]; !ok {
		dc.Keys[key] = []string{value}
		return
	}
	array := dc.Keys[key].([]string)

	dc.Keys[key] = append(array, value)
	return nil
}

func (dc *InMemoryClient) LRange(key string) (vals []string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if value, ok := dc.Keys[key]; ok {
		return value.([]string), nil
	}
	return []string{}, nil
}

func (dc *InMemoryClient) Pop(key string) (val string, err error) {
	return dc.LPop(key)
}

func (dc *InMemoryClient) LPop(key string) (val string, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values, ok := dc.Keys[key]

	if !ok {
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

func (dc *InMemoryClient) Incr(key string) (err error) {
	dc.IncrBy(key, 1)
	return nil
}

func (dc *InMemoryClient) IncrBy(key string, inc interface{}) (val interface{}, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]

	var incrValue = NumberToInt64(inc)

	if !ok {
		dc.Keys[key] = incrValue
		return incrValue, nil
	}

	var currentValue = NumberToInt64(value)

	incrValue += currentValue
	dc.Keys[key] = incrValue
	return incrValue, nil
}

func (dc *InMemoryClient) Expire(key string, expire int) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	return nil
}

func (dc *InMemoryClient) Del(key string) (err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	delete(dc.Keys, key)
	return nil
}

func (dc *InMemoryClient) MGet(keys []string) ([]string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	values := []string{}

	for _, key := range keys {
		if value, ok := dc.Keys[key]; ok {
			values = append(values, ValueToString(value))
		} else {
			values = append(values, "")
		}
	}

	return values, nil
}

func (dc *InMemoryClient) ZAdd(key string, score float64, value interface{}) (int, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	scoreAndValue := []interface{}{score, value}

	value, ok := dc.Keys[key].([][]interface{})

	if !ok {
		dc.Keys[key] = [][]interface{}{
			scoreAndValue,
		}
		return 1, nil
	}

	currentSet, ok := value.([][]interface{})
	if !ok {
		return 0, errors.New("Couldn't convert to type")
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

func (dc *InMemoryClient) ZCount(key string, min interface{}, max interface{}) (int, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	value, ok := dc.Keys[key]

	if !ok {
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
		return len(currentSet), nil
	}

	count := 0
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
