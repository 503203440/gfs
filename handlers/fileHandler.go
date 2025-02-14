package handlers

import (
	"gfs/models"
	"log"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func GetSign(c *fiber.Ctx) error {
	var signVo models.SignVo
	if err := c.BodyParser(&signVo); err != nil {
		log.Printf("body参数解析失败:%s", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code": 999,
			"msg":  "body参数解析失败",
			"data": string(c.Body()),
		})
	} else {
		jsonData, _ := json.Marshal(signVo)
		log.Println("signVO:", string(jsonData))
		return c.JSON(signVo)
	}

}
