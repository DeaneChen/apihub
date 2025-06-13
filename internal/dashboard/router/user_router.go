package router

import (
	"apihub/internal/auth/jwt"
	"apihub/internal/dashboard/handler"
	"apihub/internal/dashboard/service"
	"apihub/internal/middleware"
	"apihub/internal/model"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// UserRouter 用户路由
type UserRouter struct {
	userHandler *handler.UserHandler
	jwtService  *jwt.JWTService
}

// NewUserRouter 创建用户路由实例
func NewUserRouter(store store.Store, jwtService *jwt.JWTService) *UserRouter {
	// 创建用户服务
	userService := service.NewUserService(store)

	// 创建用户处理器
	userHandler := handler.NewUserHandler(userService)

	return &UserRouter{
		userHandler: userHandler,
		jwtService:  jwtService,
	}
}

// RegisterRoutes 注册用户相关路由
func (r *UserRouter) RegisterRoutes(router *gin.RouterGroup) {
	// 用户路由组，需要JWT认证
	userGroup := router.Group("/user")
	userGroup.Use(middleware.JWTOnlyMiddleware(r.jwtService))

	// 添加管理员角色检查中间件
	userGroup.Use(jwt.RequireRole(model.RoleAdmin))

	{
		// @Summary      获取用户列表
		// @Description  分页获取系统中的用户列表
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        page      query  int     false  "页码，默认1"       minimum(1)
		// @Param        page_size query  int     false  "每页数量，默认20"  minimum(1) maximum(100)
		// @Success      200       {object}  model.APIResponse{data=model.UserListResponse}
		// @Failure      401       {object}  model.APIResponse
		// @Failure      403       {object}  model.APIResponse
		// @Failure      500       {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/list [get]
		userGroup.GET("/list", r.userHandler.ListUsers)

		// @Summary      获取用户信息
		// @Description  根据用户ID获取用户详细信息
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        id       path      int                       true  "用户ID"
		// @Success      200      {object}  model.APIResponse{data=model.UserInfo}
		// @Failure      400      {object}  model.APIResponse
		// @Failure      401      {object}  model.APIResponse
		// @Failure      403      {object}  model.APIResponse
		// @Failure      404      {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/info/:id [get]
		userGroup.GET("/info/:id", r.userHandler.GetUserInfo)

		// @Summary      创建用户
		// @Description  创建新的系统用户
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        request  body      handler.CreateUserRequest  true  "创建用户请求"
		// @Success      200      {object}  model.APIResponse{data=model.UserInfo}
		// @Failure      400      {object}  model.APIResponse
		// @Failure      401      {object}  model.APIResponse
		// @Failure      403      {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/create [post]
		userGroup.POST("/create", r.userHandler.CreateUser)

		// @Summary      更新用户
		// @Description  更新指定用户的信息
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        id       path      int                       true  "用户ID"
		// @Param        request  body      handler.UpdateUserRequest  true  "更新用户请求"
		// @Success      200      {object}  model.APIResponse{data=model.UserInfo}
		// @Failure      400      {object}  model.APIResponse
		// @Failure      401      {object}  model.APIResponse
		// @Failure      403      {object}  model.APIResponse
		// @Failure      404      {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/update/:id [post]
		userGroup.POST("/update/:id", r.userHandler.UpdateUser)

		// @Summary      删除用户
		// @Description  删除指定的用户
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        request  body      handler.DeleteUserRequest  true  "删除用户请求"
		// @Success      200      {object}  model.APIResponse
		// @Failure      400      {object}  model.APIResponse
		// @Failure      401      {object}  model.APIResponse
		// @Failure      403      {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/delete [post]
		userGroup.POST("/delete", r.userHandler.DeleteUser)

		// @Summary      重置用户密码
		// @Description  重置指定用户的密码
		// @Tags         用户管理
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        request  body      handler.ResetPasswordRequest  true  "重置密码请求"
		// @Success      200      {object}  model.APIResponse
		// @Failure      400      {object}  model.APIResponse
		// @Failure      401      {object}  model.APIResponse
		// @Failure      403      {object}  model.APIResponse
		// @Router       /api/v1/dashboard/user/reset-password [post]
		userGroup.POST("/reset-password", r.userHandler.ResetPassword)
	}
}
