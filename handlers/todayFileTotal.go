package handlers

import (
	"gofs/models"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TodayFileTotal(c *fiber.Ctx) error {
	var total int64

	// 打开数据库
	db, err := gorm.Open(sqlite.Open("./flx.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("数据库连接失败")
	}
	// 迁移 schema
	db.AutoMigrate(&models.FileInfo{})

	db.Debug().Count(&total)

	timeStr := time.Now().Format("2006-01-02")
	return c.JSON(fiber.Map{"total": total, "today": timeStr})

}
