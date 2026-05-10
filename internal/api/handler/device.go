package handler

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type DeviceInfo struct {
	OS        string  `json:"os"`
	Arch      string  `json:"arch"`
	Hostname  string  `json:"hostname"`
	CPUModel  string  `json:"cpu_model"`
	CPUCores  int     `json:"cpu_cores"`
	CPUUsage  float64 `json:"cpu_usage"`
	RAMTotal  uint64  `json:"ram_total"`
	RAMUsed   uint64  `json:"ram_used"`
	RAMFree   uint64  `json:"ram_free"`
	RAMUsage  float64 `json:"ram_usage"`
	Uptime    uint64  `json:"uptime"`
}

func (h *Handler) GetDeviceInfo(c *gin.Context) {
	v, _ := mem.VirtualMemory()
	cPercent, _ := cpu.Percent(0, false)
	hInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()

	cpuModel := "Unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	usage := 0.0
	if len(cPercent) > 0 {
		usage = cPercent[0]
	}

	info := DeviceInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hInfo.Hostname,
		CPUModel: cpuModel,
		CPUCores: runtime.NumCPU(),
		CPUUsage: usage,
		RAMTotal: v.Total,
		RAMUsed:  v.Used,
		RAMFree:  v.Free,
		RAMUsage: v.UsedPercent,
		Uptime:   hInfo.Uptime,
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": info,
	})
}
