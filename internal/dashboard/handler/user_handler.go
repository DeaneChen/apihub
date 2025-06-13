package handler

import (
	"net/http"
	"strconv"

	"apihub/internal/dashboard/service"
	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsersRequest 列出用户请求
type ListUsersRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Role   string `json:"role" binding:"omitempty,oneof=admin user"`
	Status int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	UserID int `json:"user_id" binding:"required,min=1"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	UserID      int    `json:"user_id" binding:"required,min=1"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取系统中的用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1" minimum(1)
// @Param page_size query int false "每页数量，默认20" minimum(1) maximum(100)
// @Success 200 {object} model.APIResponse{data=model.UserListResponse}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /api/v1/dashboard/user/list [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 20

	// 绑定请求参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 调用服务层获取用户列表
	users, total, err := h.userService.ListUsers(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"获取用户列表失败: "+err.Error(),
		))
		return
	}

	// 转换为用户信息列表
	userInfos := make([]*model.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, user.ToUserInfo())
	}

	// 构造响应
	response := &model.UserListResponse{
		Total: total,
		Users: userInfos,
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} model.APIResponse{data=model.UserInfo}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Failure 404 {object} model.APIResponse
// @Router /api/v1/dashboard/user/info/{id} [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	// 获取用户ID
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"无效的用户ID",
		))
		return
	}

	// 调用服务层获取用户信息
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(
			model.CodeNotFound,
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToUserInfo()))
}

// CreateUser 创建用户
// @Summary 创建新用户
// @Description 创建新的系统用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.CreateUserRequest true "创建用户请求"
// @Success 200 {object} model.APIResponse{data=model.UserInfo}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /api/v1/dashboard/user/create [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 转换为模型请求
	modelReq := &model.CreateUserRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Role:     req.Role,
	}

	// 调用服务层创建用户
	user, err := h.userService.CreateUser(c.Request.Context(), modelReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"创建用户失败: "+err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToUserInfo()))
}

// UpdateUser 更新用户
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body handler.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} model.APIResponse{data=model.UserInfo}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 404 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /api/v1/dashboard/user/update/{id} [post]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest

	// 获取用户ID
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"无效的用户ID",
		))
		return
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 转换为模型请求
	modelReq := &model.UpdateUserRequest{
		Email:  req.Email,
		Role:   req.Role,
		Status: req.Status,
	}

	// 调用服务层更新用户
	user, err := h.userService.UpdateUser(c.Request.Context(), userID, modelReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"更新用户失败: "+err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToUserInfo()))
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.DeleteUserRequest true "删除用户请求"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /api/v1/dashboard/user/delete [post]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req DeleteUserRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 调用服务层删除用户
	err := h.userService.DeleteUser(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"删除用户失败: "+err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]string{
		"message": "用户删除成功",
	}))
}

// ResetPassword 重置用户密码
// @Summary 重置用户密码
// @Description 重置指定用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /api/v1/dashboard/user/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 调用服务层重置密码
	err := h.userService.ResetPassword(c.Request.Context(), req.UserID, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"重置密码失败: "+err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]string{
		"message": "密码重置成功",
	}))
}
