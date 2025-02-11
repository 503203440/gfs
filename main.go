package main

import (
	"flag"
	"fmt"
	"gofs/handlers"
	"gofs/models"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {

	// 获取port参数,如果没有则默认使用8080
	port := flag.Int("port", 8080, "使用-port=8080设置服务启动参数")
	// 执行解析参数
	flag.Parse()

	// 打开日志文件
	file, err := os.OpenFile("./access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 构建一个fiber实例
	app := fiber.New()

	// 配置fiber的http请求日志
	app.Use(logger.New(logger.Config{
		Output:     models.NewMultiWrite(file), // 打印到文件
		TimeFormat: "2006-01-02 15:04:05",      // Go语言的时间格式化与其他语言不同，它使用一个特定的时间点“2006年1月2日15时04分05秒”来代表格式化模板，其中每个数字部分代表不同的时间单位
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))

	// 创建日志文件
	gofsLogFile, err := os.OpenFile("./gofs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("创建系统日志文件失败!" + err.Error())
	}
	defer gofsLogFile.Close()
	log.SetOutput(models.NewMultiWrite(gofsLogFile, os.Stdout)) // 系统日志打印到文件和控制台
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)        // 设置输出格式

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
		//return c.SendFile("./static/index.html")
		return c.SendString("Hello, World!")
	})

	// 内存路由
	memGroup := app.Group("/mem")
	memGroup.Get("/info", handlers.MemInfo)
	memGroup.Get("/gc", handlers.Gc)

	// user路由
	userGroup := app.Group("/user")
	userGroup.Get("/list", handlers.GetUsers)
	userGroup.Post("/create", handlers.CreateUser)

	// 每日文件数量统计路由
	app.All("/todayFileTotal", handlers.TodayFileTotal)

	// 此处port是一个指针, 访问对应的值需要使用*port
	address := fmt.Sprintf(":%d", *port)
	log.Printf("listen address:%s", address)
	if err := app.Listen(address); err != nil {
		panic(err)
	}

}
