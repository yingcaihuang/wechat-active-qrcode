package services

import (
	"time"
	"wechat-active-qrcode/internal/models"

	"gorm.io/gorm"
)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{
		db: db,
	}
}

// GetScanStatistics 获取二维码扫描统计
func (s *StatisticsService) GetScanStatistics(qrCodeID uint) (*models.ScanStats, error) {
	var stats models.ScanStats

	// 获取总扫描次数
	s.db.Model(&models.ScanRecord{}).Where("qr_code_id = ?", qrCodeID).Count(&stats.TotalScans)

	// 获取今日扫描次数
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.ScanRecord{}).Where("qr_code_id = ? AND DATE(scan_time) = ?", qrCodeID, today).Count(&stats.TodayScans)

	// 获取本周扫描次数
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	s.db.Model(&models.ScanRecord{}).Where("qr_code_id = ? AND scan_time >= ?", qrCodeID, weekStart).Count(&stats.WeekScans)

	// 获取本月扫描次数
	monthStart := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.ScanRecord{}).Where("qr_code_id = ? AND scan_time >= ?", qrCodeID, monthStart).Count(&stats.MonthScans)

	return &stats, nil
}

// GetOverviewStats 获取总览统计
func (s *StatisticsService) GetOverviewStats() (map[string]interface{}, error) {
	var result map[string]interface{}

	// 总二维码数量
	var totalQRCodes int64
	s.db.Model(&models.QRCode{}).Count(&totalQRCodes)

	// 总扫描次数
	var totalScans int64
	s.db.Model(&models.ScanRecord{}).Count(&totalScans)

	// 今日新增二维码
	var todayNewQRCodes int64
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.QRCode{}).Where("DATE(created_at) = ?", today).Count(&todayNewQRCodes)

	// 今日扫描次数
	var todayScans int64
	s.db.Model(&models.ScanRecord{}).Where("DATE(scan_time) = ?", today).Count(&todayScans)

	// 活跃二维码数量（有扫描记录的）
	var activeQRCodes int64
	s.db.Model(&models.QRCode{}).Joins("JOIN scan_records ON qr_codes.id = scan_records.qr_code_id").Distinct().Count(&activeQRCodes)

	result = map[string]interface{}{
		"total_qr_codes":     totalQRCodes,
		"total_scans":        totalScans,
		"today_new_qr_codes": todayNewQRCodes,
		"today_scans":        todayScans,
		"active_qr_codes":    activeQRCodes,
	}

	return result, nil
}

// GetTrendData 获取趋势数据
func (s *StatisticsService) GetTrendData(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// 当日扫描次数
		var dailyScans int64
		s.db.Model(&models.ScanRecord{}).Where("DATE(scan_time) = ?", dateStr).Count(&dailyScans)

		// 当日新增二维码
		var dailyNewQRCodes int64
		s.db.Model(&models.QRCode{}).Where("DATE(created_at) = ?", dateStr).Count(&dailyNewQRCodes)

		result = append(result, map[string]interface{}{
			"date":         dateStr,
			"scans":        dailyScans,
			"new_qr_codes": dailyNewQRCodes,
		})
	}

	return result, nil
}

// GetTopQRCodes 获取热门二维码
func (s *StatisticsService) GetTopQRCodes(limit int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	rows, err := s.db.Model(&models.QRCode{}).
		Select("qr_codes.id, qr_codes.name, COUNT(scan_records.id) as scan_count").
		Joins("LEFT JOIN scan_records ON qr_codes.id = scan_records.qr_code_id").
		Group("qr_codes.id").
		Order("scan_count DESC").
		Limit(limit).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var name string
		var scanCount int64

		if err := rows.Scan(&id, &name, &scanCount); err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"name":       name,
			"scan_count": scanCount,
		})
	}

	return result, nil
}

// GetScanRecords 获取扫描记录
func (s *StatisticsService) GetScanRecords(qrCodeID uint, page, pageSize int) (*models.PaginationResponse, error) {
	var records []models.ScanRecord
	var total int64

	query := s.db.Model(&models.ScanRecord{}).Where("qr_code_id = ?", qrCodeID)

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("scan_time DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &models.PaginationResponse{
		Data:       records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetRecentScanRecords 获取最近的扫描记录
func (s *StatisticsService) GetRecentScanRecords(limit int) ([]models.ScanRecord, error) {
	var records []models.ScanRecord

	err := s.db.Preload("QRCode").Preload("ActiveQRCode").
		Order("scan_time DESC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}
