package appinit

import (
	"embed"
	"gofs/models"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var staticFS embed.FS

func AppInit(fs *embed.FS) (*os.File, *os.File) {

	staticFS = *fs

	// 初始化日志
	LogFile, accessLogFile := initLogs("./gofs.log", "./access.log")

	// 释放static文件目录
	extractStatic()

	return LogFile, accessLogFile
}

// go内嵌当前程序下某个文件夹或文件的命令
// 注意//go:embed直接不能有空格
//

// 获取可执行文件所在目录
var exePath, _ = os.Executable()
var baseDir = filepath.Dir(exePath)

func extractStatic() {

	staticDir := filepath.Join(baseDir, "static")
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
		targetPath := filepath.Join(baseDir, path)
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

func initLogs(logFilePath, accessLogFilePath string) (*os.File, *os.File) {
	gofsLogFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("创建系统日志文件失败: " + err.Error())
	}
	log.SetOutput(models.NewMultiWrite(gofsLogFile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	accessLogfile, err := os.OpenFile(accessLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("创建访问日志文件失败: " + err.Error())
	}

	return gofsLogFile, accessLogfile
}
