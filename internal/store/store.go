package store

import (
	"apihub/internal/model"
	"context"
)

// Store 存储层主接口
type Store interface {
	// 数据库连接管理
	Connect() error
	Close() error

	// 数据库迁移
	Migrate() error

	// 事务管理
	BeginTx(ctx context.Context) (Transaction, error)

	// 各个模型的CRUD接口
	Users() UserRepository
	APIKeys() APIKeyRepository
	Configs() ConfigRepository
	Quotas() QuotaRepository
	Services() ServiceRepository
	AccessLogs() AccessLogRepository
}

// Transaction 事务接口
type Transaction interface {
	Commit() error
	Rollback() error

	// 在事务中访问各个仓库
	Users() UserRepository
	APIKeys() APIKeyRepository
	Configs() ConfigRepository
	Quotas() QuotaRepository
	Services() ServiceRepository
	AccessLogs() AccessLogRepository
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	Count(ctx context.Context) (int, error)
}

// APIKeyRepository API密钥仓库接口
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *model.APIKey) error
	GetByID(ctx context.Context, id int) (*model.APIKey, error)
	GetByKey(ctx context.Context, key string) (*model.APIKey, error)
	GetByUserID(ctx context.Context, userID int) ([]*model.APIKey, error)
	Update(ctx context.Context, apiKey *model.APIKey) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.APIKey, error)
}

// ConfigRepository 系统配置仓库接口
type ConfigRepository interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context) ([]*model.SystemConfig, error)
	Delete(ctx context.Context, key string) error
	BatchSet(ctx context.Context, configs map[string]string) error
}

// QuotaRepository 服务配额仓库接口
type QuotaRepository interface {
	Create(ctx context.Context, quota *model.ServiceQuota) error
	GetByUserAndService(ctx context.Context, userID int, serviceName, timeWindow string) (*model.ServiceQuota, error)
	GetByUserID(ctx context.Context, userID int) ([]*model.ServiceQuota, error)
	Update(ctx context.Context, quota *model.ServiceQuota) error
	IncrementUsage(ctx context.Context, userID int, serviceName, timeWindow string, cost int) error
	ResetUsage(ctx context.Context, userID int, serviceName, timeWindow string) error
	List(ctx context.Context, offset, limit int) ([]*model.ServiceQuota, error)
}

// ServiceRepository 服务定义仓库接口
type ServiceRepository interface {
	Create(ctx context.Context, service *model.ServiceDefinition) error
	GetByID(ctx context.Context, id int) (*model.ServiceDefinition, error)
	GetByName(ctx context.Context, serviceName string) (*model.ServiceDefinition, error)
	Update(ctx context.Context, service *model.ServiceDefinition) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.ServiceDefinition, error)
	GetEnabled(ctx context.Context) ([]*model.ServiceDefinition, error)
}

// AccessLogRepository 访问日志仓库接口
type AccessLogRepository interface {
	Create(ctx context.Context, log *model.AccessLog) error
	GetByID(ctx context.Context, id int) (*model.AccessLog, error)
	GetByUserID(ctx context.Context, userID int, offset, limit int) ([]*model.AccessLog, error)
	GetByAPIKeyID(ctx context.Context, apiKeyID int, offset, limit int) ([]*model.AccessLog, error)
	GetUsageStats(ctx context.Context, userID int, serviceName, startDate, endDate string) (*model.UsageStatsResponse, error)
	List(ctx context.Context, offset, limit int) ([]*model.AccessLog, error)
	DeleteOldLogs(ctx context.Context, beforeDate string) error
}

// DBError 数据库错误类型
type DBError struct {
	Code    int
	Message string
	Err     error
}

func (e *DBError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// 错误代码常量
const (
	ErrConnectionFailed = iota + 1000
	ErrMigrationFailed
	ErrDataConstraint
	ErrNotFound
	ErrDuplicateKey
	ErrTransactionFailed
)
