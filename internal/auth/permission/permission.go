package permission

// 权限常量定义
const (
	// 用户相关权限
	PermCreateUser = "user:create"
	PermReadUser   = "user:read"
	PermUpdateUser = "user:update"
	PermDeleteUser = "user:delete"
	PermListUsers  = "user:list"

	// API密钥相关权限
	PermCreateAPIKey = "apikey:create"
	PermReadAPIKey   = "apikey:read"
	PermUpdateAPIKey = "apikey:update"
	PermDeleteAPIKey = "apikey:delete"
	PermListAPIKeys  = "apikey:list"

	// 服务相关权限
	PermCreateService = "service:create"
	PermReadService   = "service:read"
	PermUpdateService = "service:update"
	PermDeleteService = "service:delete"
	PermListServices  = "service:list"
	PermUseService    = "service:use"

	// 配额相关权限
	PermCreateQuota = "quota:create"
	PermReadQuota   = "quota:read"
	PermUpdateQuota = "quota:update"
	PermDeleteQuota = "quota:delete"
	PermListQuotas  = "quota:list"

	// 系统配置相关权限
	PermCreateConfig = "config:create"
	PermReadConfig   = "config:read"
	PermUpdateConfig = "config:update"
	PermDeleteConfig = "config:delete"
	PermListConfigs  = "config:list"

	// 访问日志相关权限
	PermReadAccessLog  = "accesslog:read"
	PermListAccessLogs = "accesslog:list"

	// 系统管理权限
	PermSystemAdmin = "system:admin"
	PermSystemRead  = "system:read"
)

// 角色常量定义
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleGuest = "guest"
)

// RolePermissions 角色权限映射
var RolePermissions = map[string][]string{
	RoleAdmin: {
		// 管理员拥有所有权限
		PermCreateUser, PermReadUser, PermUpdateUser, PermDeleteUser, PermListUsers,
		PermCreateAPIKey, PermReadAPIKey, PermUpdateAPIKey, PermDeleteAPIKey, PermListAPIKeys,
		PermCreateService, PermReadService, PermUpdateService, PermDeleteService, PermListServices, PermUseService,
		PermCreateQuota, PermReadQuota, PermUpdateQuota, PermDeleteQuota, PermListQuotas,
		PermCreateConfig, PermReadConfig, PermUpdateConfig, PermDeleteConfig, PermListConfigs,
		PermReadAccessLog, PermListAccessLogs,
		PermSystemAdmin, PermSystemRead,
	},
	RoleUser: {
		// 普通用户权限
		PermReadUser, PermUpdateUser, // 只能读取和更新自己的信息
		PermCreateAPIKey, PermReadAPIKey, PermUpdateAPIKey, PermDeleteAPIKey, PermListAPIKeys, // 管理自己的API密钥
		PermReadService, PermListServices, PermUseService, // 使用服务
		PermReadQuota, PermListQuotas, // 查看配额
		PermReadAccessLog, PermListAccessLogs, // 查看自己的访问日志
	},
	RoleGuest: {
		// 访客权限（最小权限）
		PermReadService, PermListServices, // 只能查看服务列表
	},
}

// PermissionService 权限服务
type PermissionService struct{}

// NewPermissionService 创建权限服务实例
func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

// HasPermission 检查角色是否具有指定权限
func (s *PermissionService) HasPermission(role, permission string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// HasAnyPermission 检查角色是否具有任意一个指定权限
func (s *PermissionService) HasAnyPermission(role string, permissions []string) bool {
	for _, permission := range permissions {
		if s.HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查角色是否具有所有指定权限
func (s *PermissionService) HasAllPermissions(role string, permissions []string) bool {
	for _, permission := range permissions {
		if !s.HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// GetRolePermissions 获取角色的所有权限
func (s *PermissionService) GetRolePermissions(role string) []string {
	permissions, exists := RolePermissions[role]
	if !exists {
		return []string{}
	}

	// 返回权限副本，避免外部修改
	result := make([]string, len(permissions))
	copy(result, permissions)
	return result
}

// IsValidRole 检查角色是否有效
func (s *PermissionService) IsValidRole(role string) bool {
	_, exists := RolePermissions[role]
	return exists
}

// GetAllRoles 获取所有可用角色
func (s *PermissionService) GetAllRoles() []string {
	roles := make([]string, 0, len(RolePermissions))
	for role := range RolePermissions {
		roles = append(roles, role)
	}
	return roles
}

// CanAccessResource 检查用户是否可以访问指定资源
func (s *PermissionService) CanAccessResource(userRole string, userID int, resourceUserID int, permission string) bool {
	// 管理员可以访问所有资源
	if userRole == RoleAdmin {
		return s.HasPermission(userRole, permission)
	}

	// 普通用户只能访问自己的资源
	if userID == resourceUserID {
		return s.HasPermission(userRole, permission)
	}

	return false
}
