package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // 不在JSON中显示密码
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`     // 'admin' or 'user'
	Status    int       `json:"status" db:"status"` // 0: disabled, 1: active
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// UserStatus 用户状态常量
const (
	UserStatusDisabled = 0
	UserStatusActive   = 1
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"email"`
	Role     string `json:"role" binding:"oneof=admin user"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Role   string `json:"role" binding:"omitempty,oneof=admin user"`
	Status int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
