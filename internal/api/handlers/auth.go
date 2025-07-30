package handlers

import (
	"net/http"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Registration successful",
		Data:    response,
	})
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Authorization header required",
		})
		return
	}

	// 提取token（去掉Bearer前缀）
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		authHeader = authHeader[7:]
	}

	newToken, err := h.authService.RefreshToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data: gin.H{
			"token": newToken,
		},
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	err := h.authService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password changed successfully",
	})
} 