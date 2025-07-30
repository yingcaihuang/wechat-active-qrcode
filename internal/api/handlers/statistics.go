package handlers

import (
	"net/http"
	"strconv"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/internal/services"
	"wechat-active-qrcode/pkg/utils"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	statisticsService *services.StatisticsService
}

func NewStatisticsHandler(statisticsService *services.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

// GetScanStatistics 获取二维码扫描统计
func (h *StatisticsHandler) GetScanStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	stats, err := h.statisticsService.GetScanStatistics(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Statistics retrieved successfully",
		Data:    stats,
	})
}

// GetOverviewStats 获取总览统计
func (h *StatisticsHandler) GetOverviewStats(c *gin.Context) {
	stats, err := h.statisticsService.GetOverviewStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Overview statistics retrieved successfully",
		Data:    stats,
	})
}

// GetTrendData 获取趋势数据
func (h *StatisticsHandler) GetTrendData(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 30 {
		days = 7
	}

	trendData, err := h.statisticsService.GetTrendData(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trend data retrieved successfully",
		Data:    trendData,
	})
}

// GetTopQRCodes 获取热门二维码
func (h *StatisticsHandler) GetTopQRCodes(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	topQRCodes, err := h.statisticsService.GetTopQRCodes(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Top QR codes retrieved successfully",
		Data:    topQRCodes,
	})
}

// GetScanRecords 获取扫描记录
func (h *StatisticsHandler) GetScanRecords(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, pageSize := utils.ParsePagination(pageStr, pageSizeStr)

	records, err := h.statisticsService.GetScanRecords(uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Scan records retrieved successfully",
		Data:    records,
	})
}

// GetRecentScanRecords 获取最近的扫描记录
func (h *StatisticsHandler) GetRecentScanRecords(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	records, err := h.statisticsService.GetRecentScanRecords(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Recent scan records retrieved successfully",
		Data:    records,
	})
}

// GetDeviceStats 获取设备类型统计
func (h *StatisticsHandler) GetDeviceStats(c *gin.Context) {
	deviceStats, err := h.statisticsService.GetDeviceStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Device statistics retrieved successfully",
		Data:    deviceStats,
	})
}

// GetRegionStats 获取地区统计
func (h *StatisticsHandler) GetRegionStats(c *gin.Context) {
	regionStats, err := h.statisticsService.GetRegionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Region statistics retrieved successfully",
		Data:    regionStats,
	})
}
