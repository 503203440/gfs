package models

import (
	"log"
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
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	log.Println("str:", str)
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
	return nil
}
