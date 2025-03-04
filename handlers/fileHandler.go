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
	"path/filepath"
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

	// 启动一个goroutine
	go func() {
		for {
			err := filepath.Walk(uploadPath, func(path string, info os.FileInfo, err error) error {
				if err == nil {
					if !info.IsDir() {
						modTime := info.ModTime()
						timeDiff := time.Since(modTime)
						if timeDiff.Hours() > 24*2 {
							fileName := info.Name()
							log.Printf("删除文件:filename:%s,modTime:%s", fileName, modTime.Format("2006-01-02 15:04:05"))
							os.Remove(path)
						}
					}
				}
				return nil
			})
			if err != nil {
				log.Println("监听文件夹失败!", err)
				break
			}
			// 每小时执行一次
			time.Sleep(time.Hour)
		}
	}()

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
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(models.ApiError("未找到客户端信息"))
			} else {
				return c.JSON(models.ApiErrorDetail("查询数据错误", "999", result.Error.Error()))
			}
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

		return c.JSON(models.ApiSuccess(tokenEntity))
	}

}

// 返回结果中不包含名称
func UploadNoName(c *fiber.Ctx) error {
	// token检查
	tokenEntity, err := tokenCheck(c)
	if err != nil {
		return c.JSON(models.ApiError(err.Error()))
	}
	urls := make([]string, 0)
	if multipartForm, err := c.MultipartForm(); err != nil {
		log.Println("表单读取错误:", err)
		return c.JSON(models.ApiError("读取文件错误" + err.Error()))
	} else {
		files := append(multipartForm.File["file"], multipartForm.File["files"]...)
		for _, f := range files {
			// 将文件保存到本地目录
			extName := strings.ToLower(path.Ext(f.Filename))
			randFileName := utils.GenerateRandomString(20) + extName
			saveFilePath := path.Join(uploadPath, randFileName)
			if err := c.SaveFile(f, saveFilePath); err != nil {
				return c.JSON(models.ApiError("保存文件错误" + err.Error()))
			} else {
				// 计算文件sha, 查询此文件是否已经上传过
				sha1, genShaErr := utils.GenerateFileSHA1(saveFilePath)
				if genShaErr == nil {
					// 查询是否存符合条件的数据
					fileInfo := findFileInfo(f.Size, *sha1)
					if fileInfo != nil {
						ossUrl := fileInfo.URL
						urls = append(urls, ossUrl)
						// 更新引用次数
						referenceIncrement(fileInfo)
						continue // 此文件已经存在, 直接返回
					}
				}
				// 若oss配置存在
				if utils.OssClient != nil {
					// 上传到oss
					ossKey := path.Join(utils.OssFolder, time.Now().Format("2006-01"), randFileName)
					ossUrl, err := utils.UploadFile(ossKey, saveFilePath)
					if err != nil {
						log.Println("oss上传失败", err)
						return c.JSON(models.ApiError("上传OSS失败" + err.Error()))
					}
					// 是图片且可以压缩
					composeUrl, composeErr := composeAndUpload(saveFilePath, randFileName)
					if composeErr == nil {
						ossUrl = *composeUrl
					}
					urls = append(urls, ossUrl)
					// 记录文件sha1信息
					if genShaErr == nil {
						fileInfo := models.FileInfo{ShaKey: *sha1, Size: f.Size, URL: ossUrl, CreateTime: time.Now(), Reference: 1}
						saveFileInfo(&fileInfo)
					}
				} else {
					// 未配置oss,直接返回当前服务对应的路径
					url := fmt.Sprintf("%s/static/%s", string(c.Request().Host()), f.Filename)
					urls = append(urls, url)
				}
			}
		}
		// 更新token使用情况
		usedToken(tokenEntity.ID)
	}
	// 如果本次上传的是多个文件则返回数组结构, 如果只有一个文件则直接返回此文件url
	if len(urls) == 1 {
		return c.JSON(models.ApiSuccess(urls[0]))
	} else {
		return c.JSON(models.ApiSuccess(urls))
	}

}

