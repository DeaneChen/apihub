package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// AccessLogRepository 访问日志仓库SQLite实现
type AccessLogRepository struct {
	db DBExecutor
}

// Create 创建访问日志
func (r *AccessLogRepository) Create(ctx context.Context, log *model.AccessLog) error {
	query := `
		INSERT INTO access_logs (api_key_id, user_id, service_name, endpoint, status, cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	log.CreatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		log.APIKeyID, log.UserID, log.ServiceName, log.Endpoint,
		log.Status, log.Cost, log.CreatedAt,
	)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to create access log",
			Err:     err,
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get access log ID",
			Err:     err,
		}
	}

	log.ID = int(id)
	return nil
}

// GetByID 根据ID获取访问日志
func (r *AccessLogRepository) GetByID(ctx context.Context, id int) (*model.AccessLog, error) {
	query := `
		SELECT id, api_key_id, user_id, service_name, endpoint, status, cost, created_at
		FROM access_logs WHERE id = ?
	`

	log := &model.AccessLog{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.APIKeyID, &log.UserID, &log.ServiceName,
		&log.Endpoint, &log.Status, &log.Cost, &log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &store.DBError{
				Code:    store.ErrNotFound,
				Message: "access log not found",
			}
		}
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get access log",
			Err:     err,
		}
	}

	return log, nil
}

// GetByUserID 根据用户ID获取访问日志
func (r *AccessLogRepository) GetByUserID(ctx context.Context, userID int, offset, limit int) ([]*model.AccessLog, error) {
	query := `
		SELECT id, api_key_id, user_id, service_name, endpoint, status, cost, created_at
		FROM access_logs 
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get access logs",
			Err:     err,
		}
	}
	defer rows.Close()

	var logs []*model.AccessLog
	for rows.Next() {
		log := &model.AccessLog{}
		err := rows.Scan(
			&log.ID, &log.APIKeyID, &log.UserID, &log.ServiceName,
			&log.Endpoint, &log.Status, &log.Cost, &log.CreatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan access log",
				Err:     err,
			}
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate access logs",
			Err:     err,
		}
	}

	return logs, nil
}

// GetByAPIKeyID 根据API密钥ID获取访问日志
func (r *AccessLogRepository) GetByAPIKeyID(ctx context.Context, apiKeyID int, offset, limit int) ([]*model.AccessLog, error) {
	query := `
		SELECT id, api_key_id, user_id, service_name, endpoint, status, cost, created_at
		FROM access_logs 
		WHERE api_key_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get access logs",
			Err:     err,
		}
	}
	defer rows.Close()

	var logs []*model.AccessLog
	for rows.Next() {
		log := &model.AccessLog{}
		err := rows.Scan(
			&log.ID, &log.APIKeyID, &log.UserID, &log.ServiceName,
			&log.Endpoint, &log.Status, &log.Cost, &log.CreatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan access log",
				Err:     err,
			}
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate access logs",
			Err:     err,
		}
	}

	return logs, nil
}

// GetUsageStats 获取使用统计
func (r *AccessLogRepository) GetUsageStats(ctx context.Context, userID int, serviceName, startDate, endDate string) (*model.UsageStatsResponse, error) {
	// 构建基础查询
	baseQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total_calls,
			SUM(CASE WHEN status >= 200 AND status < 300 THEN 1 ELSE 0 END) as success_calls,
			SUM(CASE WHEN status >= 400 THEN 1 ELSE 0 END) as error_calls,
			SUM(cost) as total_cost
		FROM access_logs 
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?
	`

	args := []interface{}{userID, startDate + " 00:00:00", endDate + " 23:59:59"}

	if serviceName != "" {
		baseQuery += " AND service_name = ?"
		args = append(args, serviceName)
	}

	baseQuery += " GROUP BY DATE(created_at) ORDER BY date"

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to get usage stats",
			Err:     err,
		}
	}
	defer rows.Close()

	stats := &model.UsageStatsResponse{
		UserID:      userID,
		ServiceName: serviceName,
		DailyUsage:  make(map[string]int),
		Details:     []model.AccessLogSummary{},
	}

	totalUsage := 0
	for rows.Next() {
		var summary model.AccessLogSummary
		err := rows.Scan(
			&summary.Date, &summary.TotalCalls, &summary.SuccessCalls,
			&summary.ErrorCalls, &summary.TotalCost,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan usage stats",
				Err:     err,
			}
		}

		stats.DailyUsage[summary.Date] = summary.TotalCalls
		stats.Details = append(stats.Details, summary)
		totalUsage += summary.TotalCalls
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate usage stats",
			Err:     err,
		}
	}

	stats.TotalUsage = totalUsage
	return stats, nil
}

// List 获取访问日志列表
func (r *AccessLogRepository) List(ctx context.Context, offset, limit int) ([]*model.AccessLog, error) {
	query := `
		SELECT id, api_key_id, user_id, service_name, endpoint, status, cost, created_at
		FROM access_logs 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to list access logs",
			Err:     err,
		}
	}
	defer rows.Close()

	var logs []*model.AccessLog
	for rows.Next() {
		log := &model.AccessLog{}
		err := rows.Scan(
			&log.ID, &log.APIKeyID, &log.UserID, &log.ServiceName,
			&log.Endpoint, &log.Status, &log.Cost, &log.CreatedAt,
		)
		if err != nil {
			return nil, &store.DBError{
				Code:    store.ErrDataConstraint,
				Message: "failed to scan access log",
				Err:     err,
			}
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to iterate access logs",
			Err:     err,
		}
	}

	return logs, nil
}

// DeleteOldLogs 删除旧日志
func (r *AccessLogRepository) DeleteOldLogs(ctx context.Context, beforeDate string) error {
	query := `DELETE FROM access_logs WHERE created_at < ?`

	result, err := r.db.ExecContext(ctx, query, beforeDate)
	if err != nil {
		return &store.DBError{
			Code:    store.ErrDataConstraint,
			Message: "failed to delete old logs",
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

	fmt.Printf("Deleted %d old access logs\n", rowsAffected)
	return nil
}
