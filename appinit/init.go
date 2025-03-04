package appinit

import (
	"embed"
	"gfs/models"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var staticFS embed.FS

// app初始化方法,返回accessLog提供给fiber的access日志输出
func AppInit(fs *embed.FS) {

	staticFS = *fs

	// 设置系统日志输出
	log.SetOutput(models.NewMultiWrite(AppLogWrite, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 释放static文件目录
	extractStatic()

}

// go内嵌当前程序下某个文件夹或文件的命令
// 注意//go:embed直接不能有空格
//

// 获取可执行文件所在目录
var ExePath, _ = os.Executable()
var BaseDir = filepath.Dir(ExePath)
var gfsLogPath = filepath.Join(BaseDir, "gfs.log")
var accessLogPath = filepath.Join(BaseDir, "access.log")
var AppLogWrite = &lumberjack.Logger{
	Filename:   gfsLogPath, // 日志文件路径
	MaxSize:    100,        // 单个文件最大大小（MB）
	MaxBackups: 3,          // 最大保留备份文件数
	MaxAge:     30,         // 文件最大保留天数
	Compress:   true,       // 是否压缩旧文件
}
var AccessLogWrite = &lumberjack.Logger{
	Filename:   accessLogPath, // 日志文件路径
	MaxSize:    100,             // 单个文件最大大小（MB）
	MaxBackups: 3,             // 最大保留备份文件数
	MaxAge:     30,            // 文件最大保留天数
	Compress:   true,          // 是否压缩旧文件
}

func extractStatic() {

	staticDir := filepath.Join(BaseDir, "static")
	log.Printf("staticDir:%s", staticDir)

	// 如果static目录不存在，则创建并复制文件
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Println("释放静态资源到:", staticDir)
		err = copyStaticFiles(staticDir)
		if err != nil {
			log.Fatal("复制静态资源失败:", err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

}

func copyStaticFiles(staticDir string) error {
	// 创建static目录
	if err := os.Mkdir(staticDir, 0755); err != nil {
		return err
	}

	// 遍历嵌入的文件系统并复制文件
	return fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		log.Println("嵌入文件:", path)
		targetPath := filepath.Join(BaseDir, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		} else {
			data, err := staticFS.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(targetPath, data, 0644)
		}
	})
}
