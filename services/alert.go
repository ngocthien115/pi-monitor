package services

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// AlertThresholds định nghĩa các ngưỡng cảnh báo
type AlertThresholds struct {
	CPUTemperature float64 // Nhiệt độ CPU (°C)
	CPUUsage       float64 // % sử dụng CPU
	MemoryUsage    float64 // % sử dụng RAM
	DiskUsage      float64 // % sử dụng Disk
}

// DefaultThresholds trả về các ngưỡng mặc định
func DefaultThresholds() AlertThresholds {
	return AlertThresholds{
		CPUTemperature: 70.0, // Cảnh báo khi > 70°C
		CPUUsage:       90.0, // Cảnh báo khi > 90%
		MemoryUsage:    85.0, // Cảnh báo khi > 85%
		DiskUsage:      90.0, // Cảnh báo khi > 90%
	}
}

// AlertType định nghĩa loại cảnh báo
type AlertType string

const (
	AlertCPUTemp  AlertType = "CPU_TEMPERATURE"
	AlertCPUUsage AlertType = "CPU_USAGE"
	AlertMemory   AlertType = "MEMORY_USAGE"
	AlertDisk     AlertType = "DISK_USAGE"
)

// Alert chứa thông tin cảnh báo
type Alert struct {
	Type      AlertType
	Value     float64
	Threshold float64
	Message   string
	Timestamp time.Time
}

// AlertChecker kiểm tra và phát hiện bất thường
type AlertChecker struct {
	Thresholds     AlertThresholds
	lastAlerts     map[AlertType]time.Time // Tracking để tránh spam
	cooldownPeriod time.Duration           // Thời gian chờ giữa các alert cùng loại
}

// NewAlertChecker tạo AlertChecker mới
func NewAlertChecker(thresholds AlertThresholds) *AlertChecker {
	return &AlertChecker{
		Thresholds:     thresholds,
		lastAlerts:     make(map[AlertType]time.Time),
		cooldownPeriod: 5 * time.Minute, // Chỉ alert lại sau 5 phút
	}
}

// CheckSystem kiểm tra hệ thống và trả về các cảnh báo
func (ac *AlertChecker) CheckSystem() ([]Alert, error) {
	info, err := GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %v", err)
	}

	var alerts []Alert
	now := time.Now()

	// Check CPU Temperature
	if info.CPU.Temperature > 0 && info.CPU.Temperature > ac.Thresholds.CPUTemperature {
		if ac.canAlert(AlertCPUTemp, now) {
			alerts = append(alerts, Alert{
				Type:      AlertCPUTemp,
				Value:     info.CPU.Temperature,
				Threshold: ac.Thresholds.CPUTemperature,
				Message:   fmt.Sprintf("🌡️ *Nhiệt độ CPU quá cao!*\n├ Hiện tại: *%.1f°C*\n└ Ngưỡng: %.1f°C", info.CPU.Temperature, ac.Thresholds.CPUTemperature),
				Timestamp: now,
			})
			ac.lastAlerts[AlertCPUTemp] = now
		}
	}

	// Check CPU Usage
	if info.CPU.UsagePercent > ac.Thresholds.CPUUsage {
		if ac.canAlert(AlertCPUUsage, now) {
			alerts = append(alerts, Alert{
				Type:      AlertCPUUsage,
				Value:     info.CPU.UsagePercent,
				Threshold: ac.Thresholds.CPUUsage,
				Message:   fmt.Sprintf("📈 *CPU đang quá tải!*\n├ Hiện tại: *%.1f%%*\n└ Ngưỡng: %.1f%%", info.CPU.UsagePercent, ac.Thresholds.CPUUsage),
				Timestamp: now,
			})
			ac.lastAlerts[AlertCPUUsage] = now
		}
	}

	// Check Memory Usage
	if info.Memory.UsedPercent > ac.Thresholds.MemoryUsage {
		if ac.canAlert(AlertMemory, now) {
			alerts = append(alerts, Alert{
				Type:      AlertMemory,
				Value:     info.Memory.UsedPercent,
				Threshold: ac.Thresholds.MemoryUsage,
				Message:   fmt.Sprintf("💾 *RAM sắp hết!*\n├ Đã dùng: *%.1f%%* (%s/%s)\n└ Ngưỡng: %.1f%%", info.Memory.UsedPercent, info.Memory.Used, info.Memory.Total, ac.Thresholds.MemoryUsage),
				Timestamp: now,
			})
			ac.lastAlerts[AlertMemory] = now
		}
	}

	// Check Disk Usage
	if info.Disk.UsedPercent > ac.Thresholds.DiskUsage {
		if ac.canAlert(AlertDisk, now) {
			alerts = append(alerts, Alert{
				Type:      AlertDisk,
				Value:     info.Disk.UsedPercent,
				Threshold: ac.Thresholds.DiskUsage,
				Message:   fmt.Sprintf("💿 *Ổ đĩa sắp đầy!*\n├ Đã dùng: *%.1f%%* (%s/%s)\n└ Ngưỡng: %.1f%%", info.Disk.UsedPercent, info.Disk.Used, info.Disk.Total, ac.Thresholds.DiskUsage),
				Timestamp: now,
			})
			ac.lastAlerts[AlertDisk] = now
		}
	}

	return alerts, nil
}

// canAlert kiểm tra xem có thể gửi alert không (cooldown)
func (ac *AlertChecker) canAlert(alertType AlertType, now time.Time) bool {
	lastTime, exists := ac.lastAlerts[alertType]
	if !exists {
		return true
	}
	return now.Sub(lastTime) >= ac.cooldownPeriod
}

// FormatAlerts format danh sách cảnh báo thành message
func FormatAlerts(alerts []Alert) string {
	if len(alerts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("🚨 *CẢNH BÁO HỆ THỐNG RASPBERRY PI*\n\n")

	for i, alert := range alerts {
		sb.WriteString(alert.Message)
		if i < len(alerts)-1 {
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString(fmt.Sprintf("\n\n⏰ _Thời gian: %s_", alerts[0].Timestamp.Format("02/01/2006 15:04:05")))

	return sb.String()
}

// StartMonitoring bắt đầu monitoring và gọi callback khi có alert
func StartMonitoring(checker *AlertChecker, interval time.Duration, onAlert func([]Alert)) {
	log.Printf("🔍 Alert monitoring started (interval: %v)", interval)
	log.Printf("📊 Thresholds: CPU Temp > %.0f°C, CPU > %.0f%%, RAM > %.0f%%, Disk > %.0f%%",
		checker.Thresholds.CPUTemperature,
		checker.Thresholds.CPUUsage,
		checker.Thresholds.MemoryUsage,
		checker.Thresholds.DiskUsage,
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		alerts, err := checker.CheckSystem()
		if err != nil {
			log.Printf("Error checking system: %v", err)
			continue
		}

		if len(alerts) > 0 {
			log.Printf("⚠️ Found %d alert(s)", len(alerts))
			onAlert(alerts)
		}
	}
}
