package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gofs/models"
	"log"
)

// GetUsers
// 获取用户集合
func GetUsers(c *fiber.Ctx) error {
	users := []models.User{
		{"aaa", 1},
		{"d4", 1},
		{"ds", 1},
		{"frd", 1},
	}
	return c.JSON(users)
}

func CreateUser(c *fiber.Ctx) error {
	var user models.User

	if err := c.BodyParser(&user); err != nil {
		log.Printf("解析body失败,无法转化为user实例, err:%s\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 9999, "msg": "解析参数失败", "data": string(c.Body())})
	}
	return c.JSON(user)
}
