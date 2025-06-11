package handler

import (
	"net/http"
	"strings"

	"apihub/internal/dashboard/service"
	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求"
// @Success 200 {object} model.APIResponse{data=model.LoginResponse}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

	// 调用服务层处理登录
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeInvalidCredentials,
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，撤销JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.APIResponse{data=model.LogoutResponse}
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从Authorization头获取Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeUnauthorized,
			"缺少Authorization头",
		))
		return
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeUnauthorized,
			"Authorization头格式错误",
		))
		return
	}

	// 提取Token
	tokenString := authHeader[len(bearerPrefix):]
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeUnauthorized,
			"Token不能为空",
		))
		return
	}

	// 调用服务层处理登出
	response, err := h.authService.Logout(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.APIResponse{data=model.UserInfo}
// @Failure 401 {object} model.APIResponse
// @Failure 404 {object} model.APIResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeUnauthorized,
			"用户信息不存在",
		))
		return
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			model.CodeUnauthorized,
			"用户ID格式错误",
		))
		return
	}

	// 获取用户信息
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(
			model.CodeNotFound,
			"用户不存在",
		))
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToUserInfo()))
}
