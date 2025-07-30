package qrcode

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Parser 二维码解析器
type Parser struct{}

// NewParser 创建新的二维码解析器
func NewParser() *Parser {
	return &Parser{}
}

// ParseFromFile 从上传的文件解析二维码
func (p *Parser) ParseFromFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 读取文件内容
	defer file.Close()
	
	// 检查文件类型
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("文件类型不支持，请上传图片文件")
	}
	
	// 解码图片
	var img image.Image
	var err error
	
	// 重置文件指针到开始位置
	file.Seek(0, 0)
	
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		img, err = jpeg.Decode(file)
	} else if strings.Contains(contentType, "png") {
		img, err = png.Decode(file)
	} else {
		// 尝试自动检测格式
		file.Seek(0, 0)
		img, _, err = image.Decode(file)
	}
	
	if err != nil {
		return "", fmt.Errorf("图片解码失败: %v", err)
	}
	
	return p.parseFromImage(img)
}

// ParseFromReader 从io.Reader解析二维码
func (p *Parser) ParseFromReader(reader io.Reader) (string, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("图片解码失败: %v", err)
	}
	
	return p.parseFromImage(img)
}

// parseFromImage 从image.Image解析二维码
func (p *Parser) parseFromImage(img image.Image) (string, error) {
	// 转换为gozxing可以处理的格式
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("图片处理失败: %v", err)
	}
	
	// 创建二维码读取器
	qrReader := qrcode.NewQRCodeReader()
	
	// 解析二维码
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("二维码解析失败: %v", err)
	}
	
	text := result.GetText()
	if text == "" {
		return "", fmt.Errorf("二维码内容为空")
	}
	
	return text, nil
}

// ValidateURL 验证解析出的内容是否为有效URL
func (p *Parser) ValidateURL(text string) bool {
	text = strings.TrimSpace(text)
	return strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")
}
