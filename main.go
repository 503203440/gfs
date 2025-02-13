package main

import (
	"embed"
	"flag"
	"fmt"
	"gofs/handlers"
	"gofs/models"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {

	LogFile, accessLogFile := initLogs("./gofs.log", "./access.log")
	// main函数结束时释放文件
	defer accessLogFile.Close()
	defer LogFile.Close()

	// 获取port参数,如果没有则默认使用8080
	port := flag.Int("port", 8080, "使用-port=8080设置服务启动参数")
	// 执行解析参数
	flag.Parse()

	// 构建一个fiber实例
	app := fiber.New()

	// 配置fiber的http请求日志
	app.Use(logger.New(logger.Config{
		Output:     models.NewMultiWrite(accessLogFile), // 打印到文件
		TimeFormat: "2006-01-02 15:04:05",               // Go语言的时间格式化与其他语言不同，它使用一个特定的时间点“2006年1月2日15时04分05秒”来代表格式化模板，其中每个数字部分代表不同的时间单位
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))

	// 避免意外导致程序退出
	app.Use(recover.New())

	// 分析性能
	// app.Use(pprof.New())

	// 释放静态文件到文件目录
	extractStatic()

	// 映射一个静态资源目录
	app.Static("/static", "./static", fiber.Static{
		Browse: true,
	})

	// 默认的路由
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
		// return c.SendString("Hello, World!")
	})

	// 内存信息路由
	memGroup := app.Group("/mem")
	memGroup.Get("/info", handlers.MemInfo)
	memGroup.Get("/gc", handlers.Gc)
	// cpu信息路由
	app.All("/cpuInfo", handlers.CpuInfo)

	// 每日文件数量统计路由
	app.All("/todayFileTotal", handlers.TodayFileTotal)

	// 此处port是一个指针, 访问对应的值需要使用*port
	address := fmt.Sprintf(":%d", *port)
	log.Printf("listen address:%s", address)
	if err := app.Listen(address); err != nil {
		panic(err)
	}

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

// go内嵌当前程序下某个文件夹或文件的命令
// 注意//go:embed直接不能有空格
//
//go:embed static
var staticFS embed.FS

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
