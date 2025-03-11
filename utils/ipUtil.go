package utils

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// 获取代理后的真实ip地址
func GetIP(c *fiber.Ctx) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	} else {
		// 如果有多个 IP，取第一个（通常是客户端 IP）
		ip = strings.Split(ip, ",")[0]
	}
	// 将真实的 IP 添加到上下文
	ctx := context.WithValue(c.UserContext(), "ip", ip)
	c.SetUserContext(ctx)
	// 继续处理请求
	return c.Next()
}
