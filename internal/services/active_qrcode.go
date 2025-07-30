package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/pkg/qrcode"

	"gorm.io/gorm"
)

// QRCodeError 二维码错误类型
type QRCodeError struct {
	Code    string // 错误代码：NOT_FOUND, DISABLED, NO_STATIC_QR, NO_MATCHING_QR
	Message string // 错误消息
}

func (e *QRCodeError) Error() string {
	return e.Message
}

type ActiveQRCodeService struct {
	db          *gorm.DB
	qrGenerator *qrcode.Generator
}

func NewActiveQRCodeService(db *gorm.DB, qrGenerator *qrcode.Generator) *ActiveQRCodeService {
	return &ActiveQRCodeService{
		db:          db,
		qrGenerator: qrGenerator,
	}
}

// GetDB 获取数据库连接（用于处理器直接访问）
func (s *ActiveQRCodeService) GetDB() *gorm.DB {
	return s.db
}

// CreateActiveQRCode 创建活码
func (s *ActiveQRCodeService) CreateActiveQRCode(req *models.ActiveQRCodeCreateRequest) (*models.ActiveQRCode, error) {
	// 生成短码
	shortCode, err := s.generateShortCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %v", err)
	}

	activeQR := &models.ActiveQRCode{
		Name:        req.Name,
		ShortCode:   shortCode,
		SwitchRule:  req.SwitchRule,
		Description: req.Description,
		Status:      1,
	}

	// 保存到数据库
	if err := s.db.Create(activeQR).Error; err != nil {
		return nil, fmt.Errorf("failed to create active QR code: %v", err)
	}

	// 生成活码二维码图片（指向中转页面）
	redirectURL := fmt.Sprintf("http://localhost:8083/r/%s", shortCode)
	qrPath, err := s.qrGenerator.GenerateQRCode(redirectURL, fmt.Sprintf("active_%d.png", activeQR.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %v", err)
	}

	// 更新二维码路径
	activeQR.QRCodePath = qrPath
	if err := s.db.Save(activeQR).Error; err != nil {
		return nil, fmt.Errorf("failed to update QR code path: %v", err)
	}

	return activeQR, nil
}

// AddStaticQRCode 为活码添加静态二维码
func (s *ActiveQRCodeService) AddStaticQRCode(activeQRCodeID uint, req *models.StaticQRCodeCreateRequest) (*models.StaticQRCode, error) {
	// 检查活码是否存在
	var activeQR models.ActiveQRCode
	if err := s.db.First(&activeQR, activeQRCodeID).Error; err != nil {
		return nil, fmt.Errorf("active QR code not found: %v", err)
	}

	// 转换允许的地区和设备为JSON字符串
	allowedRegionsJSON, _ := json.Marshal(req.AllowedRegions)
	allowedDevicesJSON, _ := json.Marshal(req.AllowedDevices)

	staticQR := &models.StaticQRCode{
		ActiveQRCodeID: activeQRCodeID,
		Name:           req.Name,
		TargetURL:      req.TargetURL,
		Weight:         req.Weight,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		AllowedRegions: string(allowedRegionsJSON),
		AllowedDevices: string(allowedDevicesJSON),
		Status:         1,
	}

	if staticQR.Weight <= 0 {
		staticQR.Weight = 1
	}

	if err := s.db.Create(staticQR).Error; err != nil {
		return nil, fmt.Errorf("failed to create static QR code: %v", err)
	}

	return staticQR, nil
}

// GetTargetURL 根据活码短码和扫描环境获取目标URL
func (s *ActiveQRCodeService) GetTargetURL(shortCode, userAgent, ipAddress, region string) (string, error) {
	// 先查找活码（不考虑状态）
	var activeQR models.ActiveQRCode
	err := s.db.Where("short_code = ?", shortCode).
		Preload("StaticQRCodes").
		First(&activeQR).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", &QRCodeError{
				Code:    "NOT_FOUND",
				Message: "二维码不存在",
			}
		}
		return "", &QRCodeError{
			Code:    "NOT_FOUND",
			Message: "查询失败",
		}
	}

	// 调试信息：输出找到的活码信息
	fmt.Printf("[DEBUG] Found activeQR: ID=%d, Name=%s, Status=%d, StaticQRCodes count=%d\n",
		activeQR.ID, activeQR.Name, activeQR.Status, len(activeQR.StaticQRCodes))

	// 检查活码是否被禁用
	if activeQR.Status != 1 {
		return "", &QRCodeError{
			Code:    "DISABLED",
			Message: "二维码已被禁用",
		}
	}

	// 筛选启用的静态码
	var enabledStaticQRs []models.StaticQRCode
	for _, sqr := range activeQR.StaticQRCodes {
		fmt.Printf("[DEBUG] StaticQR: ID=%d, Name=%s, Status=%d, TargetURL=%s\n",
			sqr.ID, sqr.Name, sqr.Status, sqr.TargetURL)
		if sqr.Status == 1 {
			enabledStaticQRs = append(enabledStaticQRs, sqr)
		}
	}

	fmt.Printf("[DEBUG] Enabled StaticQRs count: %d\n", len(enabledStaticQRs))

	if len(enabledStaticQRs) == 0 {
		return "", &QRCodeError{
			Code:    "NO_STATIC_QR",
			Message: "暂无可用的目标链接",
		}
	}

	// 筛选可用的静态码
	availableQRs := s.filterAvailableQRCodes(enabledStaticQRs, userAgent, region)
	fmt.Printf("[DEBUG] Available QRs count after filtering: %d\n", len(availableQRs))

	if len(availableQRs) == 0 {
		return "", &QRCodeError{
			Code:    "NO_MATCHING_QR",
			Message: "当前环境无匹配的目标链接",
		}
	}

	// 根据切换规则选择目标静态码
	var selectedQR *models.StaticQRCode
	switch activeQR.SwitchRule {
	case "random":
		selectedQR = s.selectRandomQR(availableQRs)
	case "weight":
		selectedQR = s.selectWeightedQR(availableQRs)
	case "time":
		selectedQR = s.selectTimeBasedQR(availableQRs)
	default:
		selectedQR = s.selectWeightedQR(availableQRs) // 默认按权重
	}

	if selectedQR == nil {
		return "", &QRCodeError{
			Code:    "NO_MATCHING_QR",
			Message: "无法选择目标链接",
		}
	}

	// 记录扫描
	go s.recordScan(&activeQR, selectedQR, userAgent, ipAddress, region)

	return selectedQR.TargetURL, nil
}

