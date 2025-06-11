package model

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"` // 访问令牌过期时间(秒)
	TokenType   string    `json:"token_type"`
	User        *UserInfo `json:"user"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	// 可以为空，从Authorization头获取token
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Message string `json:"message"`
}

// UserInfo 用户信息（用于响应）
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

// ToUserInfo 将User模型转换为UserInfo
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:       int(u.ID),
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		Status:   u.Status,
	}
}
