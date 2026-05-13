package models

import (
	"time"

	"gorm.io/gorm"
)

type SysMetric struct {
	gorm.Model
	ID        uint      `gorm:"primarykey"`
	Type      string    `gorm:"index;size:10"`
	Timestamp time.Time `gorm:"index"`
	Value     string
}

func (SysMetric) TableName() string {
	return "sys_metric"
}