// filterAvailableQRCodes 筛选可用的静态二维码
func (s *ActiveQRCodeService) filterAvailableQRCodes(staticQRs []models.StaticQRCode, userAgent, region string) []models.StaticQRCode {
	var available []models.StaticQRCode
	now := time.Now()
	device := s.detectDevice(userAgent)

	fmt.Printf("[DEBUG] Filter params: userAgent=%s, region=%s, device=%s, now=%v\n", userAgent, region, device, now)

	for _, qr := range staticQRs {
		fmt.Printf("[DEBUG] Filtering StaticQR ID=%d, Name=%s\n", qr.ID, qr.Name)
		fmt.Printf("[DEBUG] - StartTime: %v, EndTime: %v\n", qr.StartTime, qr.EndTime)
		fmt.Printf("[DEBUG] - AllowedRegions: %s\n", qr.AllowedRegions)
		fmt.Printf("[DEBUG] - AllowedDevices: %s\n", qr.AllowedDevices)

		// 检查时间范围
		if qr.StartTime != nil && now.Before(*qr.StartTime) {
			fmt.Printf("[DEBUG] - REJECTED: StartTime check failed\n")
			continue
		}
		if qr.EndTime != nil && now.After(*qr.EndTime) {
			fmt.Printf("[DEBUG] - REJECTED: EndTime check failed\n")
			continue
		}

		// 检查地区限制
		if qr.AllowedRegions != "" && qr.AllowedRegions != "null" {
			var allowedRegions []string
			if err := json.Unmarshal([]byte(qr.AllowedRegions), &allowedRegions); err == nil {
				fmt.Printf("[DEBUG] - Parsed AllowedRegions: %v\n", allowedRegions)
				if len(allowedRegions) > 0 && !contains(allowedRegions, region) {
					fmt.Printf("[DEBUG] - REJECTED: Region check failed\n")
					continue
				}
			}
		}

		// 检查设备限制
		if qr.AllowedDevices != "" && qr.AllowedDevices != "null" {
			var allowedDevices []string
			if err := json.Unmarshal([]byte(qr.AllowedDevices), &allowedDevices); err == nil {
				fmt.Printf("[DEBUG] - Parsed AllowedDevices: %v\n", allowedDevices)
				if len(allowedDevices) > 0 && !contains(allowedDevices, device) {
					fmt.Printf("[DEBUG] - REJECTED: Device check failed\n")
					continue
				}
			}
		}

		fmt.Printf("[DEBUG] - ACCEPTED: StaticQR ID=%d passed all filters\n", qr.ID)
		available = append(available, qr)
	}

	return available
}

// selectRandomQR 随机选择
func (s *ActiveQRCodeService) selectRandomQR(qrs []models.StaticQRCode) *models.StaticQRCode {
	if len(qrs) == 0 {
		return nil
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(qrs))))
	return &qrs[n.Int64()]
}

// selectWeightedQR 按权重选择
func (s *ActiveQRCodeService) selectWeightedQR(qrs []models.StaticQRCode) *models.StaticQRCode {
	if len(qrs) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	for _, qr := range qrs {
		totalWeight += qr.Weight
	}

	if totalWeight == 0 {
		return s.selectRandomQR(qrs)
	}

	// 生成随机数
	randNum, _ := rand.Int(rand.Reader, big.NewInt(int64(totalWeight)))
	target := int(randNum.Int64())

	// 根据权重选择
	currentWeight := 0
	for i, qr := range qrs {
		currentWeight += qr.Weight
		if target < currentWeight {
			return &qrs[i]
		}
	}

	return &qrs[0]
}

