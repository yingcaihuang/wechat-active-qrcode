package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wechat-active-qrcode/internal/api"
	"wechat-active-qrcode/internal/auth"
	"wechat-active-qrcode/internal/config"
	"wechat-active-qrcode/internal/database"
	"wechat-active-qrcode/internal/services"
	"wechat-active-qrcode/pkg/qrcode"

	"github.com/gin-gonic/gin"
)

func main() {
	// 打印启动横幅
	printBanner()

	// 加载配置
	log.Println("Loading configuration...")
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded: port=%s, mode=%s", cfg.Server.Port, cfg.Server.Mode)

	// 设置Gin模式
	log.Println("Setting Gin mode...")
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	log.Println("Initializing database...")
	db, err := database.NewSQLiteConnection(cfg.Database.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database initialized successfully")

	// 初始化二维码生成器
	log.Println("Initializing QR code generator...")
	qrGenerator := qrcode.NewGenerator("./data/qrcodes")
	log.Println("QR code generator initialized")

	// 初始化JWT服务
	log.Println("Initializing JWT service...")
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expire)
	log.Println("JWT service initialized")

	// 初始化服务
	log.Println("Initializing services...")
	qrCodeService := services.NewQRCodeService(db, qrGenerator)
	activeQRCodeService := services.NewActiveQRCodeService(db, qrGenerator)
	statisticsService := services.NewStatisticsService(db)
	authService := services.NewAuthService(db, jwtService)
	log.Println("Services initialized")

	// 初始化路由
	log.Println("Setting up routes...")
	router := api.NewRouter(qrCodeService, activeQRCodeService, statisticsService, authService, cfg)
	app := router.SetupRoutes()
	log.Println("Routes configured")

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: app,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func printBanner() {
	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║                    WeChat QR Code Management System         ║
║                                                              ║
║  Version: 1.0.0                                             ║
║  Author:  AI Assistant                                      ║
║  License: MIT                                               ║
╚══════════════════════════════════════════════════════════════╝
`)
}
