package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/internal/services"
	"wechat-active-qrcode/pkg/qrcode"

	"github.com/gin-gonic/gin"
)

type ActiveQRCodeHandler struct {
	activeQRCodeService *services.ActiveQRCodeService
}

func NewActiveQRCodeHandler(activeQRCodeService *services.ActiveQRCodeService) *ActiveQRCodeHandler {
	return &ActiveQRCodeHandler{
		activeQRCodeService: activeQRCodeService,
	}
}

// ListActiveQRCodes 获取活码列表
func (h *ActiveQRCodeHandler) ListActiveQRCodes(c *gin.Context) {
	page := 1
	pageSize := 10

	// 解析分页参数
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	result, err := h.activeQRCodeService.ListActiveQRCodes(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// CreateActiveQRCode 创建活码
func (h *ActiveQRCodeHandler) CreateActiveQRCode(c *gin.Context) {
	var req models.ActiveQRCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	activeQRCode, err := h.activeQRCodeService.CreateActiveQRCode(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Active QR code created successfully",
		Data:    activeQRCode,
	})
}

// GetActiveQRCode 获取单个活码
func (h *ActiveQRCodeHandler) GetActiveQRCode(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	activeQRCode, err := h.activeQRCodeService.GetActiveQRCode(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Active QR code not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    activeQRCode,
	})
}

// UpdateActiveQRCode 更新活码
func (h *ActiveQRCodeHandler) UpdateActiveQRCode(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	var req models.ActiveQRCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	activeQRCode, err := h.activeQRCodeService.UpdateActiveQRCode(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Active QR code updated successfully",
		Data:    activeQRCode,
	})
}

// DeleteActiveQRCode 删除活码
func (h *ActiveQRCodeHandler) DeleteActiveQRCode(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	err = h.activeQRCodeService.DeleteActiveQRCode(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Active QR code deleted successfully",
	})
}

// GetActiveQRCodeImage 获取活码二维码图片
func (h *ActiveQRCodeHandler) GetActiveQRCodeImage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	imageData, err := h.activeQRCodeService.GetActiveQRCodeImage(uint(id))
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

// AddStaticQRCode 为活码添加静态码
func (h *ActiveQRCodeHandler) AddStaticQRCode(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	var req models.StaticQRCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	staticQRCode, err := h.activeQRCodeService.AddStaticQRCode(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Static QR code added successfully",
		Data:    staticQRCode,
	})
}

