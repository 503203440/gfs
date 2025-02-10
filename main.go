package main

import (
	"gofs/handlers"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {

	// 打开日志文件
	file, err := os.OpenFile("./access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Output: &MultiWriter{
			writers: []io.Writer{os.Stdout, file},
		},
		TimeFormat: "2006-01-02 15:04:05", // Go语言的时间格式化与其他语言不同，它使用一个特定的时间点“2006年1月2日15时04分05秒”来代表格式化模板，其中每个数字部分代表不同的时间单位
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))
	app.Use(recover.New())

	app.Static("/static", "./static", fiber.Static{
		Browse: true,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		//return c.SendFile("./static/index.html")
		return c.SendString("Hello, World!")
	})

	userGroup := app.Group("/user")
	userGroup.Get("/list", handlers.GetUsers)
	userGroup.Post("/create", handlers.CreateUser)

	if err := app.Listen(":8080"); err != nil {
		panic(err)
	} else {
		log.Printf("Listening on port 8080")
	}

}

// 定义一个结构实现io.Writer接口
type MultiWriter struct {
	writers []io.Writer
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}
