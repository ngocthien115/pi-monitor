package services

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

// IsPCOnline kiểm tra xem PC có đang bật không bằng cách thử kết nối TCP (port 80 hoặc 445)
// Nếu WOLHost trống, luôn trả về false (không check được)
func IsPCOnline(host string) bool {
	if host == "" {
		return false
	}

	// Thử kết nối đến các port phổ biến: SMB (445), HTTP (80), RDP (3389)
	ports := []string{"445", "80", "3389", "22"}
	for _, port := range ports {
		addr := net.JoinHostPort(host, port)
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}

	// Thử ICMP ping bằng cách dial UDP (không cần quyền root)
	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, "80"), 1*time.Second)
	if err == nil {
		conn.Close()
		// UDP không xác nhận được kết nối, thử cách khác
	}

	return false
}

// SendMagicPacket gửi Wake-on-LAN magic packet đến địa chỉ MAC
// macAddr: địa chỉ MAC dạng "AA:BB:CC:DD:EE:FF" hoặc "AA-BB-CC-DD-EE-FF"
// broadcast: địa chỉ broadcast dạng "255.255.255.255:9" hoặc "192.168.1.255:9"
func SendMagicPacket(macAddr, broadcast string) error {
	// Normalize MAC address - bỏ dấu : hoặc -
	macAddr = strings.ReplaceAll(macAddr, ":", "")
	macAddr = strings.ReplaceAll(macAddr, "-", "")
	macAddr = strings.ToLower(macAddr)

	if len(macAddr) != 12 {
		return fmt.Errorf("địa chỉ MAC không hợp lệ: %s", macAddr)
	}

	// Decode MAC address hex string thành bytes
	macBytes, err := hex.DecodeString(macAddr)
	if err != nil {
		return fmt.Errorf("không thể parse MAC address: %w", err)
	}

	// Xây dựng magic packet:
	// 6 bytes 0xFF + 16 lần lặp MAC address (6 bytes) = 102 bytes tổng
	packet := make([]byte, 102)

	// 6 bytes đầu là 0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	// 16 lần lặp MAC address
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], macBytes)
	}

	// Gửi qua UDP broadcast
	udpAddr, err := net.ResolveUDPAddr("udp", broadcast)
	if err != nil {
		return fmt.Errorf("địa chỉ broadcast không hợp lệ: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("không thể mở kết nối UDP: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("không thể gửi magic packet: %w", err)
	}

	return nil
}
