package handlers

import (
	"gofs/models"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("./flx.db?_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败")
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库实例失败")
	}
	sqlDB.SetMaxIdleConns(10)           // 设置空闲连接数
	sqlDB.SetMaxOpenConns(30)           // 设置最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期

	db.AutoMigrate(&models.FileInfo{})
}

func TodayFileTotal(c *fiber.Ctx) error {
	var total int64

	err := db.Model(&models.FileInfo{}).
		Where("create_time >= ? AND create_time < ?", time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 1).Format("2006-01-02")).
		Or("update_time >= ? AND update_time < ?", time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 1).Format("2006-01-02")).
		Count(&total).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("查询失败")
	}

	timeStr := time.Now().Format("2006-01-02")
	return c.JSON(fiber.Map{"total": total, "today": timeStr})

}
