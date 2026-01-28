package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken     string
	AllowedUsers []int64
}

func Load() *Config {
	cfg := &Config{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
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
