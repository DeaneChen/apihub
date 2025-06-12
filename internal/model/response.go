package model

// APIResponse 统一API响应格式
// @Description 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code" example:"0"`     // 状态码，0表示成功
	Message string      `json:"message" example:"ok"` // 状态信息
	Data    interface{} `json:"data"`                 // 响应数据
}

// 响应状态码常量
const (
	CodeSuccess            = 0    // 成功
	CodeInvalidParams      = 1001 // 参数错误
	CodeUnauthorized       = 1002 // 未授权
	CodeForbidden          = 1003 // 禁止访问
	CodeNotFound           = 1004 // 资源不存在
	CodeInternalError      = 1005 // 内部错误
	CodeDuplicateResource  = 1006 // 资源重复
	CodeInvalidCredentials = 1007 // 凭据无效
	CodeTokenExpired       = 1008 // Token过期
	CodeTokenInvalid       = 1009 // Token无效
)

// 响应消息常量
const (
	MsgSuccess            = "ok"
	MsgInvalidParams      = "参数错误"
	MsgUnauthorized       = "未授权"
	MsgForbidden          = "禁止访问"
	MsgNotFound           = "资源不存在"
	MsgInternalError      = "内部错误"
	MsgDuplicateResource  = "资源重复"
	MsgInvalidCredentials = "用户名或密码错误"
	MsgTokenExpired       = "Token已过期"
	MsgTokenInvalid       = "Token无效"
)

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *APIResponse {
	return &APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// NewErrorResponseWithData 创建带数据的错误响应
func NewErrorResponseWithData(code int, message string, data interface{}) *APIResponse {
	return &APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
