package main

import (
	"fmt"
	"log"
	"wechat-active-qrcode/internal/database"
	"wechat-active-qrcode/internal/services"
	"wechat-active-qrcode/pkg/qrcode"
)

func main() {
	// 初始化数据库连接
	db, err := database.NewSQLiteConnection("./data/qrcode.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化二维码生成器
	qrGenerator := qrcode.NewGenerator("./data/qrcodes")

	// 初始化服务
	activeQRCodeService := services.NewActiveQRCodeService(db, qrGenerator)

	// 测试获取目标URL
	fmt.Println("Testing GetTargetURL with short code: WC71nyHf")
	targetURL, err := activeQRCodeService.GetTargetURL("WC71nyHf", "Mozilla/5.0", "127.0.0.1", "")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Error type: %T\n", err)

		// 检查是否为QRCodeError类型
		if qrErr, ok := err.(*services.QRCodeError); ok {
			fmt.Printf("QRCodeError - Code: %s, Message: %s\n", qrErr.Code, qrErr.Message)
		}
	} else {
		fmt.Printf("Success! Target URL: %s\n", targetURL)
	}
}
