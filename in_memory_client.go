package redis

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func NewInMemoryClient(server string) Client {
	return &InMemoryClient{
		Keys:    map[string]interface{}{},
		Expires: map[string]time.Time{},
	}
}

type InMemoryClient struct {
	Keys    map[string]interface{}
	Expires map[string]time.Time
}

func (dc *InMemoryClient) Close() {}

func (dc *InMemoryClient) Get(key string) (val string, err error) {

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
	dc.Keys[key] = value
	return nil
}

func (dc *InMemoryClient) SetEx(key string, expire int, value interface{}) (err error) {
	dc.Set(key, value)
	return dc.Expire(key, expire)
}

func (dc *InMemoryClient) LPush(key string, value string) (err error) {
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
	if _, ok := dc.Keys[key]; !ok {
		dc.Keys[key] = []string{value}
		return
	}
	array := dc.Keys[key].([]string)
	array = append(array, value)
	dc.Keys[key] = array
	return nil
}

func (dc *InMemoryClient) LRange(key string) (vals []string, err error) {
	if value, ok := dc.Keys[key]; ok {
		return value.([]string), nil
	}
	return []string{}, nil
}

func (dc *InMemoryClient) Pop(key string) (val string, err error) {
	return dc.LPop(key)
}

func (dc *InMemoryClient) LPop(key string) (val string, err error) {
	values, ok := dc.Keys[key]
	if !ok {
		return "", errors.New("Key not found")
	}
	stringArray := values.([]string)

	returnValue := stringArray[0]

	if len(stringArray) == 1 {
		dc.Del(key)
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
	value, ok := dc.Keys[key]

	var incrValue int64

	switch inc.(type) {
	case int:
		incrValue = int64(inc.(int))
	case int32:
		incrValue = int64(inc.(int32))
	default:
		incrValue = inc.(int64)
	}

	if !ok {
		dc.Keys[key] = incrValue
		return incrValue, nil
	}

	var currentValue int64

	switch value.(type) {
	case string:
		if cv, err := strconv.Atoi(value.(string)); err != nil {
			return 0, err
		} else {
			currentValue = int64(cv)
		}
	case int:
		currentValue = int64(value.(int))
	case int32:
		currentValue = int64(value.(int32))
	default:
		currentValue = value.(int64)
	}

	incrValue += currentValue
	dc.Keys[key] = incrValue
	return incrValue, nil
}

func (dc *InMemoryClient) Expire(key string, expire int) (err error) {
	dc.Expires[key] = time.Now().Add(time.Duration(expire) * time.Second)
	return nil
}

func (dc *InMemoryClient) Del(key string) (err error) {
	delete(dc.Keys, key)
	return nil
}

func (dc *InMemoryClient) MGet(keys []string) ([]string, error) {
	values := []string{}
	for _, key := range keys {
		if value, ok := dc.Keys[key]; ok {
			values = append(values, value.(string))
		} else {
			values = append(values, "")
		}
	}
	return values, nil
}

func (dc *InMemoryClient) ZAdd(key string, score float64, value interface{}) (int, error) {
	return 0, nil
}

func (dc *InMemoryClient) ZCount(key string, min interface{}, max interface{}) (int, error) {
	return 0, nil
}
