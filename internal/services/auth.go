package services

import (
	"errors"
	"wechat-active-qrcode/internal/auth"
	"wechat-active-qrcode/internal/models"
	"wechat-active-qrcode/pkg/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	jwtService *auth.JWTService
}

func NewAuthService(db *gorm.DB, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// Login 用户登录
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	var user models.User
	
	// 查找用户
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid username or password")
	}
	
	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}
	
	// 生成JWT token
	token, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return nil, err
	}
	
	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(req *models.LoginRequest) (*models.LoginResponse, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}
	
	// 加密密码
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	
	// 创建用户
	user := &models.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         "user",
	}
	
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	
	// 生成JWT token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, err
	}
	
	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// ValidateToken 验证token
func (s *AuthService) ValidateToken(tokenString string) (*auth.JWTClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	return s.jwtService.RefreshToken(tokenString)
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	
	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}
	
	// 加密新密码
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	
	// 更新密码
	user.PasswordHash = passwordHash
	return s.db.Save(&user).Error
} 