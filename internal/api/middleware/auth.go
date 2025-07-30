package middleware

import (
	"net/http"
	"strings"
	"wechat-active-qrcode/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// AuthRequired 需要认证的中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := m.authService.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not found",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)

		c.Next()
	}
}

// AdminRequired 需要管理员权限的中间件
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证中间件
		m.AuthRequired()(c)
		if c.IsAborted() {
			return
		}

		// 检查用户角色
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Role information not found",
			})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选的认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		user, err := m.authService.GetUserByID(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)

		c.Next()
	}
} 