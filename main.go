package main

import (
	"embed"
	"flag"
	"fmt"
	"gfs/appinit"
	"gfs/handlers"
	"io"
	"log"
	"path"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var starttime = time.Now().Format("2006-01-02 15:04:05")

//go:embed static
var staticFS embed.FS

func main() {

	// app初始化方法
	appinit.AppInit(&staticFS) // 将embed.FS对象的内存地址传给appinit方法

	// 获取port参数,如果没有则默认使用8080
	port := flag.Int("port", 8080, "使用-port=8080设置服务启动参数")
	useSSL := flag.Bool("useSSL", false, "是否使用SSL")
	certPath := flag.String("certPath", "", "公钥证书文件路径")
	keyPath := flag.String("keyPath", "", "私钥证书文件路径")
	// 执行解析参数
	flag.Parse()

	// 构建一个fiber实例
	app := fiber.New(fiber.Config{
		// json编解码使用go-json更快
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		BodyLimit:   2 * 1024 * 1024 * 1024, //最大可以上传2G
	})

	// app.Use(utils.GetIP)

	// 配置fiber的http请求日志
	app.Use(logger.New(logger.Config{
		Output:     io.MultiWriter(appinit.AccessLogWrite), // 打印到文件
		TimeFormat: "2006-01-02 15:04:05",                  // Go语言的时间格式化与其他语言不同，它使用一个特定的时间点“2006年1月2日15时04分05秒”来代表格式化模板，其中每个数字部分代表不同的时间单位
		Format:     "${time} | ${status} | ${latency} | ${ips} | ${method} | ${path} | ${error}\n",
	}))

	// 允许跨域
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	// 避免意外导致程序退出
	app.Use(recover.New())

	// 分析性能
	// app.Use(pprof.New())

	// 映射一个静态资源目录
	app.Static("/static", path.Join(appinit.BaseDir, "static"), fiber.Static{
		Browse: true,
	})
	// 上传的文件
	app.Static("/file-uploads", path.Join(appinit.BaseDir, "file-uploads"), fiber.Static{
		Browse: true,
	})

	// 默认的路由
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(path.Join(appinit.BaseDir, "static/index.html"))
		// return c.SendString("Hello, World!")
	})

	// 返回当前进程启动时间
	app.Get("/starttime", func(c *fiber.Ctx) error {
		return c.JSON(starttime)
	})

	// 内存信息路由
	memGroup := app.Group("/mem")
	memGroup.Get("/gc", handlers.Gc)
	memGroup.Get("/info", handlers.GoMemInfo)
	// cpu信息路由
	app.All("/cpuInfo", handlers.CpuInfo)
	app.All("/memInfo", handlers.MemInfo)
	// tcp连接数路由
	app.All("/tcpInfo", handlers.TcpInfo)

	// 每日文件数量统计路由
	app.All("/todayFileTotal", handlers.TodayFileTotal)

	// 获取签名
	app.Post("/sign/getSign", handlers.GetSign)
	// 文件上传
	app.Post("/upload", handlers.UploadNoName)
	app.Post("/uploadImgs", handlers.UploadReturnName)
	app.Post("/uploadNotCompress", handlers.UploadNotCompress)

	// 此处port是一个指针, 访问对应的值需要使用*port
	address := fmt.Sprintf(":%d", *port)
	log.Printf("listen address:%s", address)

	if *useSSL {
		log.Printf("使用SSL配置:certPath:%s, keyPath:%s", *certPath, *keyPath)
		if err := app.ListenTLS(address, *certPath, *keyPath); err != nil {
			panic(err)
		}
	} else {
		if err := app.Listen(address); err != nil {
			panic(err)
		}
	}

}
