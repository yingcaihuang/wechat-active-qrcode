package api

import (
	"wechat-active-qrcode/internal/api/handlers"
	"wechat-active-qrcode/internal/api/middleware"
	"wechat-active-qrcode/internal/config"
	"wechat-active-qrcode/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	qrCodeHandler       *handlers.QRCodeHandler
	activeQRCodeHandler *handlers.ActiveQRCodeHandler
	statisticsHandler   *handlers.StatisticsHandler
	authHandler         *handlers.AuthHandler
	authMiddleware      *middleware.AuthMiddleware
	config              *config.Config
}

func NewRouter(
	qrCodeService *services.QRCodeService,
	activeQRCodeService *services.ActiveQRCodeService,
	statisticsService *services.StatisticsService,
	authService *services.AuthService,
	cfg *config.Config,
) *Router {
	return &Router{
		qrCodeHandler:       handlers.NewQRCodeHandler(qrCodeService),
		activeQRCodeHandler: handlers.NewActiveQRCodeHandler(activeQRCodeService),
		statisticsHandler:   handlers.NewStatisticsHandler(statisticsService),
		authHandler:         handlers.NewAuthHandler(authService),
		authMiddleware:      middleware.NewAuthMiddleware(authService),
		config:              cfg,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()

	// CORS配置
	config := cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域名，生产环境应该设置具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12小时
	}
	router.Use(cors.New(config))

	// 静态文件服务 - 管理后台
	router.Static("/web", "./web")
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/admin", "./web/index.html")

	// 为根路径提供静态文件访问（CSS、JS等）
	router.StaticFile("/styles.css", "./web/styles.css")
	router.StaticFile("/app.js", "./web/app.js")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// 添加assets静态文件服务
	router.Static("/assets", "./web/assets")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "WeChat QR Code Management System is running",
		})
	})

	// API路由组
	api := router.Group("/api")
	api.Use(middleware.NoCacheMiddleware()) // 为所有API添加防缓存头
	{
		// 配置端点（公开访问）
		api.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"base_url": r.config.Server.BaseURL,
			})
		})

		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/refresh", r.authHandler.RefreshToken)
			auth.GET("/profile", r.authMiddleware.AuthRequired(), r.authHandler.GetProfile)
			auth.PUT("/password", r.authMiddleware.AuthRequired(), r.authHandler.ChangePassword)
		}

		// 二维码管理路由（需要认证）
		qrCodes := api.Group("/qrcodes")
		qrCodes.Use(r.authMiddleware.AuthRequired())
		{
			qrCodes.GET("", r.qrCodeHandler.ListQRCodes)
			qrCodes.POST("", r.qrCodeHandler.CreateQRCode)
			qrCodes.GET("/:id", r.qrCodeHandler.GetQRCode)
			qrCodes.PUT("/:id", r.qrCodeHandler.UpdateQRCode)
			qrCodes.DELETE("/:id", r.qrCodeHandler.DeleteQRCode)
			qrCodes.GET("/:id/image", r.qrCodeHandler.GetQRCodeImage)
		}

		// 活码管理路由（需要认证）
		activeQRCodes := api.Group("/active-qrcodes")
		activeQRCodes.Use(r.authMiddleware.AuthRequired())
		{
			activeQRCodes.GET("", r.activeQRCodeHandler.ListActiveQRCodes)
			activeQRCodes.POST("", r.activeQRCodeHandler.CreateActiveQRCode)
			activeQRCodes.GET("/:id", r.activeQRCodeHandler.GetActiveQRCode)
			activeQRCodes.PUT("/:id", r.activeQRCodeHandler.UpdateActiveQRCode)
			activeQRCodes.DELETE("/:id", r.activeQRCodeHandler.DeleteActiveQRCode)
			activeQRCodes.GET("/:id/image", r.activeQRCodeHandler.GetActiveQRCodeImage)
			activeQRCodes.GET("/:id/qrcode", r.activeQRCodeHandler.GetActiveQRCodeImage) // 别名
			activeQRCodes.POST("/:id/static-qrcodes", r.activeQRCodeHandler.AddStaticQRCode)
			activeQRCodes.PATCH("/:id/toggle-status", r.activeQRCodeHandler.ToggleActiveQRStatus) // 切换状态
		}

		// 静态码管理路由（需要认证）
		staticQRCodes := api.Group("/static-qrcodes")
		staticQRCodes.Use(r.authMiddleware.AuthRequired())
		{
			staticQRCodes.GET("", r.activeQRCodeHandler.ListStaticQRCodes)
			staticQRCodes.POST("", r.activeQRCodeHandler.CreateStaticQRCode)
			staticQRCodes.GET("/:id", r.activeQRCodeHandler.GetStaticQRCode)
			staticQRCodes.PUT("/:id", r.activeQRCodeHandler.UpdateStaticQRCode)
			staticQRCodes.DELETE("/:id", r.activeQRCodeHandler.DeleteStaticQRCode)
			staticQRCodes.PATCH("/:id/toggle-status", r.activeQRCodeHandler.ToggleStaticQRStatus) // 切换状态
		}

		// 统计相关路由（需要认证）
		statistics := api.Group("/statistics")
		statistics.Use(r.authMiddleware.AuthRequired())
		{
			statistics.GET("", r.statisticsHandler.GetOverviewStats) // 添加根路径
			statistics.GET("/overview", r.statisticsHandler.GetOverviewStats)
			statistics.GET("/trends", r.statisticsHandler.GetTrendData)
			statistics.GET("/top-qrcodes", r.statisticsHandler.GetTopQRCodes)
			statistics.GET("/scan-records", r.statisticsHandler.GetRecentScanRecords)
			statistics.GET("/qrcodes/:id/stats", r.statisticsHandler.GetScanStatistics)
			statistics.GET("/qrcodes/:id/records", r.statisticsHandler.GetScanRecords)
		}

		// 工具类路由（需要认证）
		tools := api.Group("/tools")
		tools.Use(r.authMiddleware.AuthRequired())
		{
			// 二维码解析
			tools.POST("/parse-qrcode", r.activeQRCodeHandler.ParseQRCode)
		}

		// 公开路由（不需要认证）
		public := api.Group("/public")
		{
			// 扫描记录路由（公开访问）
			public.POST("/scan/:id", r.qrCodeHandler.RecordScan)

			// 二维码图片访问（公开）
			public.GET("/qrcodes/:id/image", r.qrCodeHandler.GetQRCodeImage)
			public.GET("/active-qrcodes/:id/image", r.activeQRCodeHandler.GetActiveQRCodeImage)
			public.GET("/active-qrcodes/:id/qrcode", r.activeQRCodeHandler.GetActiveQRCodeImage)
		}
	}

	// 活码重定向路由（独立路径，不在API组下）
	router.GET("/r/:shortCode", r.activeQRCodeHandler.RedirectByShortCode)

	return router
}
