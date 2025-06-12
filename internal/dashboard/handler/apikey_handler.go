package handler

import (
	"net/http"
	"time"

	"apihub/internal/auth/apikey"
	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler API密钥处理器
type APIKeyHandler struct {
	apiKeyService *apikey.APIKeyService
}

// NewAPIKeyHandler 创建API密钥处理器实例
func NewAPIKeyHandler(apiKeyService *apikey.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// ListAPIKeys 列出当前用户的所有API密钥
// @Summary 列出API密钥
// @Description 列出当前用户的所有API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.APIResponse{data=[]model.APIKey}
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /dashboard/apikeys/list [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
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

	// 获取用户的API密钥列表
	apiKeys, err := h.apiKeyService.GetAPIKeysByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"获取API密钥失败: "+err.Error(),
		))
		return
	}

	// 返回API密钥列表
	c.JSON(http.StatusOK, model.NewSuccessResponse(apiKeys))
}

// GenerateAPIKey 请求体
type GenerateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// GenerateAPIKey 为当前用户生成新的API密钥
// @Summary 生成API密钥
// @Description 为当前用户生成新的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateAPIKeyRequest true "API密钥生成请求"
// @Success 200 {object} model.APIResponse{data=model.APIKey}
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /dashboard/apikeys/generate [post]
func (h *APIKeyHandler) GenerateAPIKey(c *gin.Context) {
	var req GenerateAPIKeyRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

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

	// 生成API密钥
	apiKey, err := h.apiKeyService.CreateAPIKey(userID, req.Name, req.Description, req.ExpiresAt, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"生成API密钥失败: "+err.Error(),
		))
		return
	}

	// 返回生成的API密钥
	c.JSON(http.StatusOK, model.NewSuccessResponse(apiKey))
}

// DeleteAPIKeyRequest 删除API密钥请求
type DeleteAPIKeyRequest struct {
	APIKeyID int `json:"api_key_id" binding:"required"`
}

// DeleteAPIKey 删除指定的API密钥
// @Summary 删除API密钥
// @Description 删除当前用户的指定API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteAPIKeyRequest true "API密钥删除请求"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Failure 500 {object} model.APIResponse
// @Router /dashboard/apikeys/delete [post]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	var req DeleteAPIKeyRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			"请求参数错误: "+err.Error(),
		))
		return
	}

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

	// 首先验证API密钥是否属于当前用户
	apiKeys, err := h.apiKeyService.GetAPIKeysByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"验证API密钥所有权失败: "+err.Error(),
		))
		return
	}

	// 检查API密钥是否属于当前用户
	found := false
	for _, apiKey := range apiKeys {
		if apiKey.ID == req.APIKeyID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(
			model.CodeForbidden,
			"无权操作此API密钥",
		))
		return
	}

	// 删除API密钥
	if err := h.apiKeyService.DeleteAPIKey(req.APIKeyID); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"删除API密钥失败: "+err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		"message": "API密钥删除成功",
	}))
}
