package utils

import (
	"gfs/appinit"
	"gfs/models"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DbConnect *gorm.DB

func init() {
	var err error
	DbConnect, err = gorm.Open(sqlite.Open("./flx.db?_journal_mode=WAL"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
		Logger: logger.New(
			log.New(
				models.NewMultiWrite(appinit.AppLogWrite, os.Stdout),
				"\r\n",
				log.LstdFlags,
			),
			logger.Config{LogLevel: logger.Info},
		),
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
}