// 返回结果包含名称
func UploadReturnName(c *fiber.Ctx) error {
	// token检查
	tokenEntity, err := tokenCheck(c)
	if err != nil {
		return c.JSON(models.ApiError(err.Error()))
	}
	resultUrls := make([]any, 0)
	if multipartForm, err := c.MultipartForm(); err != nil {
		log.Println("表单读取错误:", err)
		return c.JSON(models.ApiError("读取文件错误" + err.Error()))
	} else {
		imgs := multipartForm.File["imgs"]
		for _, img := range imgs {
			filename := img.Filename
			item := make(map[string]any)
			item["fileName"] = filename
			extName := strings.ToLower(path.Ext(filename))
			randFileName := utils.GenerateRandomString(20) + extName
			saveFilePath := path.Join(uploadPath, randFileName)
			if err := c.SaveFile(img, saveFilePath); err != nil {
				return c.JSON(models.ApiError("保存文件错误" + err.Error()))
			} else {
				// 计算文件sha, 查询此文件是否已经上传过
				sha1, genShaErr := utils.GenerateFileSHA1(saveFilePath)
				if genShaErr == nil {
					// 查询是否存符合条件的数据
					fileInfo := findFileInfo(img.Size, *sha1)
					if fileInfo != nil {
						ossUrl := fileInfo.URL
						item["url"] = ossUrl
						resultUrls = append(resultUrls, item)
						// 更新引用次数
						referenceIncrement(fileInfo)
						continue // 此文件已经存在, 直接返回
					}
				}
				// 若oss配置存在
				if utils.OssClient != nil {
					// 上传到oss
					ossKey := path.Join(utils.OssFolder, time.Now().Format("2006-01"), randFileName)
					ossUrl, err := utils.UploadFile(ossKey, saveFilePath)
					if err != nil {
						log.Println("oss上传失败", err)
						return c.JSON(models.ApiError("上传OSS失败" + err.Error()))
					}
					// 是图片且可以压缩
					composeUrl, composeErr := composeAndUpload(saveFilePath, randFileName)
					if composeErr == nil {
						ossUrl = *composeUrl
					}

					item["url"] = ossUrl
					// 记录文件sha1信息
					if genShaErr == nil {
						fileInfo := models.FileInfo{ShaKey: *sha1, Size: img.Size, URL: ossUrl, CreateTime: time.Now(), Reference: 1}
						saveFileInfo(&fileInfo)
					}
				} else {
					// 未配置oss,直接返回当前服务对应的路径
					url := fmt.Sprintf("%s/static/%s", string(c.Request().Host()), filename)
					item["url"] = url
				}
			}
			resultUrls = append(resultUrls, item)
		}
		// 更新token使用情况
		usedToken(tokenEntity.ID)
	}
	return c.JSON(models.ApiSuccess(resultUrls))
}

// 不压缩版本
func UploadNotCompress(c *fiber.Ctx) error {
	tokenEntity, err := tokenCheck(c)
	if err != nil {
		return c.JSON(models.ApiError(err.Error()))
	}
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(models.ApiError("读取文件错误"))
	}
	// 将文件保存到磁盘
	sourceFileName := file.Filename
	extName := strings.ToLower(path.Ext(sourceFileName))
	randFileName := utils.GenerateRandomString(20) + extName
	saveFilePath := path.Join(uploadPath, randFileName)

	if err := c.SaveFile(file, saveFilePath); err != nil {
		return c.JSON(models.ApiError("保存文件失败!"))
	}

	var url string
	// 计算文件sha, 查询此文件是否已经上传过
	sha1, genShaErr := utils.GenerateFileSHA1(saveFilePath)
	if genShaErr == nil {
		// 查询是否存符合条件的数据
		fileInfo := findFileInfo(file.Size, *sha1)
		if fileInfo != nil {
			// 如果是压缩文件路径则返回非压缩文件路径
			url = strings.Replace(fileInfo.URL, utils.OssFolderCompress, utils.OssFolder, 1)
			// 更新引用次数
			referenceIncrement(fileInfo)
		} else {
			if utils.OssClient != nil {
				// 上传到oss
				ossKey := path.Join(utils.OssFolder, time.Now().Format("2006-01"), randFileName)
				url, err = utils.UploadFile(ossKey, saveFilePath)
				if err != nil {
					log.Println("oss上传失败", err)
					return c.JSON(models.ApiError("上传OSS失败" + err.Error()))
				}
				// 尝试压缩并得到压缩版URL,这里虽然用不着
				composeUrl, err := composeAndUpload(saveFilePath, randFileName)
				if err == nil {
					// 记录文件sha1信息
					fileInfo := models.FileInfo{ShaKey: *sha1, Size: file.Size, URL: *composeUrl, CreateTime: time.Now(), Reference: 1}
					saveFileInfo(&fileInfo)
				}
			} else {
				// 未配置oss,直接返回当前服务对应的路径
				url = fmt.Sprintf("%s/static/%s", string(c.Request().Host()), file.Filename)
			}
		}
	} else {
		if utils.OssClient != nil {
			// 上传到oss
			ossKey := path.Join(utils.OssFolder, time.Now().Format("2006-01"), randFileName)
			url, err = utils.UploadFile(ossKey, saveFilePath)
			if err != nil {
				log.Println("oss上传失败", err)
				return c.JSON(models.ApiError("上传OSS失败" + err.Error()))
			}
		} else {
			// 未配置oss,直接返回当前服务对应的路径
			url = fmt.Sprintf("%s/static/%s", string(c.Request().Host()), file.Filename)
		}
	}
	// 更新token使用情况
	usedToken(tokenEntity.ID)

	return c.JSON(models.ApiSuccess(url))
}

