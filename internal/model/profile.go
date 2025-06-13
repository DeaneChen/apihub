package model

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// UpdateProfileResponse 更新个人资料响应
type UpdateProfileResponse struct {
	Message string    `json:"message"`
	User    *UserInfo `json:"user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Message string `json:"message"`
}
