package utils

import (
	"context"
	"gfs/appinit"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 公用ossClient对象
var OssClient *oss.Client
var bucketName string = "test50"                // 默认bucketName
var OssFolder string = "GPAI5"                  // 默认文件夹
var OssFolderCompress string = "GPAI5_Compress" // 默认压缩文件夹

// 替换为我司cdn域名
var reg, _ = regexp.Compile(`\b[a-z0-9A-Z-\.]+\.aliyuncs\.com`)

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

	accessKeyId := ossConfigMap["oss.accessKeyId"]
	accessKeySecret := ossConfigMap["oss.accessKeySecret"]
	region := ossConfigMap["oss.region"]
	endpoint := ossConfigMap["oss.endpoint"]

	// 初始化oss配置
	provider := credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret)
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(region).WithEndpoint(endpoint)
	OssClient = oss.NewClient(cfg)

}

func UploadFile(objectName, localFilePath string) (string, error) {

	file, err := os.OpenFile(localFilePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Println("打开文件失败!", err)
		return "", err
	}

	log.Println(bucketName, objectName)
	_, err = OssClient.PutObject(context.TODO(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
		Body:   file,
	})
	if err != nil {
		log.Println("文件上传失败!", err)
		return "", err
	}

	urlInfo, err := OssClient.Presign(
		context.TODO(),
		&oss.GetObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName),
		},
		oss.PresignExpires(time.Hour*24*7), //expires should be not greater than 604800(seven days)过期时间不能大于7天,没啥鸟用
	)
	if err != nil {
		return "", err
	}
	// 截断字符串, 去除后面的有效期等参数
	fileUrl := urlInfo.URL[:strings.Index(urlInfo.URL, "?")]

	fileUrl = reg.ReplaceAllString(fileUrl, "cdnimg.gpai.net")

	log.Println("upload result:", fileUrl)

	return fileUrl, nil
}
