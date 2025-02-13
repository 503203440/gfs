package handlers

import (
	"gfs/utils"
	"log"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/cpu"
)

// 创建一个队列
var cpuLoadQueue = utils.MyQueue{
	Size: 10,
}

func init() {
	// 方案一: 使用time.Ticker实现定时
	// ticker := time.NewTicker(time.Second)
	// go func() {
	// 	for range ticker.C {
	// do something
	// 	}
	// }()

	// 方案二: 使用time.Sleep
	go func() {
		for {
			// do something
			percent, err := cpu.Percent(time.Second, false)
			if err == nil && len(percent) > 0 {
				// 写入一个进入队列
				item := map[string]any{
					"time":    time.Now().Format("2006-01-02 15:04:05"),
					"cpuLoad": percent[0],
				}
				cpuLoadQueue.Enqueue(item)
			}
			// 这里就不要time.Sleep了,因为cpu.Percent(time.Second, false)方法本身就会阻塞1秒钟
			// time.Sleep(time.Second)
		}
	}()

}

func MemInfo(c *fiber.Ctx) error {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	allocValue := m.Alloc / 1024 / 1024
	sysValue := m.Sys / 1024 / 1024

	log.Printf("runtime alloc = %v Mib, sys= %v Mib \n", allocValue, sysValue)

	return c.JSON(fiber.Map{"alloc(Mib)": allocValue, "sys(Mib)": sysValue})
}

func Gc(c *fiber.Ctx) error {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	before := m.Alloc / 1024 / 1024
	log.Printf("before gc alloc = %v Mib\n", before)

	// 手动触发gc
	runtime.GC()

	runtime.ReadMemStats(&m)
	after := m.Alloc / 1024 / 1024
	log.Printf("after gc alloc = %v Mib\n", after)

	return c.JSON(fiber.Map{"before": before, "after": after})
}

func CpuInfo(c *fiber.Ctx) error {
	return c.JSON(cpuLoadQueue.List())
}
