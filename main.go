package main

import (
	"embed"
	"flag"
	"fmt"
	"gofs/appinit"
	"gofs/handlers"
	"gofs/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

//go:embed static
var staticFS embed.FS

func main() {

	// app初始化方法
	LogFile, accessLogFile := appinit.AppInit(&staticFS) // 将embed.FS对象的内存地址传给appinit方法
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
