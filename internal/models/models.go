package models

import (
	"time"
)

// ActiveQRCode 活码模型 - 主二维码
type ActiveQRCode struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Name          string         `json:"name" gorm:"not null"`
	ShortCode     string         `json:"short_code" gorm:"unique;not null"` // 短码，用于生成活码URL
	QRCodePath    string         `json:"qr_code_path"`                      // 活码二维码图片路径
	Status        int            `json:"status" gorm:"default:1"`           // 1: 启用, 0: 禁用
	SwitchRule    string         `json:"switch_rule" gorm:"default:'time'"` // 切换规则: time, random, weight, geo
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	StaticQRCodes []StaticQRCode `json:"static_qr_codes,omitempty" gorm:"foreignKey:ActiveQRCodeID"`
	ScanRecords   []ScanRecord   `json:"scan_records,omitempty" gorm:"foreignKey:ActiveQRCodeID"`
}

// StaticQRCode 静态二维码模型 - 活码对应的多个静态码
type StaticQRCode struct {
	ID             uint         `json:"id" gorm:"primaryKey"`
	ActiveQRCodeID uint         `json:"active_qr_code_id" gorm:"not null"`
	Name           string       `json:"name" gorm:"not null"`
	TargetURL      string       `json:"target_url" gorm:"not null"` // 实际跳转的目标URL
	Weight         int          `json:"weight" gorm:"default:1"`    // 权重，用于按权重分配
	Status         int          `json:"status" gorm:"default:1"`    // 1: 启用, 0: 禁用
	StartTime      *time.Time   `json:"start_time"`                 // 生效开始时间
	EndTime        *time.Time   `json:"end_time"`                   // 生效结束时间
	AllowedRegions string       `json:"allowed_regions"`            // 允许的地区，JSON格式
	AllowedDevices string       `json:"allowed_devices"`            // 允许的设备类型，JSON格式
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	ActiveQRCode   ActiveQRCode `json:"active_qr_code,omitempty" gorm:"foreignKey:ActiveQRCodeID"`
}

// QRCode 保留原有的简单二维码模型（向后兼容）
type QRCode struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null"`
	OriginalURL string       `json:"original_url" gorm:"not null"`
	QRCodePath  string       `json:"qr_code_path"`
	Status      int          `json:"status" gorm:"default:1"` // 1: 启用, 0: 禁用
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	ScanRecords []ScanRecord `json:"scan_records,omitempty" gorm:"foreignKey:QRCodeID"`
}

// ScanRecord 扫描记录模型
type ScanRecord struct {
	ID             uint          `json:"id" gorm:"primaryKey"`
	QRCodeID       *uint         `json:"qr_code_id"`        // 普通二维码ID（可为空）
	ActiveQRCodeID *uint         `json:"active_qr_code_id"` // 活码ID（可为空）
	StaticQRCodeID *uint         `json:"static_qr_code_id"` // 实际跳转的静态码ID（可为空）
	IPAddress      string        `json:"ip_address"`
	UserAgent      string        `json:"user_agent"`
	ScanTime       time.Time     `json:"scan_time"`
	Location       string        `json:"location"`
	Device         string        `json:"device"`     // 设备类型：mobile, desktop, tablet
	Region         string        `json:"region"`     // 地区信息
	TargetURL      string        `json:"target_url"` // 实际跳转的URL
	QRCode         *QRCode       `json:"qr_code,omitempty" gorm:"foreignKey:QRCodeID"`
	ActiveQRCode   *ActiveQRCode `json:"active_qr_code,omitempty" gorm:"foreignKey:ActiveQRCodeID"`
	StaticQRCode   *StaticQRCode `json:"static_qr_code,omitempty" gorm:"foreignKey:StaticQRCodeID"`
}

// User 用户模型
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'user'"`
	CreatedAt    time.Time `json:"created_at"`
}

// ScanStats 扫描统计
type ScanStats struct {
	TotalScans int64 `json:"total_scans"`
	TodayScans int64 `json:"today_scans"`
	WeekScans  int64 `json:"week_scans"`
	MonthScans int64 `json:"month_scans"`
}

// ActiveQRCodeCreateRequest 创建活码请求
type ActiveQRCodeCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	SwitchRule  string `json:"switch_rule"` // time, random, weight, geo
	Description string `json:"description"`
}

// StaticQRCodeCreateRequest 创建静态码请求
type StaticQRCodeCreateRequest struct {
	ActiveQRCodeID uint       `json:"active_qr_code_id" binding:"required"`
	Name           string     `json:"name" binding:"required"`
	TargetURL      string     `json:"target_url" binding:"required"`
	Weight         int        `json:"weight"`
	Status         int        `json:"status"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	AllowedRegions string     `json:"allowed_regions"`
	AllowedDevices string     `json:"allowed_devices"`
}

// StaticQRCodeUpdateRequest 更新静态码请求
type StaticQRCodeUpdateRequest struct {
	ActiveQRCodeID *uint      `json:"active_qr_code_id"`
	Name           *string    `json:"name"`
	TargetURL      *string    `json:"target_url"`
	Weight         *int       `json:"weight"`
	Status         *int       `json:"status"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	AllowedRegions *string    `json:"allowed_regions"`
	AllowedDevices *string    `json:"allowed_devices"`
}

// ActiveQRCodeUpdateRequest 更新活码请求
type ActiveQRCodeUpdateRequest struct {
	Name        string `json:"name"`
	SwitchRule  string `json:"switch_rule"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

// QRCodeCreateRequest 创建二维码请求
type QRCodeCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	OriginalURL string `json:"original_url" binding:"required"`
	Description string `json:"description"`
}

// QRCodeUpdateRequest 更新二维码请求
type QRCodeUpdateRequest struct {
	Name        string `json:"name"`
	OriginalURL string `json:"original_url"`
	Status      *int   `json:"status"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// APIResponse API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse 分页响应（通用）
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// AppError 应用错误
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}
