package handlers

import (
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

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
