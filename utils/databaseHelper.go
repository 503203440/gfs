package utils

import (
	"context"
	"gfs/appinit"
	"gfs/models"
	"log"
	"os"
	"path/filepath"
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
	// 使用可执行文件所在目录绝对路径，避免在不同工作目录下运行导致数据库路径分裂或权限问题
	dbPath := filepath.Join(appinit.BaseDir, "flx.db")

	// DSN 参数优化说明：
	// 1. _pragma=journal_mode(WAL): 开启预写日志模式（WAL），大幅提升并发读写性能并允许读写并发
	// 2. _pragma=busy_timeout(5000): 设置 5000ms 锁忙等待超时，避免并发写入竞争时直接报错 database is locked
	// 3. _pragma=synchronous(NORMAL): WAL 模式下结合 NORMAL 模式兼顾性能与事务持久性
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"

	DbConnect, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
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
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 配置连接池
	// SQLite 单文件数据库写操作为串行独占锁，过大的连接池不仅无益还会增加锁竞争，适当限制最大打开连接数
	sqlDB, err := DbConnect.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(5)            // 设置空闲连接数
	sqlDB.SetMaxOpenConns(10)           // 设置最大连接数（控制并发连接，配合 busy_timeout 避免锁争用）
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期

	// 自动绑定并迁移模型，捕获错误
	if err := DbConnect.AutoMigrate(
		&models.FileInfo{},
		&models.ClientInfoEntity{},
		&models.TokenEntity{},
		&models.SysMetric{},
	); err != nil {
		log.Fatalf("数据库模型自动迁移失败: %v", err)
	}
}
