package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// UserRepository 用户仓库SQLite实现
type UserRepository struct {
	db DBExecutor
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password, email, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Password, user.Email, user.Role, user.Status,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return &store.DBError{
				Code:    store.ErrDuplicateKey,
				Message: "username or email already exists",
				Err:     err,
			}
		}
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to create user",
			Err:     err,
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get user ID",
			Err:     err,
		}
	}

	user.ID = int(id)
	return nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, password, email, role, status, created_at, updated_at
		FROM users WHERE id = ?
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "user not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get user",
			Err:     err,
		}
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password, email, role, status, created_at, updated_at
		FROM users WHERE username = ?
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "user not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get user",
			Err:     err,
		}
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, password, email, role, status, created_at, updated_at
		FROM users WHERE email = ?
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "user not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get user",
			Err:     err,
		}
	}

	return user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users 
		SET username = ?, password = ?, email = ?, role = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Password, user.Email, user.Role, user.Status,
		user.UpdatedAt, user.ID,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return &store.DBError{
				Code:    store.ErrDuplicateKey,
				Message: "username or email already exists",
				Err:     err,
			}
		}
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to update user",
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
			Message: "user not found",
		}
	}

	return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to delete user",
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
			Message: "user not found",
		}
	}

	return nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	query := `
		SELECT id, username, password, email, role, status, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to list users",
			Err:     err,
		}
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Email,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan user",
				Err:     err,
			}
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate users",
			Err:     err,
		}
	}

	return users, nil
}

// Count 获取用户总数
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to count users",
			Err:     err,
		}
	}

	return count, nil
}
