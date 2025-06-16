package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
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

	// 创建迁移表（如果不存在）
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrMigrationFailed,
			Message: "failed to create migrations table",
			Err:     err,
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

	// 获取已应用的迁移
	rows, err := s.db.Query("SELECT name FROM migrations")
	if err != nil {
		return &store.DBError{
			Code:    store.ErrMigrationFailed,
			Message: "failed to query migrations",
			Err:     err,
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("关闭迁移查询时出错: %v", closeErr)
		}
	}()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: "failed to scan migration name",
				Err:     err,
			}
		}
		appliedMigrations[name] = true
	}

	if err := rows.Err(); err != nil {
		return &store.DBError{
			Code:    store.ErrMigrationFailed,
			Message: "failed to iterate migrations",
			Err:     err,
		}
	}

	// 按文件名排序
	var migrationNames []string
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		migrationNames = append(migrationNames, entry.Name())
	}
	sort.Strings(migrationNames)

	// 按顺序执行迁移
	for _, name := range migrationNames {
		// 如果已应用，跳过
		if appliedMigrations[name] {
			log.Printf("迁移 %s 已应用，跳过", name)
			continue
		}

		log.Printf("应用迁移 %s", name)

		// 注意：embed.FS 总是使用正斜杠，即使在 Windows 上也是如此
		migrationPath := "migrations/" + name
		content, err := migrationFiles.ReadFile(migrationPath)
		if err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to read migration file %s", name),
				Err:     err,
			}
		}

		// 开始事务
		tx, err := s.db.Begin()
		if err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to begin transaction for migration %s", name),
				Err:     err,
			}
		}

		// 执行迁移SQL
		if _, err := tx.Exec(string(content)); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("回滚事务时出错: %v", rollbackErr)
			}
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to execute migration %s", name),
				Err:     err,
			}
		}

		// 记录迁移
		if _, err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", name); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("回滚事务时出错: %v", rollbackErr)
			}
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to record migration %s", name),
				Err:     err,
			}
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			return &store.DBError{
				Code:    store.ErrMigrationFailed,
				Message: fmt.Sprintf("failed to commit migration %s", name),
				Err:     err,
			}
		}

		log.Printf("迁移 %s 应用成功", name)
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
