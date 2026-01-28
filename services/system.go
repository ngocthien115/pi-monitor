package services

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

type SystemInfo struct {
	CPU       CPUInfo
	Memory    MemoryInfo
	Disk      DiskInfo
	Network   NetworkInfo
	Uptime    string
	Timestamp string
}

type CPUInfo struct {
	UsagePercent float64
	Temperature  float64
	Cores        int
	Frequency    float64
}

type MemoryInfo struct {
	Total       string
	Used        string
	Available   string
	UsedPercent float64
}

type DiskInfo struct {
	Total       string
	Used        string
	Free        string
	UsedPercent float64
}

type NetworkInfo struct {
	IP        string
	BytesSent string
	BytesRecv string
}

func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Timestamp: time.Now().Format("02/01/2006 15:04:05"),
	}

	// CPU Info
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPU.UsagePercent = cpuPercent[0]
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPU.Cores = int(cpuInfo[0].Cores)
		info.CPU.Frequency = cpuInfo[0].Mhz
	}

	// CPU Temperature (Raspberry Pi specific)
	info.CPU.Temperature = getCPUTemperature()

	// Memory Info
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.Memory.Total = formatBytes(memInfo.Total)
		info.Memory.Used = formatBytes(memInfo.Used)
		info.Memory.Available = formatBytes(memInfo.Available)
		info.Memory.UsedPercent = memInfo.UsedPercent
	}

	// Disk Info
	diskInfo, err := disk.Usage("/")
	if err == nil {
		info.Disk.Total = formatBytes(diskInfo.Total)
		info.Disk.Used = formatBytes(diskInfo.Used)
		info.Disk.Free = formatBytes(diskInfo.Free)
		info.Disk.UsedPercent = diskInfo.UsedPercent
	}

	// Network Info
	info.Network.IP = getLocalIP()
	netIO, err := psnet.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		info.Network.BytesSent = formatBytes(netIO[0].BytesSent)
		info.Network.BytesRecv = formatBytes(netIO[0].BytesRecv)
	}

	// Uptime
	uptime, err := host.Uptime()
	if err == nil {
		info.Uptime = formatDuration(uptime)
	}

	return info, nil
}

func getCPUTemperature() float64 {
	// Try reading from Raspberry Pi thermal zone
	paths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/host/sys/class/thermal/thermal_zone0/temp", // When mounted in Docker
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			tempStr := strings.TrimSpace(string(data))
			temp, err := strconv.ParseFloat(tempStr, 64)
			if err == nil {
				return temp / 1000.0 // Convert from millidegrees
			}
		}
	}

	return 0.0
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "N/A"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "N/A"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%d ngày %d giờ %d phút", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d giờ %d phút", hours, minutes)
	}
	return fmt.Sprintf("%d phút", minutes)
}
