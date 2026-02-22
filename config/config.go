package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BotToken     string
	AllowedUsers []int64

	// Alert settings
	AlertEnabled  bool
	AlertInterval time.Duration // Khoảng thời gian kiểm tra

	// Alert thresholds
	CPUTempThreshold  float64
	CPUUsageThreshold float64
	MemoryThreshold   float64
	DiskThreshold     float64

	// Wake-on-LAN settings
	WOLMACAddress string // MAC address của PC (vd: AA:BB:CC:DD:EE:FF)
	WOLBroadcast  string // Broadcast address (vd: 192.168.1.255:9)
	WOLHost       string // IP/hostname của PC để kiểm tra xem có đang bật không
}

func Load() *Config {
	cfg := &Config{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),

		// Default alert settings
		AlertEnabled:  os.Getenv("ALERT_ENABLED") == "true",
		AlertInterval: 30 * time.Second, // Default: check every 30 seconds

		// Default thresholds
		CPUTempThreshold:  70.0,
		CPUUsageThreshold: 90.0,
		MemoryThreshold:   85.0,
		DiskThreshold:     90.0,

		// Wake-on-LAN
		WOLMACAddress: os.Getenv("WOL_MAC_ADDRESS"),
		WOLBroadcast:  getEnvOrDefault("WOL_BROADCAST", "255.255.255.255:9"),
		WOLHost:       os.Getenv("WOL_HOST"),
	}

	// Parse Alert Interval (in seconds)
	if intervalStr := os.Getenv("ALERT_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			cfg.AlertInterval = time.Duration(interval) * time.Second
		}
	}

	// Parse custom thresholds
	if temp := os.Getenv("ALERT_CPU_TEMP"); temp != "" {
		if v, err := strconv.ParseFloat(temp, 64); err == nil {
			cfg.CPUTempThreshold = v
		}
	}
	if usage := os.Getenv("ALERT_CPU_USAGE"); usage != "" {
		if v, err := strconv.ParseFloat(usage, 64); err == nil {
			cfg.CPUUsageThreshold = v
		}
	}
	if mem := os.Getenv("ALERT_MEMORY"); mem != "" {
		if v, err := strconv.ParseFloat(mem, 64); err == nil {
			cfg.MemoryThreshold = v
		}
	}
	if disk := os.Getenv("ALERT_DISK"); disk != "" {
		if v, err := strconv.ParseFloat(disk, 64); err == nil {
			cfg.DiskThreshold = v
		}
	}

	// Parse allowed users from comma-separated string
	// Example: ALLOWED_USERS=123456789,987654321
	allowedStr := os.Getenv("ALLOWED_USERS")
	if allowedStr != "" {
		for _, idStr := range strings.Split(allowedStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				cfg.AllowedUsers = append(cfg.AllowedUsers, id)
			}
		}
	}

	return cfg
}

// getEnvOrDefault trả về giá trị env hoặc giá trị mặc định nếu env không được set
func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// IsUserAllowed checks if a user ID is in the whitelist
// If whitelist is empty, all users are allowed
func (c *Config) IsUserAllowed(userID int64) bool {
	// Nếu không set whitelist, cho phép tất cả
	if len(c.AllowedUsers) == 0 {
		return true
	}

	for _, id := range c.AllowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}
