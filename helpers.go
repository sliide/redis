package redis

import (
	"fmt"
	"math"
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

func NumberToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	default:
		return 0.0, false
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		value, err := strconv.ParseFloat(v, 64)
		return value, err == nil
	}
}

func NumberToInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	default:
		return 0, false
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case float32:
		if v > math.MaxInt64 || v < math.MinInt64 {
			return 0, false
		}
		return int64(v), true
	case float64:
		if v > math.MaxInt64 || v < math.MinInt64 {
			return 0, false
		}
		return int64(v), true
	case string:
		value, err := strconv.ParseInt(v, 10, 64)
		return value, err == nil
	}
}

func interfaceSlice(strings []string) []interface{} {
	interfaces := make([]interface{}, 0, len(strings))
	for _, s := range strings {
		interfaces = append(interfaces, s)
	}
	return interfaces
}
