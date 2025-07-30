package handlers

import (
	"net/http"
	"strconv"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/internal/services"
	"wechat-active-qrcode/pkg/utils"

	"github.com/gin-gonic/gin"
)

type QRCodeHandler struct {
	qrCodeService *services.QRCodeService
}

func NewQRCodeHandler(qrCodeService *services.QRCodeService) *QRCodeHandler {
	return &QRCodeHandler{
		qrCodeService: qrCodeService,
	}
}

// CreateQRCode 创建二维码
func (h *QRCodeHandler) CreateQRCode(c *gin.Context) {
	var req models.QRCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	qrCode, err := h.qrCodeService.CreateQRCode(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "QR code created successfully",
		Data:    qrCode,
	})
}

// GetQRCode 获取二维码详情
func (h *QRCodeHandler) GetQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	qrCode, err := h.qrCodeService.GetQRCode(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "QR code not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "QR code retrieved successfully",
		Data:    qrCode,
	})
}

// UpdateQRCode 更新二维码
func (h *QRCodeHandler) UpdateQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	var req models.QRCodeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	qrCode, err := h.qrCodeService.UpdateQRCode(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "QR code updated successfully",
		Data:    qrCode,
	})
}

// DeleteQRCode 删除二维码
func (h *QRCodeHandler) DeleteQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	err = h.qrCodeService.DeleteQRCode(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "QR code deleted successfully",
	})
}

// ListQRCodes 获取二维码列表
func (h *QRCodeHandler) ListQRCodes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, pageSize := utils.ParsePagination(pageStr, pageSizeStr)

	result, err := h.qrCodeService.ListQRCodes(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "QR codes retrieved successfully",
		Data:    result,
	})
}

// GetQRCodeImage 获取二维码图片
func (h *QRCodeHandler) GetQRCodeImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	imageData, err := h.qrCodeService.GetQRCodeImage(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "QR code image not found",
		})
		return
	}

	// 添加防缓存响应头
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "image/png", imageData)
}

// RecordScan 记录扫描（公开接口）
func (h *QRCodeHandler) RecordScan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid QR code ID",
		})
		return
	}

	// 获取客户端IP和User-Agent
	ipAddress := utils.GetClientIP(
		c.GetHeader("X-Forwarded-For"),
		c.GetHeader("X-Real-IP"),
		c.ClientIP(),
	)
	userAgent := c.GetHeader("User-Agent")

	err = h.qrCodeService.RecordScan(uint(id), ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Scan recorded successfully",
	})
}
