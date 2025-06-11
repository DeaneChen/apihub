package service

import (
	"context"
	"errors"

	"apihub/internal/auth/jwt"
	"apihub/internal/model"
	"apihub/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	store      store.Store
	jwtService *jwt.JWTService
}

// NewAuthService 创建认证服务实例
func NewAuthService(store store.Store, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		store:      store,
		jwtService: jwtService,
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// 根据用户名查找用户
	user, err := s.store.Users().GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, errors.New("用户账户已被禁用")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT Token
	tokenResponse, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 构造响应
	response := &model.LoginResponse{
		AccessToken: tokenResponse.AccessToken,
		ExpiresIn:   tokenResponse.ExpiresIn,
		TokenType:   "Bearer",
		User:        user.ToUserInfo(),
	}

	return response, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, tokenString string) (*model.LogoutResponse, error) {
	// 撤销Token（加入黑名单）
	err := s.jwtService.RevokeToken(tokenString)
	if err != nil {
		return nil, errors.New("登出失败")
	}

	// 构造响应
	response := &model.LogoutResponse{
		Message: "登出成功",
	}

	return response, nil
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(tokenString string) (*jwt.CustomClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	return s.store.Users().GetByID(ctx, userID)
}
