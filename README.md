# 🍓 Pi Monitor Telegram Bot

Telegram bot để giám sát Raspberry Pi. Sử dụng lệnh `/pi` để xem thông tin hệ thống.

## ✨ Tính năng

- 🖥️ **CPU**: % sử dụng, nhiệt độ, số cores, tần số
- 💾 **RAM**: Tổng/Đã dùng/Còn trống
- 💿 **Disk**: Dung lượng/Đã dùng/Còn trống  
- 🌐 **Network**: IP, bytes sent/received
- ⏱️ **Uptime**: Thời gian hoạt động

## 📋 Yêu cầu

- Docker & Docker Compose
- Telegram Bot Token (tạo từ [@BotFather](https://t.me/BotFather))

## 🚀 Cài đặt

1. Clone repository:
```bash
git clone <repo-url>
cd pi-monitor
```

2. Tạo file `.env`:
```bash
cp .env.example .env
```

3. Sửa file `.env` và thêm Bot Token:
```
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

4. Chạy với Docker Compose:
```bash
docker-compose up -d
```

## 📱 Sử dụng

- `/start` - Bắt đầu
- `/pi` - Xem thông tin hệ thống
- `/help` - Trợ giúp

## 📸 Demo

```
🍓 Raspberry Pi Status

🖥️ CPU
├ Sử dụng: 15.2%
├ Nhiệt độ: 45.3°C
├ Cores: 4
└ Tần số: 1500 MHz

💾 RAM
├ Tổng: 3.7 GB
├ Đã dùng: 1.2 GB (32.4%)
└ Còn trống: 2.5 GB

💿 Disk
├ Tổng: 29.5 GB
├ Đã dùng: 8.2 GB (27.8%)
└ Còn trống: 21.3 GB

🌐 Network
├ IP: 192.168.1.100
├ Gửi: 156.3 MB
└ Nhận: 1.2 GB

⏱️ Uptime: 5 ngày 12 giờ 30 phút
🕐 Cập nhật: 29/01/2026 00:20:00
```

## 🔧 Development

```bash
# Chạy local (không Docker)
go mod download
go run .

# Build
go build -o pi-monitor .
```

## 📄 License

MIT
