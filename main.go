package main

import (
	"log"
	"os"

	"pi-monitor/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("🤖 Bot authorized on account %s", bot.Self.UserName)

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

		var msg tgbotapi.MessageConfig

		switch update.Message.Command() {
		case "pi":
			msg = handlers.HandlePiCommand(update.Message)
		case "start":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "👋 Xin chào! Sử dụng lệnh /pi để xem thông tin hệ thống Raspberry Pi.")
		case "help":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "📖 *Danh sách lệnh:*\n\n/pi - Xem thông tin hệ thống (CPU, RAM, Disk, Network)\n/help - Hiển thị trợ giúp")
			msg.ParseMode = "Markdown"
		default:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "❓ Lệnh không hợp lệ. Sử dụng /help để xem danh sách lệnh.")
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
