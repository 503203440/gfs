package handlers

import (
	"errors"
	"fmt"
	"gfs/appinit"
	"gfs/models"
	"gfs/utils"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var uploadPath string

func init() {
	// 创建文件上传目录,如果不存在
	uploadPath = path.Join(appinit.BaseDir, "file-uploads")

	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		err := os.MkdirAll(uploadPath, 0755)
		if err != nil {
			log.Println("创建file-uploads文件夹失败!", err)
		} else {
			log.Println("创建file-uploads成功!")
		}
	}

}

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
		result := utils.DbConnect.Model(&models.ClientInfoEntity{}).Where("id = ?", &signVo.ClientId).Take(&clientInfoEntity)

		if result.Error != nil {
			// if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 	return c.JSON(models.ApiError("未找到客户端信息"))
			// } else {
			return c.JSON(models.ApiErrorDetail("查询数据错误", "999", result.Error.Error()))
			// }
		}
		clientId := clientInfoEntity.Id
		log.Println("clientInfo:", *clientInfoEntity.Id, clientInfoEntity.SecertKey)

		// 将body参数转化为map
		// signData := utils.StructToMap(signVo)
		// 由于原来的版本就是没有任何内容参与计算,所以这里直接给一个空的map
		signData := make(map[string]any)
		// 计算签名
		calculateSignStr := utils.GetSign(signData, clientInfoEntity.SecertKey, utils.HMACSHA256)
		log.Println("计算签名结果", calculateSignStr)
		if calculateSignStr != signVo.Sign {
			return c.JSON(models.ApiError("签名验证不通过"))
		}

		tokenEntity := models.TokenEntity{
			Token:       strings.ToUpper(utils.GenerateRandomString(18)),
			ClientID:    *clientId,
			Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
			ExpiresTime: signVo.ExpiresTime,
			Used:        false,
		}
		utils.DbConnect.Save(&tokenEntity)

		return c.JSON(tokenEntity)
	}

}

// 返回结果中不包含名称
func UploadNoName(c *fiber.Ctx) error {
	token := c.FormValue("token")
	if token == "" {
		return c.JSON(models.ApiError("token不能为空"))
	}
	var tokenEntity models.TokenEntity
	queryResult := utils.DbConnect.Model(&models.TokenEntity{}).Where("token = ?", token).Take(&tokenEntity)
	if queryResult.Error != nil {
		if errors.Is(queryResult.Error, gorm.ErrRecordNotFound) {
			return c.JSON(models.ApiError("token无效 token:" + token))
		} else {
			log.Printf("查询发生错误!%v", queryResult.Error)
			return c.JSON(models.ApiError("服务器错误:" + queryResult.Error.Error()))
		}
	} else {
		if tokenEntity.Used {
			return c.JSON(models.ApiError("token已使用"))
		}
		urls := make([]string, 0)
		if multipartForm, err := c.MultipartForm(); err != nil {
			log.Println("表单读取错误:", err)
			return c.JSON(models.ApiError("读取文件错误" + err.Error()))
		} else {
			files := multipartForm.File["file"]

			for _, f := range files {
				// 将文件保存到本地目录
				log.Println("文件名:", f.Filename)
				extName := strings.ToLower(path.Ext(f.Filename))
				randFileName := utils.GenerateRandomString(20) + extName
				saveFilePath := path.Join(uploadPath, randFileName)

				if err := c.SaveFile(f, saveFilePath); err != nil {
					return c.JSON(models.ApiError("保存文件错误" + err.Error()))
				} else {
					url := fmt.Sprintf("%s/file-uploads/%s", string(c.Request().Host()), randFileName)
					urls = append(urls, url)

					if extName == ".png" || extName == ".jpg" || extName == ".jpeg" || extName == "bmp" {
						composeFileName := "compress_" + randFileName
						composeFilePath := path.Join(uploadPath, composeFileName)
						// 压缩文件
						if err := utils.ComposeImg(saveFilePath, composeFilePath, 1000, false); err != nil {
							log.Println("文件压缩失败", err.Error())
						} else {
							// 压缩文件成功
							log.Println("压缩文件成功")
						}
					}
				}
			}
		}

		return c.JSON(models.ApiSuccess(urls))
	}
}

// 根据文件sha1和size, 查询数据库中是否存在这么一个文件
func findFileInfo(size int, sha string) *models.FileInfo {
	var fileInfo models.FileInfo
	queryResult := utils.DbConnect.Model(&models.FileInfo{}).Where("sha_key = ? and size = ?", sha, size).Take(&fileInfo)
	if queryResult.Error != nil {
		if errors.Is(queryResult.Error, gorm.ErrRecordNotFound) {
			return nil
		} else {
			log.Println("查询文件信息错误!", queryResult.Error)
			return nil
		}
	} else {
		return &fileInfo
	}
}
