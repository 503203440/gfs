package handlers

import (
	"encoding/json"
	"gfs/models"
	"gfs/utils"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// 采样间隔 5 秒，24h = 17280 条
const (
	sampleInterval = 5 * time.Second
	queueSize24h   = 17280
	batchInterval  = 60 * time.Second
)

var ServerPort int

// 使用具体类型替代 map[string]any，大幅减少内存分配
type cpuSample struct {
	Time    string  `json:"time"`
	CpuLoad float64 `json:"cpuLoad"`
}

type memSample struct {
	Time    string `json:"time"`
	MemInfo struct {
		TotalMemory     float64 `json:"totalMemory"`
		UsedMemory      float64 `json:"usedMemory"`
		AvailableMemory float64 `json:"availableMemory"`
		UsedPercent     float64 `json:"usedPercent"`
	} `json:"memInfo"`
}

type tcpSample struct {
	Time    string `json:"time"`
	TcpConn int    `json:"tcpConn"`
}

// 环形缓冲区：固定容量，写入时覆盖最旧数据，无 slice 重分配
type ringBuffer struct {
	buf   []any
	size  int
	head  int
	count int
	mu    sync.RWMutex
}

func newRingBuffer(capacity int) *ringBuffer {
	return &ringBuffer{buf: make([]any, capacity), size: capacity}
}

func (r *ringBuffer) push(v any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = v
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *ringBuffer) list() []any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.count == 0 {
		return nil
	}
	out := make([]any, 0, r.count)
	start := (r.head - r.count + r.size) % r.size
	for i := 0; i < r.count; i++ {
		out = append(out, r.buf[(start+i)%r.size])
	}
	return out
}

var (
	cpuLoadQueue *ringBuffer
	memInfoQueue *ringBuffer
	tcpInfoQueue *ringBuffer

	// 批量写入缓冲
	pendingMetrics   []models.SysMetric
	pendingMetricsMu sync.Mutex
)

// 降采样：数据超过 maxPoints 条时，均匀跳着取
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

func flushMetrics() {
	pendingMetricsMu.Lock()
	defer pendingMetricsMu.Unlock()
	if len(pendingMetrics) == 0 {
		return
	}
	if err := utils.DbConnect.Create(&pendingMetrics).Error; err != nil {
		log.Printf("批量保存监控指标失败: %v", err)
	}
	pendingMetrics = pendingMetrics[:0]
}

// 从数据库恢复最近的数据到队列
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
		switch metricType {
		case "cpu":
			var v cpuSample
			if err := json.Unmarshal([]byte(metrics[i].Value), &v); err == nil {
				items = append(items, v)
			}
		case "mem":
			var v memSample
			if err := json.Unmarshal([]byte(metrics[i].Value), &v); err == nil {
				items = append(items, v)
			}
		case "tcp":
			var v tcpSample
			if err := json.Unmarshal([]byte(metrics[i].Value), &v); err == nil {
				items = append(items, v)
			}
		}
	}
	return items
}

func init() {
	cpuLoadQueue = newRingBuffer(queueSize24h)
	memInfoQueue = newRingBuffer(queueSize24h)
	tcpInfoQueue = newRingBuffer(queueSize24h)
	pendingMetrics = make([]models.SysMetric, 0, 60)

	// 从数据库恢复历史数据
	if utils.DbConnect != nil {
		for _, item := range loadMetricsFromDB("cpu", queueSize24h) {
			cpuLoadQueue.push(item)
		}
		for _, item := range loadMetricsFromDB("mem", queueSize24h) {
			memInfoQueue.push(item)
		}
		for _, item := range loadMetricsFromDB("tcp", queueSize24h) {
			tcpInfoQueue.push(item)
		}
		log.Println("已从数据库恢复历史监控数据")
	}

	batchTicker := time.NewTicker(batchInterval)

	go func() {
		defer batchTicker.Stop()

		for {
			nowTimeStr := time.Now().Format("2006-01-02 15:04:05")

			// CPU：阻塞 sampleInterval 秒，返回该时段真实平均值
			percent, err := cpu.Percent(sampleInterval, false)
			if err == nil && len(percent) > 0 {
				item := cpuSample{Time: nowTimeStr, CpuLoad: percent[0]}
				cpuLoadQueue.push(item)
				queueMetric("cpu", nowTimeStr, item)
			}

			// Memory（滞后 <0.1s，对趋势图无影响）
			memInfo, err := mem.VirtualMemory()
			if err == nil {
				var m memSample
				m.Time = nowTimeStr
				m.MemInfo.TotalMemory = float64(memInfo.Total) / 1024 / 1024 / 1024
				m.MemInfo.UsedMemory = float64(memInfo.Used) / 1024 / 1024 / 1024
				m.MemInfo.AvailableMemory = float64(memInfo.Available) / 1024 / 1024 / 1024
				m.MemInfo.UsedPercent = memInfo.UsedPercent
				memInfoQueue.push(m)
				queueMetric("mem", nowTimeStr, m)
			}

			// TCP
			if ServerPort > 0 {
				tcpConns, err := net.Connections("tcp")
				if err == nil {
					count := 0
					for _, conn := range tcpConns {
						if conn.Laddr.Port == uint32(ServerPort) {
							count++
						}
					}
					item := tcpSample{Time: nowTimeStr, TcpConn: count}
					tcpInfoQueue.push(item)
					queueMetric("tcp", nowTimeStr, item)
				}
			}

			// 批量 flush 由独立 ticker 触发
			select {
			case <-batchTicker.C:
				flushMetrics()
			default:
			}
		}
	}()
}

// 将指标加入批量写入缓冲（线程安全）
func queueMetric(metricType string, timestamp string, value any) {
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
	pendingMetricsMu.Lock()
	defer pendingMetricsMu.Unlock()
	pendingMetrics = append(pendingMetrics, models.SysMetric{
		Type:      metricType,
		Timestamp: t,
		Value:     string(jsonBytes),
	})
}

func GoMemInfo(c *fiber.Ctx) error {
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
	runtime.GC()
	runtime.ReadMemStats(&m)
	after := m.Alloc / 1024 / 1024
	log.Printf("after gc alloc = %v Mib\n", after)
	return c.JSON(fiber.Map{"before": before, "after": after})
}

func CpuInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(cpuLoadQueue.list(), 500))
}

func MemInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(memInfoQueue.list(), 500))
}

func TcpInfo(c *fiber.Ctx) error {
	return c.JSON(downsample(tcpInfoQueue.list(), 500))
}
