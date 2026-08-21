package handlers

import (
	"gfs/models"
	"gfs/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TodayFileTotal(c *fiber.Ctx) error {
	var total int64

	// 计算今日的起始时间与明日零点 [todayStart, todayEnd)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	// 优化说明：
	// 1. 使用标准 time.Time 类型传递时间，由 SQLite/GORM 正确序列化与比较
	// 2. 显式使用括号包含 OR 条件，避免 GORM 软删除 deleted_at 条件与 OR 发生运算符优先级歧义
	err := utils.DbConnect.Model(&models.FileInfo{}).
		Where("(create_time >= ? AND create_time < ?) OR (update_time >= ? AND update_time < ?)", todayStart, todayEnd, todayStart, todayEnd).
		Count(&total).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("查询失败")
	}

	timeStr := todayStart.Format("2006-01-02")
	return c.JSON(fiber.Map{"total": total, "today": timeStr})

}
