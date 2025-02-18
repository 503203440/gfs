package handlers

import (
	"gfs/models"
	"gfs/utils"
	"log"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func GetSign(c *fiber.Ctx) error {
	var signVo models.SignVo
	if err := c.BodyParser(&signVo); err != nil {
		log.Printf("body参数解析失败:%s", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ApiErrorDetail("参数解析失败", "999", string(c.Body())))
	} else {
		jsonData, _ := json.Marshal(signVo)
		log.Println("signVO:", string(jsonData))

		if signVo.ClientId == nil {
			// br := models.BaseResponse{}
			return c.JSON(models.ApiError("clientId不能为空"))
		}
		if signVo.Sign == "" {
			return c.JSON(models.ApiError("签名不能为空"))
		}
		var clientInfoEntity models.ClientInfoEntity

		// 查询数据库
		utils.DbConnect.AutoMigrate(&models.ClientInfoEntity{})
		result := utils.DbConnect.Model(&models.ClientInfoEntity{}).Where("id = ?", &signVo.ClientId).Take(&clientInfoEntity)

		if result.Error != nil {
			// if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 	return c.JSON(models.ApiError("未找到客户端信息"))
			// } else {
			return c.JSON(models.ApiErrorDetail("查询数据错误", "999", result.Error.Error()))
			// }
		}
		log.Println("clientInfo:", clientInfoEntity.Id, clientInfoEntity.SecertKey)

		// 将body参数转化为map
		// signData := utils.StructToMap(signVo)
		// 由于原来的版本就是没有任何内容参与计算,所以这里直接给一个空的map
		signData := make(map[string]any)
		// 计算签名
		checkResult := utils.GetSign(signData, clientInfoEntity.SecertKey, utils.HMACSHA256)
		if checkResult != signVo.Sign {
			return c.JSON(models.ApiError("签名验证不通过"))
		}

		return c.JSON(signVo)
	}

}
