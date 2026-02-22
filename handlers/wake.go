package handlers

import (
	"fmt"
	"pi-monitor/config"
	"pi-monitor/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleWakeCommand xử lý lệnh /wake - gửi magic packet Wake-on-LAN đến PC
func HandleWakeCommand(message *tgbotapi.Message, cfg *config.Config) tgbotapi.MessageConfig {
	chatID := message.Chat.ID

	// Kiểm tra cấu hình WOL
	if cfg.WOLMACAddress == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ *Chưa cấu hình Wake-on-LAN*\n\nVui lòng thiết lập biến môi trường:\n`WOL_MAC_ADDRESS=AA:BB:CC:DD:EE:FF`\n`WOL_HOST=192.168.1.100` _(tuỳ chọn, để kiểm tra trạng thái)_")
		msg.ParseMode = "Markdown"
		return msg
	}

	// Kiểm tra xem PC có đang bật không
	if cfg.WOLHost != "" && services.IsPCOnline(cfg.WOLHost) {
		text := fmt.Sprintf(
			"✅ *PC đã đang bật!*\n\n🖥️ Host: `%s`\n📡 MAC: `%s`\n\n_Không cần gửi magic packet._",
			cfg.WOLHost,
			cfg.WOLMACAddress,
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		return msg
	}

	// PC chưa bật (hoặc không thể kiểm tra) -> gửi magic packet
	err := services.SendMagicPacket(cfg.WOLMACAddress, cfg.WOLBroadcast)
	if err != nil {
		text := fmt.Sprintf("❌ *Gửi magic packet thất bại!*\n\n`%v`", err)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		return msg
	}

	var text string
	if cfg.WOLHost != "" {
		text = fmt.Sprintf(
			"🚀 *Đã gửi lệnh khởi động PC thành công!*\n\n🖥️ Host: `%s`\n📡 MAC: `%s`\n📦 Broadcast: `%s`\n\n⏳ _PC sẽ khởi động trong vài giây..._",
			cfg.WOLHost,
			cfg.WOLMACAddress,
			cfg.WOLBroadcast,
		)
	} else {
		text = fmt.Sprintf(
			"🚀 *Đã gửi magic packet Wake-on-LAN thành công!*\n\n📡 MAC: `%s`\n📦 Broadcast: `%s`\n\n⏳ _PC sẽ khởi động trong vài giây..._",
			cfg.WOLMACAddress,
			cfg.WOLBroadcast,
		)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}
