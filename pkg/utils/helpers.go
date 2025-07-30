package utils

import (
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HashPassword 密码加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ParsePagination 解析分页参数
func ParsePagination(page, pageSize string) (int, int) {
	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)
	
	if p <= 0 {
		p = 1
	}
	if ps <= 0 {
		ps = 10
	}
	if ps > 100 {
		ps = 100
	}
	
	return p, ps
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetClientIP 获取客户端IP
func GetClientIP(forwardedFor, realIP, remoteAddr string) string {
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	if realIP != "" {
		return realIP
	}
	
	if remoteAddr != "" {
		parts := strings.Split(remoteAddr, ":")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	
	return "unknown"
}

// IsValidURL 验证URL格式
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
} 