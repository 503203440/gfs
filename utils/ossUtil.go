package utils

import (
	"gfs/appinit"
	"log"
	"os"
	"path"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 公用ossClient对象
var Client oss.Client
var bucketName string = "test50"             // 默认bucketName
var folder string = "GPAI5"                  // 默认文件夹
var folderCompress string = "GPAI5_Compress" // 默认压缩文件夹

func init() {
	// 从配置文件获取oss配置信息, 读取当前目录下的oss.properties文件
	ossPropertiesPath := path.Join(appinit.BaseDir, "oss.properties")

	if _, err := os.Stat(ossPropertiesPath); os.IsNotExist(err) {
		log.Println("当前目录未发现oss.properties")
		return
	} else if err != nil {
		log.Println("读取oss.properties失败", err)
		return
	}

	ossConfigMap, err := ReadProperties(ossPropertiesPath)
	if err != nil {
		log.Println("读取oss配置文件错误!", err)
		return
	}
	log.Println("ossConfigMap", ossConfigMap)

	accessKeyId := ossConfigMap["oss.accessKeyId"]
	accessKeySecret := ossConfigMap["oss.accessKeySecret"]
	endpoint := ossConfigMap["oss.endpoint"]

	// 初始化oss配置
	provider := credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret)
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithEndpoint(endpoint)
	Client = *oss.NewClient(cfg)

}

// 创建OSS文件夹,如果不存在,存在也不会报错
func CreateFolderIfNotExists(path string) error {
	// TODO

	return nil
}
