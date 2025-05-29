package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// QuotaRepository 服务配额仓库SQLite实现
type QuotaRepository struct {
	db DBExecutor
}

// Create 创建服务配额
func (r *QuotaRepository) Create(ctx context.Context, quota *model.ServiceQuota) error {
	query := `
		INSERT INTO service_quotas (user_id, service_name, time_window, usage, limit_value, reset_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	quota.CreatedAt = now
	quota.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		quota.UserID, quota.ServiceName, quota.TimeWindow, quota.Usage,
		quota.LimitValue, quota.ResetTime, quota.CreatedAt, quota.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return &store.DBError{
				Code:    store.ErrDuplicateKey,
				Message: "quota already exists for this user and service",
				Err:     err,
			}
		}
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to create quota",
			Err:     err,
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get quota ID",
			Err:     err,
		}
	}

	quota.ID = int(id)
	return nil
}

// GetByUserAndService 根据用户和服务获取配额
func (r *QuotaRepository) GetByUserAndService(ctx context.Context, userID int, serviceName, timeWindow string) (*model.ServiceQuota, error) {
	query := `
		SELECT id, user_id, service_name, time_window, usage, limit_value, reset_time, created_at, updated_at
		FROM service_quotas 
		WHERE user_id = ? AND service_name = ? AND time_window = ?
	`

	quota := &model.ServiceQuota{}
	err := r.db.QueryRowContext(ctx, query, userID, serviceName, timeWindow).Scan(
		&quota.ID, &quota.UserID, &quota.ServiceName, &quota.TimeWindow,
		&quota.Usage, &quota.LimitValue, &quota.ResetTime, &quota.CreatedAt, &quota.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "quota not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get quota",
			Err:     err,
		}
	}

	return quota, nil
}

// GetByUserID 根据用户ID获取所有配额
func (r *QuotaRepository) GetByUserID(ctx context.Context, userID int) ([]*model.ServiceQuota, error) {
	query := `
		SELECT id, user_id, service_name, time_window, usage, limit_value, reset_time, created_at, updated_at
		FROM service_quotas 
		WHERE user_id = ?
		ORDER BY service_name, time_window
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get quotas",
			Err:     err,
		}
	}
	defer rows.Close()

	var quotas []*model.ServiceQuota
	for rows.Next() {
		quota := &model.ServiceQuota{}
		err := rows.Scan(
			&quota.ID, &quota.UserID, &quota.ServiceName, &quota.TimeWindow,
			&quota.Usage, &quota.LimitValue, &quota.ResetTime, &quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan quota",
				Err:     err,
			}
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate quotas",
			Err:     err,
		}
	}

	return quotas, nil
}

// Update 更新配额
func (r *QuotaRepository) Update(ctx context.Context, quota *model.ServiceQuota) error {
	query := `
		UPDATE service_quotas 
		SET usage = ?, limit_value = ?, reset_time = ?, updated_at = ?
		WHERE id = ?
	`

	quota.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		quota.Usage, quota.LimitValue, quota.ResetTime, quota.UpdatedAt, quota.ID,
	)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to update quota",
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
			Message: "quota not found",
		}
	}

	return nil
}

// IncrementUsage 增加使用量
func (r *QuotaRepository) IncrementUsage(ctx context.Context, userID int, serviceName, timeWindow string, cost int) error {
	query := `
		UPDATE service_quotas 
		SET usage = usage + ?, updated_at = ?
		WHERE user_id = ? AND service_name = ? AND time_window = ?
	`

	result, err := r.db.ExecContext(ctx, query, cost, time.Now(), userID, serviceName, timeWindow)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to increment usage",
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
			Message: "quota not found",
		}
	}

	return nil
}

// ResetUsage 重置使用量
func (r *QuotaRepository) ResetUsage(ctx context.Context, userID int, serviceName, timeWindow string) error {
	query := `
		UPDATE service_quotas 
		SET usage = 0, updated_at = ?
		WHERE user_id = ? AND service_name = ? AND time_window = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID, serviceName, timeWindow)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to reset usage",
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
			Message: "quota not found",
		}
	}

	return nil
}

// List 获取配额列表
func (r *QuotaRepository) List(ctx context.Context, offset, limit int) ([]*model.ServiceQuota, error) {
	query := `
		SELECT id, user_id, service_name, time_window, usage, limit_value, reset_time, created_at, updated_at
		FROM service_quotas 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to list quotas",
			Err:     err,
		}
	}
	defer rows.Close()

	var quotas []*model.ServiceQuota
	for rows.Next() {
		quota := &model.ServiceQuota{}
		err := rows.Scan(
			&quota.ID, &quota.UserID, &quota.ServiceName, &quota.TimeWindow,
			&quota.Usage, &quota.LimitValue, &quota.ResetTime, &quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan quota",
				Err:     err,
			}
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate quotas",
			Err:     err,
		}
	}

	return quotas, nil
}
