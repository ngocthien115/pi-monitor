package main

import (
	"fmt"
	"log"
	"time"

	"pi-monitor/config"
	"pi-monitor/handlers"
	"pi-monitor/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg := config.Load()

	if cfg.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("🤖 Bot authorized on account %s", bot.Self.UserName)

	if len(cfg.AllowedUsers) > 0 {
		log.Printf("🔒 Whitelist enabled: %d users allowed", len(cfg.AllowedUsers))
	} else {
		log.Printf("⚠️  Whitelist disabled: all users can use this bot")
	}

	// Start alert monitoring if enabled
	if cfg.AlertEnabled && len(cfg.AllowedUsers) > 0 {
		thresholds := services.AlertThresholds{
			CPUTemperature: cfg.CPUTempThreshold,
			CPUUsage:       cfg.CPUUsageThreshold,
			MemoryUsage:    cfg.MemoryThreshold,
			DiskUsage:      cfg.DiskThreshold,
		}
		checker := services.NewAlertChecker(thresholds)

		go services.StartMonitoring(checker, cfg.AlertInterval, func(alerts []services.Alert) {
			message := services.FormatAlerts(alerts)

			// Gửi alert đến tất cả allowed users
			for _, userID := range cfg.AllowedUsers {
				msg := tgbotapi.NewMessage(userID, message)
				msg.ParseMode = "Markdown"

				if _, err := bot.Send(msg); err != nil {
					log.Printf("❌ Error sending alert to %d: %v", userID, err)
				} else {
					log.Printf("✅ Alert sent to user %d", userID)
				}
			}
		})

		log.Printf("🚨 Alert monitoring enabled (Users: %d, Interval: %v)", len(cfg.AllowedUsers), cfg.AlertInterval)
	} else if cfg.AlertEnabled && len(cfg.AllowedUsers) == 0 {
		log.Printf("⚠️  Alert enabled but ALLOWED_USERS not set - alerts disabled")
	} else {
		log.Printf("ℹ️  Alert monitoring disabled (set ALERT_ENABLED=true to enable)")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Gửi thông báo khởi động đến tất cả allowed users
	if len(cfg.AllowedUsers) > 0 {
		startupMsg := fmt.Sprintf(
			"🟢 *Bot đã khởi động lại!*\n\n🤖 Bot: @%s\n🕐 Thời gian: `%s`\n\n_Sử dụng /help để xem danh sách lệnh._",
			bot.Self.UserName,
			formatTime(),
		)
		for _, userID := range cfg.AllowedUsers {
			msg := tgbotapi.NewMessage(userID, startupMsg)
			msg.ParseMode = "Markdown"
			if _, err := bot.Send(msg); err != nil {
				log.Printf("⚠️ Không thể gửi thông báo khởi động đến %d: %v", userID, err)
			} else {
				log.Printf("✅ Đã gửi thông báo khởi động đến user %d", userID)
			}
		}
	}

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		username := update.Message.From.UserName

		// Check whitelist
		if !cfg.IsUserAllowed(userID) {
			log.Printf("🚫 Unauthorized access attempt from user %d (@%s)", userID, username)
			continue
		}

		var msg tgbotapi.MessageConfig

		switch update.Message.Command() {
		case "pi":
			msg = handlers.HandlePiCommand(update.Message)
		case "id":
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("🆔 Your User ID: `%d`", userID))
			msg.ParseMode = "Markdown"
		case "start":
			msg = tgbotapi.NewMessage(chatID, "👋 Xin chào! Sử dụng lệnh /pi để xem thông tin hệ thống Raspberry Pi.")
		case "help":
			helpText := "📖 *Danh sách lệnh:*\n\n" +
				"/pi - Xem thông tin hệ thống (CPU, RAM, Disk, Network)\n" +
				"/wake - Bật PC qua Wake-on-LAN\n" +
				"/id - Xem User ID của bạn\n" +
				"/alert - Xem trạng thái cảnh báo\n" +
				"/help - Hiển thị trợ giúp"
			msg = tgbotapi.NewMessage(chatID, helpText)
			msg.ParseMode = "Markdown"
		case "alert":
			msg = handleAlertStatus(chatID, cfg)
		case "wake":
			msg = handlers.HandleWakeCommand(update.Message, cfg)
		default:
			msg = tgbotapi.NewMessage(chatID, "❓ Lệnh không hợp lệ. Sử dụng /help để xem danh sách lệnh.")
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

// handleAlertStatus trả về thông tin về trạng thái alert
func handleAlertStatus(chatID int64, cfg *config.Config) tgbotapi.MessageConfig {
	var status string
	if cfg.AlertEnabled && len(cfg.AllowedUsers) > 0 {
		status = fmt.Sprintf(`🚨 *Trạng thái cảnh báo*

✅ *Trạng thái:* Đang hoạt động
⏱️ *Kiểm tra mỗi:* %v
👥 *Gửi đến:* %d người dùng

📊 *Ngưỡng cảnh báo:*
├ 🌡️ Nhiệt độ CPU: > %.0f°C
├ 📈 Sử dụng CPU: > %.0f%%
├ 💾 Sử dụng RAM: > %.0f%%
└ 💿 Sử dụng Disk: > %.0f%%

_Bạn sẽ nhận cảnh báo khi hệ thống vượt ngưỡng_`,
			cfg.AlertInterval,
			len(cfg.AllowedUsers),
			cfg.CPUTempThreshold,
			cfg.CPUUsageThreshold,
			cfg.MemoryThreshold,
			cfg.DiskThreshold,
		)
	} else {
		status = "🚨 *Trạng thái cảnh báo*\n\n❌ *Trạng thái:* Đã tắt\n\n_Đặt ALERT\\_ENABLED=true và ALLOWED\\_USERS để bật_"
	}

	msg := tgbotapi.NewMessage(chatID, status)
	msg.ParseMode = "Markdown"
	return msg
}

// formatTime trả về thời gian hiện tại dạng dễ đọc theo timezone Asia/Ho_Chi_Minh
func formatTime() string {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("02/01/2006 15:04:05")
}
