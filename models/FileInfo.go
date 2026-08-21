package models

import (
	"time"

	"gorm.io/gorm"
)

type FileInfo struct {
	gorm.Model
	Reference  int
	// 为 (sha_key, size) 建立联合索引，大幅提升文件秒传/去重查询时的索引命中率与查询速度
	ShaKey     string    `gorm:"index:idx_sha_size;size:64"`
	Size       int64     `gorm:"index:idx_sha_size"`
	// 为按时间范围统计今日上传总数建立索引
	CreateTime time.Time `gorm:"index"`
	UpdateTime time.Time
	URL        string
}

// 在初始化 GORM 时，指定表名
func (FileInfo) TableName() string {
	return "file_info"
}
