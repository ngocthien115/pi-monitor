package main

import (
	"fmt"
	"log"

	"pi-monitor/config"
	"pi-monitor/handlers"

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

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
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🚫 Bạn không có quyền sử dụng bot này.\n\n🆔 Your User ID: `%d`", userID))
			msg.ParseMode = "Markdown"
			bot.Send(msg)
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
			msg = tgbotapi.NewMessage(chatID, "📖 *Danh sách lệnh:*\n\n/pi - Xem thông tin hệ thống (CPU, RAM, Disk, Network)\n/id - Xem User ID của bạn\n/help - Hiển thị trợ giúp")
			msg.ParseMode = "Markdown"
		default:
			msg = tgbotapi.NewMessage(chatID, "❓ Lệnh không hợp lệ. Sử dụng /help để xem danh sách lệnh.")
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
