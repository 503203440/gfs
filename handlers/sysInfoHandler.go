package handlers

import (
	"encoding/json"
	"gfs/models"
	"gfs/utils"
	"log"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// 24小时 ≈ 86400 秒
const queueSize24h = 86400

// 创建一个队列
var cpuLoadQueue = utils.MyQueue{
	Size: queueSize24h,
}

var memInfoQueue = utils.MyQueue{
	Size: queueSize24h,
}

var tcpInfoQueue = utils.MyQueue{
	Size: queueSize24h,
}

var ServerPort int

// EMA 平滑系数，越小越平滑，0.3 是监控场景的常用值
const cpuEmaAlpha = 0.3

// 上一次 EMA 平滑后的 CPU 值
var cpuEmaValue float64

func saveMetric(metricType string, timestamp string, value any) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("序列化 %s 指标失败: %v", metricType, err)
		return
	}
	t, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		log.Printf("解析 %s 时间失败: %v", metricType, err)
		return
	}
	metric := models.SysMetric{
		Type:      metricType,
		Timestamp: t,
		Value:     string(jsonBytes),
	}
	if err := utils.DbConnect.Create(&metric).Error; err != nil {
		log.Printf("保存 %s 指标到数据库失败: %v", metricType, err)
	}
}

func loadMetricsFromDB(metricType string, limit int) []any {
	var metrics []models.SysMetric
	result := utils.DbConnect.Where("type = ?", metricType).
		Order("timestamp desc").
		Limit(limit).
		Find(&metrics)
	if result.Error != nil {
		log.Printf("从数据库加载 %s 指标失败: %v", metricType, result.Error)
		return nil
	}
	items := make([]any, 0, len(metrics))
	for i := len(metrics) - 1; i >= 0; i-- {
		var value map[string]any
		if err := json.Unmarshal([]byte(metrics[i].Value), &value); err != nil {
			continue
		}
		items = append(items, value)
	}
	return items
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

	// 从数据库恢复最近24小时数据到队列
	if utils.DbConnect != nil {
		for _, item := range loadMetricsFromDB("cpu", queueSize24h) {
			cpuLoadQueue.Enqueue(item)
		}
		for _, item := range loadMetricsFromDB("mem", queueSize24h) {
			memInfoQueue.Enqueue(item)
		}
		for _, item := range loadMetricsFromDB("tcp", queueSize24h) {
			tcpInfoQueue.Enqueue(item)
		}
		log.Println("已从数据库恢复历史监控数据")
	}

	go func() {
		for {
			// do something
			percent, err := cpu.Percent(time.Second, false)
			nowTimeStr := time.Now().Format("2006-01-02 15:04:05")
			if err == nil && len(percent) > 0 {
				// EMA 平滑：smoothed = α * raw + (1 - α) * prevSmoothed
				raw := percent[0]
				cpuEmaValue = cpuEmaAlpha*raw + (1-cpuEmaAlpha)*cpuEmaValue
				// 写入平滑后的值到队列
				item := map[string]any{
					"time":    nowTimeStr,
					"cpuLoad": cpuEmaValue,
				}
				cpuLoadQueue.Enqueue(item)
				saveMetric("cpu", nowTimeStr, item)
			}
			memInfo, err := mem.VirtualMemory()
			if err == nil {
				totalMemory := float64(memInfo.Total) / 1024 / 1024 / 1024
				usedMemory := float64(memInfo.Used) / 1024 / 1024 / 1024
				availableMemory := float64(memInfo.Available) / 1024 / 1024 / 1024
				usedPercent := memInfo.UsedPercent
				item := map[string]any{
					"time": nowTimeStr,
					"memInfo": map[string]any{
						"totalMemory":     totalMemory,
						"usedMemory":      usedMemory,
						"availableMemory": availableMemory,
						"usedPercent":     usedPercent,
					},
				}
				memInfoQueue.Enqueue(item)
				saveMetric("mem", nowTimeStr, item)
			}

			tcpConns, err := net.Connections("tcp")
			if err == nil && ServerPort > 0 {
				count := 0
				for _, conn := range tcpConns {
					if conn.Laddr.Port == uint32(ServerPort) {
						count++
					}
				}
				tcpItem := map[string]any{
					"time":    nowTimeStr,
					"tcpConn": count,
				}
				tcpInfoQueue.Enqueue(tcpItem)
				saveMetric("tcp", nowTimeStr, tcpItem)
			}
			// 这里就不要time.Sleep了,因为cpu.Percent(time.Second, false)方法本身就会阻塞1秒钟
			// time.Sleep(time.Second)
		}
	}()

}

// go的内存使用信息
func GoMemInfo(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocValue := m.Alloc / 1024 / 1024
	sysValue := m.Sys / 1024 / 1024
	log.Printf("runtime alloc = %v Mib, sys= %v Mib \n", allocValue, sysValue)
	return c.JSON(fiber.Map{"alloc(Mib)": allocValue, "sys(Mib)": sysValue})
}

// 触发GC
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

// 降采样：数据超过 maxPoints 条时，均匀跳着取，保证返回 ≈maxPoints 条
func downsample(data []any, maxPoints int) []any {
	if len(data) <= maxPoints {
		return data
	}
	step := float64(len(data)) / float64(maxPoints)
	result := make([]any, 0, maxPoints)
	for i := 0; i < len(data); i++ {
		if float64(i) >= float64(len(result))*step-0.5 {
			result = append(result, data[i])
		}
	}
	return result
}

// cpu使用率
func CpuInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(cpuLoadQueue.List(), 500))
}

// 系统内存使用率
func MemInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(memInfoQueue.List(), 500))
}

// TCP连接数
func TcpInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(tcpInfoQueue.List(), 500))
}
