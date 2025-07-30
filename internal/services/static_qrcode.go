package services

import (
	"wechat-active-qrcode/internal/models"

	"gorm.io/gorm"
)

type StaticQRCodeService struct {
	db *gorm.DB
}

func NewStaticQRCodeService(db *gorm.DB) *StaticQRCodeService {
	return &StaticQRCodeService{
		db: db,
	}
}

// ListStaticQRCodes 获取静态码列表
func (s *StaticQRCodeService) ListStaticQRCodes(page, limit int, activeQRCodeID *uint) (*models.PaginatedResponse, error) {
	var staticQRCodes []models.StaticQRCode
	var total int64

	query := s.db.Model(&models.StaticQRCode{}).Preload("ActiveQRCode")

	// 如果指定了活码ID，则过滤
	if activeQRCodeID != nil {
		query = query.Where("active_qr_code_id = ?", *activeQRCodeID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&staticQRCodes).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedResponse{
		Data:  staticQRCodes,
		Total: int(total),
		Page:  page,
		Limit: limit,
	}, nil
}

// CreateStaticQRCode 创建静态码
func (s *StaticQRCodeService) CreateStaticQRCode(req *models.StaticQRCodeCreateRequest) (*models.StaticQRCode, error) {
	// 验证关联的活码是否存在
	var activeQR models.ActiveQRCode
	if err := s.db.First(&activeQR, req.ActiveQRCodeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &models.AppError{
				Code:    "ACTIVE_QR_NOT_FOUND",
				Message: "关联的活码不存在",
			}
		}
		return nil, err
	}

	staticQR := &models.StaticQRCode{
		ActiveQRCodeID: req.ActiveQRCodeID,
		Name:           req.Name,
		TargetURL:      req.TargetURL,
		Weight:         req.Weight,
		Status:         req.Status,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		AllowedRegions: req.AllowedRegions,
		AllowedDevices: req.AllowedDevices,
	}

	if err := s.db.Create(staticQR).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	if err := s.db.Preload("ActiveQRCode").First(staticQR, staticQR.ID).Error; err != nil {
		return nil, err
	}

	return staticQR, nil
}

// GetStaticQRCode 获取静态码详情
func (s *StaticQRCodeService) GetStaticQRCode(id uint) (*models.StaticQRCode, error) {
	var staticQR models.StaticQRCode
	if err := s.db.Preload("ActiveQRCode").First(&staticQR, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &models.AppError{
				Code:    "STATIC_QR_NOT_FOUND",
				Message: "静态码不存在",
			}
		}
		return nil, err
	}

	return &staticQR, nil
}

// UpdateStaticQRCode 更新静态码
func (s *StaticQRCodeService) UpdateStaticQRCode(id uint, req *models.StaticQRCodeUpdateRequest) (*models.StaticQRCode, error) {
	var staticQR models.StaticQRCode
	if err := s.db.First(&staticQR, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &models.AppError{
				Code:    "STATIC_QR_NOT_FOUND",
				Message: "静态码不存在",
			}
		}
		return nil, err
	}

	// 如果更新了活码ID，验证新的活码是否存在
	if req.ActiveQRCodeID != nil && *req.ActiveQRCodeID != staticQR.ActiveQRCodeID {
		var activeQR models.ActiveQRCode
		if err := s.db.First(&activeQR, *req.ActiveQRCodeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, &models.AppError{
					Code:    "ACTIVE_QR_NOT_FOUND",
					Message: "关联的活码不存在",
				}
			}
			return nil, err
		}
		staticQR.ActiveQRCodeID = *req.ActiveQRCodeID
	}

	// 更新字段
	if req.Name != nil {
		staticQR.Name = *req.Name
	}
	if req.TargetURL != nil {
		staticQR.TargetURL = *req.TargetURL
	}
	if req.Weight != nil {
		staticQR.Weight = *req.Weight
	}
	if req.Status != nil {
		staticQR.Status = *req.Status
	}
	if req.StartTime != nil {
		staticQR.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		staticQR.EndTime = req.EndTime
	}
	if req.AllowedRegions != nil {
		staticQR.AllowedRegions = *req.AllowedRegions
	}
	if req.AllowedDevices != nil {
		staticQR.AllowedDevices = *req.AllowedDevices
	}

	if err := s.db.Save(&staticQR).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	if err := s.db.Preload("ActiveQRCode").First(&staticQR, staticQR.ID).Error; err != nil {
		return nil, err
	}

	return &staticQR, nil
}

// DeleteStaticQRCode 删除静态码
func (s *StaticQRCodeService) DeleteStaticQRCode(id uint) error {
	var staticQR models.StaticQRCode
	if err := s.db.First(&staticQR, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.AppError{
				Code:    "STATIC_QR_NOT_FOUND",
				Message: "静态码不存在",
			}
		}
		return err
	}

	return s.db.Delete(&staticQR).Error
}

// GetStaticQRCodesByActiveQRCode 根据活码ID获取所有静态码
func (s *StaticQRCodeService) GetStaticQRCodesByActiveQRCode(activeQRCodeID uint) ([]models.StaticQRCode, error) {
	var staticQRCodes []models.StaticQRCode
	if err := s.db.Where("active_qr_code_id = ?", activeQRCodeID).Find(&staticQRCodes).Error; err != nil {
		return nil, err
	}

	return staticQRCodes, nil
}

// BatchUpdateStatus 批量更新静态码状态
func (s *StaticQRCodeService) BatchUpdateStatus(ids []uint, status int) error {
	return s.db.Model(&models.StaticQRCode{}).Where("id IN ?", ids).Update("status", status).Error
}
