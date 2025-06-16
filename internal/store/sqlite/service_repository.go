package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// ServiceRepository 服务定义仓库SQLite实现
type ServiceRepository struct {
	db DBExecutor
}

// Create 创建服务定义
func (r *ServiceRepository) Create(ctx context.Context, service *model.ServiceDefinition) error {
	query := `
		INSERT INTO service_definitions (service_name, description, default_limit, status, created_at, updated_at, allow_anonymous, rate_limit, quota_cost)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		service.ServiceName, service.Description, service.DefaultLimit,
		service.Status, service.CreatedAt, service.UpdatedAt,
		service.AllowAnonymous, service.RateLimit, service.QuotaCost,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return &store.DBError{
				Code:    store.ErrDuplicateKey,
				Message: "service name already exists",
				Err:     err,
			}
		}
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to create service",
			Err:     err,
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get service ID",
			Err:     err,
		}
	}

	service.ID = int(id)
	return nil
}

// GetByID 根据ID获取服务定义
func (r *ServiceRepository) GetByID(ctx context.Context, id int) (*model.ServiceDefinition, error) {
	query := `
		SELECT id, service_name, description, default_limit, status, created_at, updated_at, allow_anonymous, rate_limit, quota_cost
		FROM service_definitions WHERE id = ?
	`

	service := &model.ServiceDefinition{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&service.ID, &service.ServiceName, &service.Description,
		&service.DefaultLimit, &service.Status, &service.CreatedAt, &service.UpdatedAt,
		&service.AllowAnonymous, &service.RateLimit, &service.QuotaCost,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "service not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get service",
			Err:     err,
		}
	}

	return service, nil
}

// GetByName 根据服务名获取服务定义
func (r *ServiceRepository) GetByName(ctx context.Context, serviceName string) (*model.ServiceDefinition, error) {
	query := `
		SELECT id, service_name, description, default_limit, status, created_at, updated_at, allow_anonymous, rate_limit, quota_cost
		FROM service_definitions WHERE service_name = ?
	`

	service := &model.ServiceDefinition{}
	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(
		&service.ID, &service.ServiceName, &service.Description,
		&service.DefaultLimit, &service.Status, &service.CreatedAt, &service.UpdatedAt,
		&service.AllowAnonymous, &service.RateLimit, &service.QuotaCost,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "service not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get service",
			Err:     err,
		}
	}

	return service, nil
}

// Update 更新服务定义
func (r *ServiceRepository) Update(ctx context.Context, service *model.ServiceDefinition) error {
	query := `
		UPDATE service_definitions 
		SET description = ?, default_limit = ?, status = ?, updated_at = ?, allow_anonymous = ?, rate_limit = ?, quota_cost = ?
		WHERE id = ?
	`

	service.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		service.Description, service.DefaultLimit, service.Status,
		service.UpdatedAt, service.AllowAnonymous, service.RateLimit, service.QuotaCost,
		service.ID,
	)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to update service",
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
			Message: "service not found",
		}
	}

	return nil
}

// Delete 删除服务定义
func (r *ServiceRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM service_definitions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to delete service",
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
			Message: "service not found",
		}
	}

	return nil
}

// List 获取服务定义列表
func (r *ServiceRepository) List(ctx context.Context, offset, limit int) ([]*model.ServiceDefinition, error) {
	query := `
		SELECT id, service_name, description, default_limit, status, created_at, updated_at, allow_anonymous, rate_limit, quota_cost
		FROM service_definitions 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to list services",
			Err:     err,
		}
	}
	defer rows.Close()

	var services []*model.ServiceDefinition
	for rows.Next() {
		service := &model.ServiceDefinition{}
		err := rows.Scan(
			&service.ID, &service.ServiceName, &service.Description,
			&service.DefaultLimit, &service.Status, &service.CreatedAt, &service.UpdatedAt,
			&service.AllowAnonymous, &service.RateLimit, &service.QuotaCost,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan service",
				Err:     err,
			}
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate services",
			Err:     err,
		}
	}

	return services, nil
}

// GetEnabled 获取启用的服务定义列表
func (r *ServiceRepository) GetEnabled(ctx context.Context) ([]*model.ServiceDefinition, error) {
	query := `
		SELECT id, service_name, description, default_limit, status, created_at, updated_at, allow_anonymous, rate_limit, quota_cost
		FROM service_definitions 
		WHERE status = ?
		ORDER BY service_name
	`

	rows, err := r.db.QueryContext(ctx, query, model.ServiceStatusEnabled)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get enabled services",
			Err:     err,
		}
	}
	defer rows.Close()

	var services []*model.ServiceDefinition
	for rows.Next() {
		service := &model.ServiceDefinition{}
		err := rows.Scan(
			&service.ID, &service.ServiceName, &service.Description,
			&service.DefaultLimit, &service.Status, &service.CreatedAt, &service.UpdatedAt,
			&service.AllowAnonymous, &service.RateLimit, &service.QuotaCost,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan service",
				Err:     err,
			}
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate services",
			Err:     err,
		}
	}

	return services, nil
}
