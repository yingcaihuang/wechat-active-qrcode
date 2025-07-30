package services

import (
	"errors"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/pkg/qrcode"
	"wechat-active-qrcode/pkg/utils"

	"gorm.io/gorm"
)

type QRCodeService struct {
	db        *gorm.DB
	generator *qrcode.Generator
}

func NewQRCodeService(db *gorm.DB, generator *qrcode.Generator) *QRCodeService {
	return &QRCodeService{
		db:        db,
		generator: generator,
	}
}

// CreateQRCode 创建二维码
func (s *QRCodeService) CreateQRCode(req *models.QRCodeCreateRequest) (*models.QRCode, error) {
	// 验证URL格式
	if !utils.IsValidURL(req.OriginalURL) {
		return nil, errors.New("invalid URL format")
	}

	// 生成二维码文件名
	filename := s.generator.GenerateFilename("qr")

	// 生成二维码图片
	qrCodePath, err := s.generator.GenerateQRCode(req.OriginalURL, filename)
	if err != nil {
		return nil, err
	}

	// 创建二维码记录
	qrCode := &models.QRCode{
		Name:        req.Name,
		OriginalURL: req.OriginalURL,
		QRCodePath:  qrCodePath,
		Status:      1,
	}

	if err := s.db.Create(qrCode).Error; err != nil {
		// 如果数据库创建失败，删除已生成的二维码文件
		s.generator.DeleteQRCode(filename)
		return nil, err
	}

	return qrCode, nil
}

// GetQRCode 获取二维码详情
func (s *QRCodeService) GetQRCode(id uint) (*models.QRCode, error) {
	var qrCode models.QRCode
	err := s.db.Preload("ScanRecords").First(&qrCode, id).Error
	if err != nil {
		return nil, err
	}
	return &qrCode, nil
}

// UpdateQRCode 更新二维码
func (s *QRCodeService) UpdateQRCode(id uint, req *models.QRCodeUpdateRequest) (*models.QRCode, error) {
	var qrCode models.QRCode
	if err := s.db.First(&qrCode, id).Error; err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		qrCode.Name = req.Name
	}
	if req.OriginalURL != "" {
		if !utils.IsValidURL(req.OriginalURL) {
			return nil, errors.New("invalid URL format")
		}
		qrCode.OriginalURL = req.OriginalURL

		// 重新生成二维码
		filename := s.generator.GenerateFilename("qr")
		qrCodePath, err := s.generator.GenerateQRCode(req.OriginalURL, filename)
		if err != nil {
			return nil, err
		}

		// 删除旧二维码文件
		if qrCode.QRCodePath != "" {
			s.generator.DeleteQRCode(qrCode.QRCodePath)
		}

		qrCode.QRCodePath = qrCodePath
	}
	if req.Status != nil {
		qrCode.Status = *req.Status
	}

	if err := s.db.Save(&qrCode).Error; err != nil {
		return nil, err
	}

	return &qrCode, nil
}

// DeleteQRCode 删除二维码
func (s *QRCodeService) DeleteQRCode(id uint) error {
	var qrCode models.QRCode
	if err := s.db.First(&qrCode, id).Error; err != nil {
		return err
	}

	// 删除二维码文件
	if qrCode.QRCodePath != "" {
		s.generator.DeleteQRCode(qrCode.QRCodePath)
	}

	// 删除数据库记录
	return s.db.Delete(&qrCode).Error
}

// ListQRCodes 获取二维码列表
func (s *QRCodeService) ListQRCodes(page, pageSize int) (*models.PaginationResponse, error) {
	var qrCodes []models.QRCode
	var total int64

	// 获取总数
	s.db.Model(&models.QRCode{}).Count(&total)

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := s.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&qrCodes).Error
	if err != nil {
		return nil, err
	}

	totalPages := utils.CalculateTotalPages(total, pageSize)

	return &models.PaginationResponse{
		Data:       qrCodes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// RecordScan 记录扫描
func (s *QRCodeService) RecordScan(qrCodeID uint, ipAddress, userAgent string) error {
	// 检查二维码是否存在且启用
	var qrCode models.QRCode
	if err := s.db.First(&qrCode, qrCodeID).Error; err != nil {
		return err
	}

	if qrCode.Status != 1 {
		return errors.New("QR code is disabled")
	}

	// 创建扫描记录
	scanRecord := &models.ScanRecord{
		QRCodeID:  &qrCodeID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	return s.db.Create(scanRecord).Error
}

// GetQRCodeImage 获取二维码图片
func (s *QRCodeService) GetQRCodeImage(id uint) ([]byte, error) {
	var qrCode models.QRCode
	if err := s.db.First(&qrCode, id).Error; err != nil {
		return nil, err
	}

	if qrCode.QRCodePath == "" {
		return nil, errors.New("QR code image not found")
	}

	// 读取二维码图片文件
	return s.generator.ReadQRCodeFile(qrCode.QRCodePath)
}
