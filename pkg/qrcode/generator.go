package qrcode

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/skip2/go-qrcode"
)

type Generator struct {
	StoragePath string
}

func NewGenerator(storagePath string) *Generator {
	// 确保存储目录存在
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create storage directory: %v", err))
	}
	
	return &Generator{
		StoragePath: storagePath,
	}
}

// GenerateQRCode 生成二维码并保存到文件
func (g *Generator) GenerateQRCode(content string, filename string) (string, error) {
	// 生成二维码
	qr, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	
	// 保存到文件
	filePath := filepath.Join(g.StoragePath, filename)
	if err := os.WriteFile(filePath, qr, 0644); err != nil {
		return "", err
	}
	
	return filePath, nil
}

// GenerateQRCodeBase64 生成二维码并返回base64编码
func (g *Generator) GenerateQRCodeBase64(content string) (string, error) {
	qr, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	
	return base64.StdEncoding.EncodeToString(qr), nil
}

// GenerateFilename 生成文件名
func (g *Generator) GenerateFilename(prefix string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d.png", prefix, timestamp)
}

// GetQRCodePath 获取二维码文件路径
func (g *Generator) GetQRCodePath(filename string) string {
	return filepath.Join(g.StoragePath, filename)
}

// DeleteQRCode 删除二维码文件
func (g *Generator) DeleteQRCode(filename string) error {
	filePath := g.GetQRCodePath(filename)
	return os.Remove(filePath)
}

// ReadQRCodeFile 读取二维码文件
func (g *Generator) ReadQRCodeFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
} 