package xkit

import (
	"encoding/json"
	"fmt"
	"time"
)

func MustToTime(value any, formats ...string) time.Time {
	t, err := ToTime(value, formats...)
	if err != nil {
		panic(err)
	}
	return t
}

// ToTime 灵活时间解析，支持 time.Time、字符串、int/int64 时间戳（秒/毫秒）
func ToTime(value any, formats ...string) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case int:
		return parseIntTime(int64(v)), nil
	case int64:
		return parseIntTime(v), nil
	case string:
		return parseStringTime(v, formats...)
	default:
		if i64, ok := toInt64(value); ok && i64 != 0 {
			return parseIntTime(i64), nil
		}
	}
	return time.Time{}, fmt.Errorf("%v unable to parse as time.Time", value)
}

// toInt64 把整数、浮点与 json.Number 转为 int64，替代 cast 以保持根包零第三方依赖。
func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
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
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		i, err := v.Int64()
		return i, err == nil
	}
	return 0, false
}

// parseIntTime 自动识别秒/毫秒时间戳
func parseIntTime(ts int64) time.Time {
	// 毫秒时间戳大于 13 位数字（year 2286+ 的秒时间戳）
	if ts > 1e12 {
		return time.UnixMilli(ts)
	}
	return time.Unix(ts, 0)
}

// parseStringTime 自动识别时间字符串
func parseStringTime(value string, formats ...string) (time.Time, error) {
	l := time.Now().Location()
	var layouts = []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var t time.Time
	var err error
	for _, layout := range append(layouts, formats...) {
		if t, err = time.ParseInLocation(layout, value, l); err == nil {
			return t, nil
		}
	}
	return time.Time{}, err
}

// FormatMilli 将毫秒时间戳格式化为字符串
func FormatMilli(value int64, layout string) string {
	return time.UnixMilli(value).Format(layout)
}

// GetPrevDay 获取前 N 天的时间
func GetPrevDay(prev time.Duration) time.Time {
	return time.Now().Add(-prev)
}
