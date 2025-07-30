package database

import (
	"log"
	"os"
	"path/filepath"
	"wechat-active-qrcode/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSQLiteConnection(dbPath string) (*gorm.DB, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&models.QRCode{},
		&models.ActiveQRCode{},
		&models.StaticQRCode{},
		&models.ScanRecord{},
		&models.User{},
	)

	if err != nil {
		return nil, err
	}

	// 创建默认管理员用户
	createDefaultAdmin(db)

	log.Println("Database initialized successfully")
	return db, nil
}

func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count == 0 {
		// 创建默认管理员用户
		adminUser := models.User{
			Username:     "admin",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:         "admin",
		}

		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("Failed to create default admin user: %v", err)
		} else {
			log.Println("Default admin user created: admin/password")
		}
	}
}
