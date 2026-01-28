package handlers

import (
	"fmt"

	"pi-monitor/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandlePiCommand(message *tgbotapi.Message) tgbotapi.MessageConfig {
	info, err := services.GetSystemInfo()
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Lỗi khi lấy thông tin hệ thống: %v", err))
		return msg
	}

	text := formatSystemInfo(info)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func formatSystemInfo(info *services.SystemInfo) string {
	return fmt.Sprintf(`🍓 *Raspberry Pi Status*

🖥️ *CPU*
├ Sử dụng: %.1f%%
├ Nhiệt độ: %.1f°C
├ Cores: %d
└ Tần số: %.0f MHz

💾 *RAM*
├ Tổng: %s
├ Đã dùng: %s (%.1f%%)
└ Còn trống: %s

💿 *Disk*
├ Tổng: %s
├ Đã dùng: %s (%.1f%%)
└ Còn trống: %s

🌐 *Network*
├ IP: %s
├ Gửi: %s
└ Nhận: %s

⏱️ *Uptime*: %s
🕐 *Cập nhật*: %s`,
		info.CPU.UsagePercent,
		info.CPU.Temperature,
		info.CPU.Cores,
		info.CPU.Frequency,
		info.Memory.Total,
		info.Memory.Used,
		info.Memory.UsedPercent,
		info.Memory.Available,
		info.Disk.Total,
		info.Disk.Used,
		info.Disk.UsedPercent,
		info.Disk.Free,
		info.Network.IP,
		info.Network.BytesSent,
		info.Network.BytesRecv,
		info.Uptime,
		info.Timestamp,
	)
}