// RedirectByShortCode 通过短码重定向
func (h *ActiveQRCodeHandler) RedirectByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// 添加防缓存响应头
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 获取用户信息
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	region := "" // 可以根据IP获取地区信息

	targetURL, err := h.activeQRCodeService.GetTargetURL(shortCode, userAgent, ipAddress, region)
	if err != nil {
		// 检查是否为自定义的QRCodeError
		if _, ok := err.(*services.QRCodeError); ok {
			// 对于二维码相关错误，返回友好的HTML页面
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK) // 返回200状态码

			// 读取错误页面模板
			errorHTML := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>二维码无效 - 活码管理系统</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.8.1/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        body {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }
        .error-container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
            padding: 40px;
            text-align: center;
            max-width: 500px;
            width: 90%;
        }
        .error-icon {
            font-size: 4rem;
            color: #dc3545;
            margin-bottom: 20px;
        }
        .error-title {
            color: #2c3e50;
            font-size: 1.8rem;
            font-weight: 600;
            margin-bottom: 15px;
        }
        .error-message {
            color: #6c757d;
            font-size: 1.1rem;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .contact-info {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            margin-top: 20px;
        }
        .contact-title {
            color: #495057;
            font-weight: 600;
            margin-bottom: 10px;
        }
        .contact-text {
            color: #6c757d;
            font-size: 0.95rem;
        }
        @media (max-width: 576px) {
            .error-container {
                padding: 30px 20px;
            }
            .error-title {
                font-size: 1.5rem;
            }
            .error-message {
                font-size: 1rem;
            }
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-icon">
            <i class="bi bi-exclamation-triangle-fill"></i>
        </div>
        
        <h1 class="error-title">二维码已过期或不存在</h1>
        
        <p class="error-message">
            抱歉，您扫描的二维码可能已过期、被禁用或不存在。<br>
            请确认二维码是否有效，或联系管理员获取帮助。
        </p>
        
        <div class="contact-info">
            <div class="contact-title">
                <i class="bi bi-person-lines-fill me-2"></i>
                需要帮助？
            </div>
            <div class="contact-text">
                如有疑问，请联系系统管理员<br>
                或重新获取有效的二维码
            </div>
        </div>
        
        <div class="mt-4">
            <small class="text-muted">
                <i class="bi bi-shield-check me-1"></i>
                活码管理系统 - 安全可靠
            </small>
        </div>
    </div>
</body>
</html>`

			c.String(http.StatusOK, errorHTML)
			return
		}

		// 其他错误返回JSON响应
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Short code not found",
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, targetURL)
}

// ListStaticQRCodes 获取静态码列表
func (h *ActiveQRCodeHandler) ListStaticQRCodes(c *gin.Context) {
	// 从查询参数获取分页信息
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if l := c.Query("limit"); l != "" {
		if limitNum, err := strconv.Atoi(l); err == nil && limitNum > 0 && limitNum <= 100 {
			limit = limitNum
		}
	}

	// 从服务层获取静态码列表 - 目前使用简单的查询
	var staticQRs []models.StaticQRCode
	var total int64

	db := h.activeQRCodeService.GetDB() // 需要在服务中添加GetDB方法
	query := db.Model(&models.StaticQRCode{}).Preload("ActiveQRCode")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "查询失败",
		})
		return
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&staticQRs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取成功",
		Data: map[string]interface{}{
			"data":  staticQRs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// CreateStaticQRCode 创建静态码
func (h *ActiveQRCodeHandler) CreateStaticQRCode(c *gin.Context) {
	var req models.StaticQRCodeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务层创建静态码
	staticQR, err := h.activeQRCodeService.AddStaticQRCode(req.ActiveQRCodeID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "创建成功",
		Data:    staticQR,
	})
}

// GetStaticQRCode 获取静态码详情
func (h *ActiveQRCodeHandler) GetStaticQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	var staticQR models.StaticQRCode
	db := h.activeQRCodeService.GetDB()
	if err := db.Preload("ActiveQRCode").First(&staticQR, uint(id)).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "静态码不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "查询失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取成功",
		Data:    staticQR,
	})
}

// UpdateStaticQRCode 更新静态码
func (h *ActiveQRCodeHandler) UpdateStaticQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	var req models.StaticQRCodeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	db := h.activeQRCodeService.GetDB()
	var staticQR models.StaticQRCode
	if err := db.First(&staticQR, uint(id)).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "静态码不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "查询失败",
			})
		}
		return
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

	if err := db.Save(&staticQR).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "更新失败",
		})
		return
	}

	// 重新加载关联数据
	if err := db.Preload("ActiveQRCode").First(&staticQR, staticQR.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "重新加载失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "更新成功",
		Data:    staticQR,
	})
}

// DeleteStaticQRCode 删除静态码
func (h *ActiveQRCodeHandler) DeleteStaticQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	db := h.activeQRCodeService.GetDB()
	var staticQR models.StaticQRCode
	if err := db.First(&staticQR, uint(id)).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "静态码不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "查询失败",
			})
		}
		return
	}

	if err := db.Delete(&staticQR).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "删除成功",
	})
}

// ToggleActiveQRStatus 切换活码启用/禁用状态
func (h *ActiveQRCodeHandler) ToggleActiveQRStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	db := h.activeQRCodeService.GetDB()
	var activeQR models.ActiveQRCode
	if err := db.First(&activeQR, id).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "活码不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "查询失败",
			})
		}
		return
	}

	// 切换状态：1 -> 0, 0 -> 1
	newStatus := 1 - activeQR.Status
	if err := db.Model(&activeQR).Update("status", newStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "状态更新失败",
		})
		return
	}

	statusText := "启用"
	if newStatus == 0 {
		statusText = "禁用"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("活码已%s", statusText),
		Data: gin.H{
			"id":     activeQR.ID,
			"status": newStatus,
		},
	})
}

// ToggleStaticQRStatus 切换静态码启用/禁用状态
func (h *ActiveQRCodeHandler) ToggleStaticQRStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid ID parameter",
		})
		return
	}

	db := h.activeQRCodeService.GetDB()
	var staticQR models.StaticQRCode
	if err := db.First(&staticQR, id).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "静态码不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "查询失败",
			})
		}
		return
	}

	// 切换状态：1 -> 0, 0 -> 1
	newStatus := 1 - staticQR.Status
	if err := db.Model(&staticQR).Update("status", newStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "状态更新失败",
		})
		return
	}

	statusText := "启用"
	if newStatus == 0 {
		statusText = "禁用"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("静态码已%s", statusText),
		Data: gin.H{
			"id":     staticQR.ID,
			"status": newStatus,
		},
	})
}

// ParseQRCode 解析上传的二维码图片
func (h *ActiveQRCodeHandler) ParseQRCode(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("qrcode")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请选择要上传的二维码图片",
		})
		return
	}

	// 创建二维码解析器
	parser := qrcode.NewParser()
	
	// 解析二维码
	content, err := parser.ParseFromFile(file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("二维码解析失败: %s", err.Error()),
		})
		return
	}

	// 验证是否为有效URL
	if !parser.ValidateURL(content) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "解析的内容不是有效的URL地址",
			Data: gin.H{
				"content": content,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "二维码解析成功",
		Data: gin.H{
			"url":     content,
			"content": content,
		},
	})
}