// 压缩并上传文件, 返回压缩后的url
func composeAndUpload(sourceImgPath, randFileName string) (*string, error) {
	extName := path.Ext(sourceImgPath)
	if extName == ".png" || extName == ".jpg" || extName == ".jpeg" || extName == "bmp" {
		width, err := utils.ImageWidth(sourceImgPath)
		if width <= 1000 || err != nil {
			return nil, errors.New("文件宽度不足或无法读取文件宽度信息")
		}
		composeFileName := "compress_" + randFileName
		composeFilePath := path.Join(uploadPath, composeFileName)
		// 压缩文件
		if err := utils.ComposeImg(sourceImgPath, composeFilePath, 1000, false); err != nil {
			log.Println("文件压缩失败", err.Error())
			return nil, err
		} else {
			// 压缩文件成功
			ossKey := path.Join(utils.OssFolderCompress, time.Now().Format("2006-01"), randFileName)
			// 上传一份压缩版,并返回压缩版的值
			composeUrl, err := utils.UploadFile(ossKey, composeFilePath)
			return &composeUrl, err
		}
	} else {
		return nil, errors.New("文件类型不支持压缩")
	}
}

// token检查
func tokenCheck(c *fiber.Ctx) (*models.TokenEntity, error) {
	token := c.FormValue("token")
	if token == "" {
		return nil, errors.New("token不能为空")
	}
	var tokenEntity models.TokenEntity
	queryResult := utils.DbConnect.Model(&models.TokenEntity{}).Where("token = ?", token).Take(&tokenEntity)
	if queryResult.Error != nil {
		if errors.Is(queryResult.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("token无效 token:" + token)
		} else {
			log.Printf("查询发生错误!%v", queryResult.Error)
			return nil, errors.New("服务器错误:" + queryResult.Error.Error())
		}
	}
	if tokenEntity.Used {
		return nil, errors.New("token已使用")
	}
	return &tokenEntity, nil
}

// 更新token使用情况(参数tokenId)
func usedToken(id int64) {
	// 更新token使用情况
	updateResult := utils.DbConnect.Model(&models.TokenEntity{}).Where("id = ?", id).Update("used", true)
	if updateResult.Error != nil {
		log.Println("更新token使用情况错误!", updateResult.Error)
	}
}

// 保存文件信息
func saveFileInfo(fileInfo *models.FileInfo) {
	saveResult := utils.DbConnect.Model(&models.FileInfo{}).Save(&fileInfo)
	if saveResult.Error != nil {
		log.Println("保存文件引用信息失败!", saveResult)
	}
}

// 更新文件引用记录次数(fileInfo.ID)
func referenceIncrement(fileInfo *models.FileInfo) {
	utils.DbConnect.Model(&models.FileInfo{}).Where("id = ?", fileInfo.ID).Updates(map[string]any{
		"reference":   fileInfo.Reference + 1,
		"update_time": time.Now(),
	})
}

// 根据文件sha1和size, 查询数据库中是否存在这么一个文件
func findFileInfo(size int64, sha string) *models.FileInfo {
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
