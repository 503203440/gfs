package handlers

import (
	"gfs/models"
	"gfs/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TodayFileTotal(c *fiber.Ctx) error {
	var total int64

	utils.DbConnect.AutoMigrate(&models.FileInfo{})
	err := utils.DbConnect.Model(&models.FileInfo{}).
		Where("create_time >= ? AND create_time < ?", time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 1).Format("2006-01-02")).
		Or("update_time >= ? AND update_time < ?", time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 1).Format("2006-01-02")).
		Count(&total).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("查询失败")
	}

	timeStr := time.Now().Format("2006-01-02")
	return c.JSON(fiber.Map{"total": total, "today": timeStr})

}
