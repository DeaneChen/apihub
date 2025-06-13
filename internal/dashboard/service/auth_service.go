package service

import (
	"context"
	"errors"
	"time"

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

// UpdateProfile 更新用户个人资料
func (s *AuthService) UpdateProfile(ctx context.Context, userID int, req *model.UpdateProfileRequest) (*model.User, error) {
	// 获取用户
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 更新邮箱（如果提供）
	if req.Email != "" && req.Email != user.Email {
		// 检查邮箱是否已被其他用户使用
		existingUser, _ := s.store.Users().GetByEmail(ctx, req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("邮箱已被其他用户使用")
		}
		user.Email = req.Email
	}

	// 更新时间
	user.UpdatedAt = time.Now()

	// 保存到数据库
	err = s.store.Users().Update(ctx, user)
	if err != nil {
		return nil, errors.New("更新个人资料失败: " + err.Error())
	}

	return user, nil
}

// ChangePassword 修改用户密码
func (s *AuthService) ChangePassword(ctx context.Context, userID int, req *model.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证当前密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("当前密码错误")
	}

	// 检查新密码是否与当前密码相同
	if req.CurrentPassword == req.NewPassword {
		return errors.New("新密码不能与当前密码相同")
	}

	// 对新密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码处理失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	// 保存到数据库
	err = s.store.Users().Update(ctx, user)
	if err != nil {
		return errors.New("修改密码失败: " + err.Error())
	}

	return nil
}
