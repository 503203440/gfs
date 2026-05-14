package utils

import (
	"log"
	"time"
)

func StartCleanupTask() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			log.Printf("cleanup下次执行时间：%s", next)
			time.Sleep(next.Sub(now))

			cutoff := time.Now().Add(-7 * 24 * time.Hour)
			result := DbConnect.Exec("DELETE FROM sys_metric WHERE timestamp < ?", cutoff)
			if result.Error != nil {
				log.Printf("[cleanup] 删除 sys_metric 失败: %v", result.Error)
			} else {
				log.Printf("[cleanup] sys_metric 清理完成，删除了 %d 条记录", result.RowsAffected)
			}
		}
	}()
}
