package redis

import (
	"fmt"
	"strconv"
)

func ValueToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case int:
		return fmt.Sprintf("%d", value.(int))
	case int32:
		return fmt.Sprintf("%d", value.(int32))
	case int64:
		return fmt.Sprintf("%d", value.(int64))
	case float32:
		return fmt.Sprintf("%f", value.(float32))
	case float64:
		return fmt.Sprintf("%f", value.(float64))
	}
	return ""
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

func interfaceSlice(strings []string) []interface{} {
	interfaces := make([]interface{}, 0, len(strings))
	for _, s := range strings {
		interfaces = append(interfaces, s)
	}
	return interfaces
}
