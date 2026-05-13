package utils

import (
	"context"
	"gfs/appinit"
	"gfs/models"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DbConnect *gorm.DB

// 静音 sys_metric 表的 SQL 日志
type silentMetricLogger struct {
	logger.Interface
}

func (l *silentMetricLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	if strings.Contains(sql, "sys_metric") {
		return
	}
	l.Interface.Trace(ctx, begin, func() (string, int64) { return sql, rows }, err)
}

func init() {
	var err error
	DbConnect, err = gorm.Open(sqlite.Open("./flx.db?_journal_mode=WAL"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
		Logger: &silentMetricLogger{
			logger.New(
				log.New(
					models.NewMultiWrite(appinit.AppLogWrite, os.Stdout),
					"\r\n",
					log.LstdFlags,
				),
				logger.Config{LogLevel: logger.Info},
			),
		},
	})
	if err != nil {
		log.Fatal("数据库连接失败")
	}

	// 配置连接池
	sqlDB, err := DbConnect.DB()
	if err != nil {
		log.Fatal("获取数据库实例失败")
	}
	sqlDB.SetMaxIdleConns(10)           // 设置空闲连接数
	sqlDB.SetMaxOpenConns(30)           // 设置最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期

	// 自动绑定模型
	DbConnect.AutoMigrate(&models.FileInfo{})
	DbConnect.AutoMigrate(&models.ClientInfoEntity{})
	DbConnect.AutoMigrate(&models.TokenEntity{})
	DbConnect.AutoMigrate(&models.SysMetric{})
}
