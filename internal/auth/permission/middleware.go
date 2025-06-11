package permission

import (
	"net/http"

	"apihub/internal/auth/jwt"
	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// RequirePermissionMiddleware 要求特定权限的中间件
func RequirePermissionMiddleware(permissionService *PermissionService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := jwt.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未找到用户角色"))
			c.Abort()
			return
		}

		// 检查权限
		if !permissionService.HasPermission(userRole, requiredPermission) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "权限不足"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermissionMiddleware 要求任意一个权限的中间件
func RequireAnyPermissionMiddleware(permissionService *PermissionService, requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := jwt.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未找到用户角色"))
			c.Abort()
			return
		}

		// 检查权限
		if !permissionService.HasAnyPermission(userRole, requiredPermissions) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "权限不足"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissionsMiddleware 要求所有权限的中间件
func RequireAllPermissionsMiddleware(permissionService *PermissionService, requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := jwt.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未找到用户角色"))
			c.Abort()
			return
		}

		// 检查权限
		if !permissionService.HasAllPermissions(userRole, requiredPermissions) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "权限不足"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireResourceAccessMiddleware 要求资源访问权限的中间件
func RequireResourceAccessMiddleware(permissionService *PermissionService, permission string, getResourceUserID func(*gin.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userRole, roleExists := jwt.GetUserRole(c)
		userID, userIDExists := jwt.GetUserID(c)

		if !roleExists || !userIDExists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未找到用户信息"))
			c.Abort()
			return
		}

		// 获取资源所属用户ID
		resourceUserID, err := getResourceUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "获取资源信息失败: "+err.Error()))
			c.Abort()
			return
		}

		// 检查资源访问权限
		if !permissionService.CanAccessResource(userRole, userID, resourceUserID, permission) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "无权访问该资源"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员可访问的中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := jwt.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未找到用户角色"))
			c.Abort()
			return
		}

		if userRole != RoleAdmin {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "需要管理员权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}
