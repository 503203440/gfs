package models

import (
	"testing"
	"time"
)

func TestCustomTime_Scan(t *testing.T) {
	var ct CustomTime

	// 1. 测试 time.Time 类型
	now := time.Now().Truncate(time.Second)
	if err := ct.Scan(now); err != nil {
		t.Fatalf("Scan(time.Time) 失败: %v", err)
	}
	if !ct.Time.Equal(now) {
		t.Errorf("Scan(time.Time) 期望 %v, 得到 %v", now, ct.Time)
	}

	// 2. 测试 int64 秒级时间戳
	sec := int64(1700000000)
	expectedSecTime := time.Unix(sec, 0)
	if err := ct.Scan(sec); err != nil {
		t.Fatalf("Scan(int64 sec) 失败: %v", err)
	}
	if !ct.Time.Equal(expectedSecTime) {
		t.Errorf("Scan(int64 sec) 期望 %v, 得到 %v", expectedSecTime, ct.Time)
	}

	// 3. 测试 int64 毫秒级时间戳
	milli := int64(1700000000000)
	expectedMilliTime := time.UnixMilli(milli)
	if err := ct.Scan(milli); err != nil {
		t.Fatalf("Scan(int64 milli) 失败: %v", err)
	}
	if !ct.Time.Equal(expectedMilliTime) {
		t.Errorf("Scan(int64 milli) 期望 %v, 得到 %v", expectedMilliTime, ct.Time)
	}

	// 4. 测试 string 格式日期
	dateStr := "2026-08-21 15:30:00"
	if err := ct.Scan(dateStr); err != nil {
		t.Fatalf("Scan(string) 失败: %v", err)
	}
	if ct.String() != dateStr {
		t.Errorf("Scan(string) 期望 %s, 得到 %s", dateStr, ct.String())
	}

	// 5. 测试 []byte 格式日期
	if err := ct.Scan([]byte(dateStr)); err != nil {
		t.Fatalf("Scan([]byte) 失败: %v", err)
	}
	if ct.String() != dateStr {
		t.Errorf("Scan([]byte) 期望 %s, 得到 %s", dateStr, ct.String())
	}

	// 6. 测试 nil
	if err := ct.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) 失败: %v", err)
	}
	if !ct.Time.IsZero() {
		t.Errorf("Scan(nil) 期望零值, 得到 %v", ct.Time)
	}
}