// selectTimeBasedQR 基于时间选择（轮询）
func (s *ActiveQRCodeService) selectTimeBasedQR(qrs []models.StaticQRCode) *models.StaticQRCode {
	if len(qrs) == 0 {
		return nil
	}

	// 基于当前小时进行轮询
	hour := time.Now().Hour()
	index := hour % len(qrs)
	return &qrs[index]
}

// detectDevice 检测设备类型
func (s *ActiveQRCodeService) detectDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

// recordScan 记录扫描
func (s *ActiveQRCodeService) recordScan(activeQR *models.ActiveQRCode, selectedQR *models.StaticQRCode, userAgent, ipAddress, region string) {
	device := s.detectDevice(userAgent)

	scanRecord := &models.ScanRecord{
		ActiveQRCodeID: &activeQR.ID,
		StaticQRCodeID: &selectedQR.ID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ScanTime:       time.Now(),
		Region:         region,
		Device:         device,
		TargetURL:      selectedQR.TargetURL,
	}

	s.db.Create(scanRecord)
}

// generateShortCode 生成短码
func (s *ActiveQRCodeService) generateShortCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	for attempts := 0; attempts < 10; attempts++ {
		shortCode := make([]byte, length)
		for i := range shortCode {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			shortCode[i] = charset[n.Int64()]
		}

		// 检查是否已存在
		var count int64
		s.db.Model(&models.ActiveQRCode{}).Where("short_code = ?", string(shortCode)).Count(&count)
		if count == 0 {
			return string(shortCode), nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after 10 attempts")
}

// ListActiveQRCodes 获取活码列表
func (s *ActiveQRCodeService) ListActiveQRCodes(page, pageSize int) (*models.PaginationResponse, error) {
	var activeQRs []models.ActiveQRCode
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := s.db.Model(&models.ActiveQRCode{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count active QR codes: %v", err)
	}

	// 获取数据
	if err := s.db.Preload("StaticQRCodes").
		Offset(offset).Limit(pageSize).
		Find(&activeQRs).Error; err != nil {
		return nil, fmt.Errorf("failed to get active QR codes: %v", err)
	}

	totalPages := int(total+int64(pageSize)-1) / pageSize

	return &models.PaginationResponse{
		Data:       activeQRs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetActiveQRCode 获取活码详情
func (s *ActiveQRCodeService) GetActiveQRCode(id uint) (*models.ActiveQRCode, error) {
	var activeQR models.ActiveQRCode
	if err := s.db.Preload("StaticQRCodes").First(&activeQR, id).Error; err != nil {
		return nil, fmt.Errorf("active QR code not found: %v", err)
	}
	return &activeQR, nil
}

// UpdateActiveQRCode 更新活码
func (s *ActiveQRCodeService) UpdateActiveQRCode(id uint, req *models.ActiveQRCodeCreateRequest) (*models.ActiveQRCode, error) {
	var activeQR models.ActiveQRCode
	if err := s.db.First(&activeQR, id).Error; err != nil {
		return nil, fmt.Errorf("active QR code not found: %v", err)
	}

	// 更新字段
	activeQR.Name = req.Name
	activeQR.SwitchRule = req.SwitchRule
	activeQR.Description = req.Description

	if err := s.db.Save(&activeQR).Error; err != nil {
		return nil, fmt.Errorf("failed to update active QR code: %v", err)
	}

	// 重新加载带关联数据的记录
	return s.GetActiveQRCode(id)
}

// DeleteActiveQRCode 删除活码
func (s *ActiveQRCodeService) DeleteActiveQRCode(id uint) error {
	// 检查活码是否存在
	var activeQR models.ActiveQRCode
	if err := s.db.First(&activeQR, id).Error; err != nil {
		return fmt.Errorf("active QR code not found: %v", err)
	}

	// 开始事务
	tx := s.db.Begin()

	// 删除关联的静态码
	if err := tx.Where("active_qr_code_id = ?", id).Delete(&models.StaticQRCode{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete static QR codes: %v", err)
	}

	// 删除活码
	if err := tx.Delete(&activeQR).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete active QR code: %v", err)
	}

	return tx.Commit().Error
}

// GetActiveQRCodeImage 获取活码二维码图片
func (s *ActiveQRCodeService) GetActiveQRCodeImage(id uint) ([]byte, error) {
	var activeQR models.ActiveQRCode
	if err := s.db.First(&activeQR, id).Error; err != nil {
		return nil, fmt.Errorf("active QR code not found: %v", err)
	}

	// 重新生成二维码图片
	redirectURL := fmt.Sprintf("http://localhost:8083/r/%s", activeQR.ShortCode)

	// 如果有现有的文件路径，尝试读取
	if activeQR.QRCodePath != "" {
		imageData, err := s.qrGenerator.ReadQRCodeFile(activeQR.QRCodePath)
		if err == nil {
			return imageData, nil
		}
	}

	// 如果文件不存在，重新生成
	base64Data, err := s.qrGenerator.GenerateQRCodeBase64(redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %v", err)
	}

	// 解码base64为字节数组
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QR code image: %v", err)
	}

	return imageData, nil
}

// contains 辅助函数：检查切片是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
