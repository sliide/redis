package redis

import (
	"math/rand"
	"strconv"
	"time"
)

// TODO: move it somewhere so we don't copy this everywhere
func RandSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func NumberToFloat64(value interface{}) float64 {
	var returnValue float64

	switch value.(type) {
	case int:
		return float64(value.(int))
	case int32:
		return float64(value.(int32))
	case int64:
		return float64(value.(int64))
	case float32:
		return float64(value.(float32))
	case float64:
		return value.(float64)
	}

	return returnValue
}

func NumberToInt64(value interface{}) int64 {
	var returnValue int64

	switch value.(type) {
	case int:
		return int64(value.(int))
	case int32:
		return int64(value.(int32))
	case int64:
		return int64(value.(int64))
	case float32:
		return int64(value.(float32))
	case float64:
		return int64(value.(float64))
	case string:
		if cv, err := strconv.Atoi(value.(string)); err != nil {
			return 0
		} else {
			return int64(cv)
		}
	}

	return returnValue
}
