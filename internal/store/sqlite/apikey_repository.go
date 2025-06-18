package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// APIKeyRepository API密钥仓库SQLite实现
type APIKeyRepository struct {
	db DBExecutor
}

// Create 创建API密钥
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *model.APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, key_name, api_key, status, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	apiKey.CreatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		apiKey.UserID, apiKey.KeyName, apiKey.APIKey, apiKey.Status,
		apiKey.CreatedAt, apiKey.ExpiresAt,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return &store.DBError{
				Code:    store.ErrDuplicateKey,
				Message: "API key already exists",
				Err:     err,
			}
		}
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to create API key",
			Err:     err,
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get API key ID",
			Err:     err,
		}
	}

	apiKey.ID = int(id)
	return nil
}

// GetByID 根据ID获取API密钥
func (r *APIKeyRepository) GetByID(ctx context.Context, id int) (*model.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, status, created_at, expires_at
		FROM api_keys WHERE id = ?
	`

	apiKey := &model.APIKey{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyName, &apiKey.APIKey,
		&apiKey.Status, &apiKey.CreatedAt, &apiKey.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "API key not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get API key",
			Err:     err,
		}
	}

	return apiKey, nil
}

// GetByKey 根据密钥获取API密钥
func (r *APIKeyRepository) GetByKey(ctx context.Context, key string) (*model.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, status, created_at, expires_at
		FROM api_keys WHERE api_key = ?
	`

	apiKey := &model.APIKey{}
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyName, &apiKey.APIKey,
		&apiKey.Status, &apiKey.CreatedAt, &expiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "API密钥未找到",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "获取API密钥失败",
			Err:     err,
		}
	}

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	return apiKey, nil
}

// GetByUserID 根据用户ID获取API密钥列表
func (r *APIKeyRepository) GetByUserID(ctx context.Context, userID int) ([]*model.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, status, created_at, expires_at
		FROM api_keys WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get API keys",
			Err:     err,
		}
	}
	defer rows.Close()

	var apiKeys []*model.APIKey
	for rows.Next() {
		apiKey := &model.APIKey{}
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.KeyName, &apiKey.APIKey,
			&apiKey.Status, &apiKey.CreatedAt, &apiKey.ExpiresAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan API key",
				Err:     err,
			}
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate API keys",
			Err:     err,
		}
	}

	return apiKeys, nil
}

// Update 更新API密钥
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *model.APIKey) error {
	query := `
		UPDATE api_keys 
		SET key_name = ?, api_key = ?, status = ?, expires_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		apiKey.KeyName, apiKey.APIKey, apiKey.Status, apiKey.ExpiresAt, apiKey.ID,
	)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "更新API密钥失败",
			Err:     err,
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "获取受影响行数失败",
			Err:     err,
		}
	}

	if rowsAffected == 0 {
		return &store.DBError{
			Code:    store.ErrNotFound,
			Message: "API密钥未找到",
		}
	}

	return nil
}

// Delete 删除API密钥
func (r *APIKeyRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM api_keys WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to delete API key",
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
			Message: "API key not found",
		}
	}

	return nil
}

// List 获取API密钥列表
func (r *APIKeyRepository) List(ctx context.Context, offset, limit int) ([]*model.APIKey, error) {
	query := `
		SELECT id, user_id, key_name, api_key, status, created_at, expires_at
		FROM api_keys 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to list API keys",
			Err:     err,
		}
	}
	defer rows.Close()

	var apiKeys []*model.APIKey
	for rows.Next() {
		apiKey := &model.APIKey{}
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.KeyName, &apiKey.APIKey,
			&apiKey.Status, &apiKey.CreatedAt, &apiKey.ExpiresAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan API key",
				Err:     err,
			}
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate API keys",
			Err:     err,
		}
	}

	return apiKeys, nil
}
