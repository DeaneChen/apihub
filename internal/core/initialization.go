package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"apihub/internal/model"
	"apihub/internal/store"
	"apihub/internal/store/sqlite"
)

// InitializationService 系统初始化服务
type InitializationService struct {
	store store.Store
}

// NewInitializationService 创建初始化服务
func NewInitializationService(store store.Store) *InitializationService {
	return &InitializationService{
		store: store,
	}
}

// InitializeSystem 初始化系统
func (s *InitializationService) InitializeSystem(ctx context.Context) error {
	log.Println("开始系统初始化...")

	// 1. 连接数据库
	if err := s.store.Connect(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	log.Println("数据库连接成功")

	// 2. 执行数据库迁移
	if err := s.store.Migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	log.Println("数据库迁移完成")

	// 3. 检查系统是否已初始化
	initialized, err := s.isSystemInitialized(ctx)
	if err != nil {
		return fmt.Errorf("检查系统初始化状态失败: %w", err)
	}

	if initialized {
		log.Println("系统已初始化，跳过初始化步骤")
		return nil
	}

	// 4. 创建默认管理员账号
	adminUser, err := s.createDefaultAdmin(ctx)
	if err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}
	log.Printf("默认管理员创建成功: %s", adminUser.Username)

	// 5. 生成JWT密钥
	if err := s.generateJWTSecret(ctx); err != nil {
		return fmt.Errorf("生成JWT密钥失败: %w", err)
	}
	log.Println("JWT密钥生成完成")

	// 6. 生成APIKey密钥
	if err := s.generateAPIKeySecret(ctx); err != nil {
		return fmt.Errorf("生成APIKey密钥失败: %w", err)
	}
	log.Println("APIKey密钥生成完成")

	// 7. 标记系统已初始化
	if err := s.markSystemInitialized(ctx); err != nil {
		return fmt.Errorf("标记系统初始化失败: %w", err)
	}

	log.Println("系统初始化完成")
	return nil
}

// isSystemInitialized 检查系统是否已初始化
func (s *InitializationService) isSystemInitialized(ctx context.Context) (bool, error) {
	value, err := s.store.Configs().Get(ctx, model.ConfigKeySystemInitialized)
	if err != nil {
		if dbErr, ok := err.(*store.DBError); ok && dbErr.Code == store.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return value == "true", nil
}

// createDefaultAdmin 创建默认管理员账号
func (s *InitializationService) createDefaultAdmin(ctx context.Context) (*model.User, error) {
	// 检查是否已存在管理员
	users, err := s.store.Users().List(ctx, 0, 1)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		// 如果已有用户，检查是否有管理员
		for _, user := range users {
			if user.IsAdmin() {
				return user, nil
			}
		}
	}

	// 生成默认管理员密码
	defaultPassword := s.generateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建管理员用户
	admin := &model.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@apihub.local",
		Role:     model.RoleAdmin,
		Status:   model.UserStatusActive,
	}

	if err := s.store.Users().Create(ctx, admin); err != nil {
		return nil, err
	}

	// 输出默认密码到日志（仅用于开发环境）
	log.Printf("默认管理员账号创建成功:")
	log.Printf("用户名: %s", admin.Username)
	log.Printf("密码: %s", defaultPassword)
	log.Printf("请及时修改默认密码！")

	return admin, nil
}

// generateJWTSecret 生成JWT密钥
func (s *InitializationService) generateJWTSecret(ctx context.Context) error {
	// 检查是否已存在JWT密钥
	_, err := s.store.Configs().Get(ctx, model.ConfigKeyJWTSecret)
	if err == nil {
		return nil // 已存在，跳过
	}

	if dbErr, ok := err.(*store.DBError); !ok || dbErr.Code != store.ErrNotFound {
		return err // 其他错误
	}

	// 生成32字节的随机密钥
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("生成随机密钥失败: %w", err)
	}

	secretHex := hex.EncodeToString(secret)
	return s.store.Configs().Set(ctx, model.ConfigKeyJWTSecret, secretHex)
}

// generateAPIKeySecret 生成APIKey密钥
func (s *InitializationService) generateAPIKeySecret(ctx context.Context) error {
	// 检查是否已存在APIKey密钥
	_, err := s.store.Configs().Get(ctx, model.ConfigKeyAPIKeySecret)
	if err == nil {
		return nil // 已存在，跳过
	}

	if dbErr, ok := err.(*store.DBError); !ok || dbErr.Code != store.ErrNotFound {
		return err // 其他错误
	}

	// 生成32字节的随机密钥
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("生成随机密钥失败: %w", err)
	}

	secretHex := hex.EncodeToString(secret)
	return s.store.Configs().Set(ctx, model.ConfigKeyAPIKeySecret, secretHex)
}

// markSystemInitialized 标记系统已初始化
func (s *InitializationService) markSystemInitialized(ctx context.Context) error {
	return s.store.Configs().Set(ctx, model.ConfigKeySystemInitialized, "true")
}

// generateRandomPassword 生成随机密码
func (s *InitializationService) generateRandomPassword() string {
	// 生成8字节随机数据
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetSystemStatus 获取系统状态
func (s *InitializationService) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	status := &SystemStatus{}

	// 检查数据库连接
	if err := s.store.Connect(); err != nil {
		status.DatabaseConnected = false
		status.Errors = append(status.Errors, "数据库连接失败: "+err.Error())
	} else {
		status.DatabaseConnected = true
	}

	// 检查系统初始化状态
	initialized, err := s.isSystemInitialized(ctx)
	if err != nil {
		status.Errors = append(status.Errors, "检查初始化状态失败: "+err.Error())
	} else {
		status.SystemInitialized = initialized
	}

	// 统计用户数量
	userCount, err := s.store.Users().Count(ctx)
	if err != nil {
		status.Errors = append(status.Errors, "获取用户数量失败: "+err.Error())
	} else {
		status.UserCount = userCount
	}

	return status, nil
}

// SystemStatus 系统状态
type SystemStatus struct {
	DatabaseConnected bool     `json:"database_connected"`
	SystemInitialized bool     `json:"system_initialized"`
	UserCount         int      `json:"user_count"`
	Errors            []string `json:"errors,omitempty"`
}

// CreateSQLiteStore 创建SQLite存储实例
func CreateSQLiteStore(dsn string) store.Store {
	return sqlite.NewSQLiteStore(dsn)
}
