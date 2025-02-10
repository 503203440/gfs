package models

import (
	"time"

	"gorm.io/gorm"
)

type FileInfo struct {
	gorm.Model
	ID         uint
	Reference  int
	ShaKey     string
	Size       int
	CreateTime time.Time
	UpdateTime time.Time
	URL        string
}

// 在初始化 GORM 时，指定表名
func (FileInfo) TableName() string {
	return "file_info"
}
