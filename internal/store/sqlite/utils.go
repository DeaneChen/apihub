package sqlite

import (
	"strings"
)

// isUniqueConstraintError 检查是否为唯一约束错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "duplicate")
}

// isForeignKeyConstraintError 检查是否为外键约束错误
func isForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "foreign key")
}

// isNotNullConstraintError 检查是否为非空约束错误
func isNotNullConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not null constraint") ||
		strings.Contains(errStr, "not null")
}
