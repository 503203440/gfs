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
	// 在 Go 语言中，value.(time.Time) 是类型断言（Type Assertion）的语法。
	// 它的作用是尝试将一个接口类型的变量 value 断言为具体的类型 time.Time。
	// 如果断言成功，value 会被转换为 time.Time 类型；如果失败，则会引发错误
	t, ok := value.(time.Time)
	if !ok {
		// return fmt.Errorf("failed to convert %T to CustomTime", value)
		var timestamp int64
		timestamp, ok = value.(int64)
		if !ok {
			return fmt.Errorf("failed to convert %T to CustomTime", value)
		}
		timestampLength := len(strconv.FormatInt(timestamp, 10))
		if timestampLength >= 13 { // 如果你看到一个时间戳是10位数，那么很可能是秒级时间戳；如果是13位数，则很可能是毫秒级时间戳。
			// 毫秒
			ct.Time = time.UnixMilli(timestamp)
		} else {
			// 秒
			ct.Time = time.Unix(timestamp, 0)
		}
	}
	ct.Time = t
	return nil
}
