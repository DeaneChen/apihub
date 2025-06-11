package permission

import (
	"fmt"
	"net/http"

	jwtAuth "apihub/internal/auth/jwt"

	"github.com/gin-gonic/gin"
)

// RequirePermissionMiddleware 要求特定权限的中间件
func RequirePermissionMiddleware(permissionService *PermissionService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := jwtAuth.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		// 检查权限
		if !permissionService.HasPermission(userRole, requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
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
		userRole, exists := jwtAuth.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		// 检查是否具有任意一个权限
		if !permissionService.HasAnyPermission(userRole, requiredPermissions) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
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
		userRole, exists := jwtAuth.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		// 检查是否具有所有权限
		if !permissionService.HasAllPermissions(userRole, requiredPermissions) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireResourceAccessMiddleware 要求资源访问权限的中间件
// 用于检查用户是否可以访问特定用户的资源
func RequireResourceAccessMiddleware(permissionService *PermissionService, requiredPermission string, resourceUserIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户信息
		userRole, exists := jwtAuth.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		userID, exists := jwtAuth.GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user ID not found",
			})
			c.Abort()
			return
		}

		// 获取资源用户ID
		resourceUserIDStr := c.Param(resourceUserIDParam)
		if resourceUserIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "resource user ID not provided",
			})
			c.Abort()
			return
		}

		// 将字符串转换为整数
		var resourceUserID int
		if _, err := fmt.Sscanf(resourceUserIDStr, "%d", &resourceUserID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "invalid resource user ID",
			})
			c.Abort()
			return
		}

		// 检查资源访问权限
		if !permissionService.CanAccessResource(userRole, userID, resourceUserID, requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员可访问的中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := jwtAuth.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		if userRole != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
