package models

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

// 自定义时间类型,便于统一序列化
type CustomTime struct {
	time.Time
}

// MarshalJSON 自定义 JSON 序列化方法
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Time.Format("2006-01-02 15:04:05"))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// 去掉 JSON 字符串的引号
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "null" {
			// 如果是 null，设置为零值
			ct.Time = time.Time{}
			return nil
		}
		// 解析时间字符串
		var err error
		ct.Time, err = time.Parse("2006-01-02 15:04:05", str)
		if err != nil {
			return err
		}
	} else {
		var timestamp int64
		if err = json.Unmarshal(data, &timestamp); err == nil {
			timestampLength := len(strconv.FormatInt(timestamp, 10))
			if timestampLength >= 13 { // 如果你看到一个时间戳是10位数，那么很可能是秒级时间戳；如果是13位数，则很可能是毫秒级时间戳。
				// 毫秒
				ct.Time = time.UnixMilli(timestamp)
			} else {
				// 秒
				ct.Time = time.Unix(timestamp, 0)
			}
		} else {
			log.Println("尝试timestamp仍无法解析json数据", err)
			return err
		}
	}

	return nil
}

// 重写toString方法
func (ct CustomTime) String() string {
	return ct.Time.Format("2006-01-02 15:04:05")
}

// Valuer 接口：将 CustomTime 转换为可以存储到数据库的值
func (ct CustomTime) Value() (driver.Value, error) {
	return ct.Time, nil
}

// Scanner 接口：从数据库读取值并转换为 CustomTime
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	// 兼容 SQLite 及不同驱动读取时的多种可能数据类型
	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		return nil

	case int64:
		// 数字时间戳（秒级 10 位 / 毫秒级 13 位）
		timestampLength := len(strconv.FormatInt(v, 10))
		if timestampLength >= 13 {
			ct.Time = time.UnixMilli(v)
		} else {
			ct.Time = time.Unix(v, 0)
		}
		return nil

	case string:
		// SQLite TEXT 字段可能返回标准日期字符串
		if v == "" || v == "null" {
			ct.Time = time.Time{}
			return nil
		}
		layouts := []string{
			"2006-01-02 15:04:05",
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				ct.Time = t
				return nil
			}
		}
		// 若为数字字符串格式时间戳
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			if len(v) >= 13 {
				ct.Time = time.UnixMilli(ts)
			} else {
				ct.Time = time.Unix(ts, 0)
			}
			return nil
		}
		return fmt.Errorf("failed to parse string '%s' to CustomTime", v)

	case []byte:
		// 某些 SQLite 驱动将文本或时间返回为 []byte
		return ct.Scan(string(v))

	default:
		return fmt.Errorf("failed to convert %T to CustomTime", value)
	}
}
