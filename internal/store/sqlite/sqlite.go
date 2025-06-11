package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"apihub/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// SQLiteStore SQLite存储实现
type SQLiteStore struct {
	db  *sql.DB
	dsn string
}

// SQLiteTransaction SQLite事务实现
type SQLiteTransaction struct {
	tx    *sql.Tx
	store *SQLiteStore
}

// NewSQLiteStore 创建新的SQLite存储实例
func NewSQLiteStore(dsn string) *SQLiteStore {
	return &SQLiteStore{
		dsn: dsn,
	}
}

// Connect 连接数据库
func (s *SQLiteStore) Connect() error {
	db, err := sql.Open("sqlite3", s.dsn)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrConnectionFailed,
			Message: "failed to open database",
			Err:     err,
		}
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return &store.DBError{
			Code:    store.ErrConnectionFailed,
			Message: "failed to ping database",
			Err:     err,
		}
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return &store.DBError{
			Code:    store.ErrConnectionFailed,
			Message: "failed to enable foreign keys",
			Err:     err,
		}
	}

	s.db = db
	return nil
}

// Close 关闭数据库连接
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Migrate 执行数据库迁移
func (s *SQLiteStore) Migrate() error {
	if s.db == nil {
		return &store.DBError{
			Code:    store.ErrMigrationFailed,
			Message: "database not connected",
		}
	}

	// 读取迁移文件
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return &store.DBError{
			Code:    store.ErrMigrationFailed,
			Message: "failed to read migration files",
			Err:     err,
		}
	}

	// 按文件名排序执行迁移
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// 注意：embed.FS 总是使用正斜杠，即使在 Windows 上也是如此
		migrationPath := "migrations/" + entry.Name()
		content, err := migrationFiles.ReadFile(migrationPath)
		if err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to read migration file %s", entry.Name()),
				Err:     err,
			}
		}

		// 执行迁移SQL
		if _, err := s.db.Exec(string(content)); err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to execute migration %s", entry.Name()),
				Err:     err,
			}
		}
	}

	return nil
}

// BeginTx 开始事务
func (s *SQLiteStore) BeginTx(ctx context.Context) (store.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrTransactionFailed,
			Message: "failed to begin transaction",
			Err:     err,
		}
	}

	return &SQLiteTransaction{
		tx:    tx,
		store: s,
	}, nil
}

// Users 返回用户仓库
func (s *SQLiteStore) Users() store.UserRepository {
	return &UserRepository{db: s.db}
}

// APIKeys 返回API密钥仓库
func (s *SQLiteStore) APIKeys() store.APIKeyRepository {
	return &APIKeyRepository{db: s.db}
}

// Configs 返回系统配置仓库
func (s *SQLiteStore) Configs() store.ConfigRepository {
	return &ConfigRepository{db: s.db}
}

// Quotas 返回服务配额仓库
func (s *SQLiteStore) Quotas() store.QuotaRepository {
	return &QuotaRepository{db: s.db}
}

// Services 返回服务定义仓库
func (s *SQLiteStore) Services() store.ServiceRepository {
	return &ServiceRepository{db: s.db}
}

// AccessLogs 返回访问日志仓库
func (s *SQLiteStore) AccessLogs() store.AccessLogRepository {
	return &AccessLogRepository{db: s.db}
}

// 事务方法实现

// Commit 提交事务
func (tx *SQLiteTransaction) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		return &store.DBError{
			Code:    store.ErrTransactionFailed,
			Message: "failed to commit transaction",
			Err:     err,
		}
	}
	return nil
}

// Rollback 回滚事务
func (tx *SQLiteTransaction) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		return &store.DBError{
			Code:    store.ErrTransactionFailed,
			Message: "failed to rollback transaction",
			Err:     err,
		}
	}
	return nil
}

// Users 返回事务中的用户仓库
func (tx *SQLiteTransaction) Users() store.UserRepository {
	return &UserRepository{db: tx.tx}
}

// APIKeys 返回事务中的API密钥仓库
func (tx *SQLiteTransaction) APIKeys() store.APIKeyRepository {
	return &APIKeyRepository{db: tx.tx}
}

// Configs 返回事务中的系统配置仓库
func (tx *SQLiteTransaction) Configs() store.ConfigRepository {
	return &ConfigRepository{db: tx.tx}
}

// Quotas 返回事务中的服务配额仓库
func (tx *SQLiteTransaction) Quotas() store.QuotaRepository {
	return &QuotaRepository{db: tx.tx}
}

// Services 返回事务中的服务定义仓库
func (tx *SQLiteTransaction) Services() store.ServiceRepository {
	return &ServiceRepository{db: tx.tx}
}

// AccessLogs 返回事务中的访问日志仓库
func (tx *SQLiteTransaction) AccessLogs() store.AccessLogRepository {
	return &AccessLogRepository{db: tx.tx}
}

// DBExecutor 数据库执行器接口，用于统一处理 *sql.DB 和 *sql.Tx
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
