package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// ConfigRepository 系统配置仓库SQLite实现
type ConfigRepository struct {
	db DBExecutor
}

// Set 设置配置项
func (r *ConfigRepository) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO system_configs (config_key, config_value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(config_key) DO UPDATE SET
			config_value = excluded.config_value,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query, key, value, time.Now())
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to set config",
			Err:     err,
		}
	}

	return nil
}

// Get 获取配置项
func (r *ConfigRepository) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT config_value FROM system_configs WHERE config_key = ?`

	var value string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", &store.DBError{
				Code:    store.ErrNotFound,
				Message: "config not found",
			}
		}
		return "", &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get config",
			Err:     err,
		}
	}

	return value, nil
}

// GetAll 获取所有配置项
func (r *ConfigRepository) GetAll(ctx context.Context) ([]*model.SystemConfig, error) {
	query := `
		SELECT id, config_key, config_value, updated_at
		FROM system_configs
		ORDER BY config_key
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get all configs",
			Err:     err,
		}
	}
	defer rows.Close()

	var configs []*model.SystemConfig
	for rows.Next() {
		config := &model.SystemConfig{}
		err := rows.Scan(
			&config.ID, &config.ConfigKey, &config.ConfigValue, &config.UpdatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan config",
				Err:     err,
			}
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate configs",
			Err:     err,
		}
	}

	return configs, nil
}

// Delete 删除配置项
func (r *ConfigRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM system_configs WHERE config_key = ?`

	result, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to delete config",
			Err:     err,
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get affected rows",
			Err:     err,
		}
	}

	if rowsAffected == 0 {
		return &store.DBError{
			Code:    store.ErrNotFound,
			Message: "config not found",
		}
	}

	return nil
}

// BatchSet 批量设置配置项
func (r *ConfigRepository) BatchSet(ctx context.Context, configs map[string]string) error {
	// 开始事务（如果当前不在事务中）
	_, ok := r.db.(*sql.Tx)
	if !ok {
		// 如果不是事务，需要创建事务
		db, ok := r.db.(*sql.DB)
		if !ok {
			return &store.DBError{
				Code:    store.ErrTransactionFailed,
				Message: "invalid database executor",
			}
		}

		newTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return &store.DBError{
				Code:    store.ErrTransactionFailed,
				Message: "failed to begin transaction",
				Err:     err,
			}
		}
		defer newTx.Rollback()

		// 使用新事务执行批量操作
		txRepo := &ConfigRepository{db: newTx}
		for key, value := range configs {
			if err := txRepo.Set(ctx, key, value); err != nil {
				return err
			}
		}

		return newTx.Commit()
	}

	// 如果已经在事务中，直接执行
	for key, value := range configs {
		if err := r.Set(ctx, key, value); err != nil {
			return err
		}
	}

	return nil
}
