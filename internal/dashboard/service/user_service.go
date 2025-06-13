package service

import (
	"context"
	"errors"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// 系统常量
const (
	// 系统管理员ID
	SystemAdminID = 1
)

// UserService 用户服务
type UserService struct {
	store store.Store
}

// NewUserService 创建用户服务实例
func NewUserService(store store.Store) *UserService {
	return &UserService{
		store: store,
	}
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.store.Users().GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		existingUser, _ = s.store.Users().GetByEmail(ctx, req.Email)
		if existingUser != nil {
			return nil, errors.New("邮箱已被使用")
		}
	}

	// 对密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码处理失败")
	}

	// 创建用户对象
	now := time.Now()
	user := &model.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Role:      req.Role,
		Status:    model.UserStatusActive, // 默认激活状态
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 保存到数据库
	err = s.store.Users().Create(ctx, user)
	if err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID int, req *model.UpdateUserRequest) (*model.User, error) {
	// 获取用户
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 如果是系统管理员(ID=1)，不允许将其角色设置为非管理员
	if userID == SystemAdminID && req.Role != "" && req.Role != model.RoleAdmin {
		return nil, errors.New("不能将系统管理员设置为非管理员角色")
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

	// 更新角色（如果提供）
	if req.Role != "" {
		user.Role = req.Role
	}

	// 更新状态（如果提供）
	if req.Status == model.UserStatusActive || req.Status == model.UserStatusDisabled {
		user.Status = req.Status
	}

	// 更新时间
	user.UpdatedAt = time.Now()

	// 保存到数据库
	err = s.store.Users().Update(ctx, user)
	if err != nil {
		return nil, errors.New("更新用户失败: " + err.Error())
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	// 检查用户是否存在
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 不允许删除管理员用户
	if user.Role == model.RoleAdmin {
		return errors.New("不能删除管理员用户")
	}

	// 删除用户
	err = s.store.Users().Delete(ctx, userID)
	if err != nil {
		return errors.New("删除用户失败: " + err.Error())
	}

	return nil
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(ctx context.Context, userID int, newPassword string) error {
	// 检查用户是否存在
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 对新密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码处理失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	// 保存到数据库
	err = s.store.Users().Update(ctx, user)
	if err != nil {
		return errors.New("重置密码失败: " + err.Error())
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	user, err := s.store.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 获取用户总数
	total, err := s.store.Users().Count(ctx)
	if err != nil {
		return nil, 0, errors.New("获取用户总数失败: " + err.Error())
	}

	// 获取用户列表
	users, err := s.store.Users().List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取用户列表失败: " + err.Error())
	}

	return users, total, nil
}
