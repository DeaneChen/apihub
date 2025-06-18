package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"apihub/internal/model"
	"apihub/internal/store"
)

// Config 系统配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type"`
	DSN      string `json:"dsn"`
	MaxConns int    `json:"max_conns"`
	MaxIdle  int    `json:"max_idle"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWT struct {
		Secret       string        `json:"secret"`
		AccessExpiry time.Duration `json:"access_expiry"`
		Issuer       string        `json:"issuer"`
	} `json:"jwt"`
	APIKey struct {
		Secret string `json:"secret"`
	} `json:"apikey"`
	Cache struct {
		DefaultExpiration time.Duration `json:"default_expiration"`
		CleanupInterval   time.Duration `json:"cleanup_interval"`
	} `json:"cache"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Path   string `json:"path"`
}

// LoadConfig 加载配置
// 优先级: 环境变量 > 配置文件 > 数据库 > 默认值
func LoadConfig(configPath string, store store.Store) (*Config, error) {
	// 1. 加载默认配置
	config := defaultConfig()

	// 2. 尝试加载.env文件（如果有godotenv包）
	// _ = godotenv.Load()

	// 3. 尝试从配置文件加载
	if configPath != "" {
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %w", err)
		}
	}

	// 4. 从环境变量覆盖
	overrideFromEnv(config)

	// 5. 如果提供了存储实例，尝试从数据库加载密钥
	if store != nil {
		if err := loadSecretsFromDB(store, config); err != nil {
			return nil, fmt.Errorf("从数据库加载密钥失败: %w", err)
		}
	}

	return config, nil
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  60,
			WriteTimeout: 60,
		},
		Database: DatabaseConfig{
			Type:     "sqlite",
			DSN:      "apihub.db",
			MaxConns: 10,
			MaxIdle:  5,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
			Path:   "logs",
		},
	}

	// 设置JWT配置
	config.Auth.JWT.Secret = ""
	config.Auth.JWT.AccessExpiry = 30 * time.Minute
	config.Auth.JWT.Issuer = "apihub"

	// 设置APIKey配置
	config.Auth.APIKey.Secret = ""

	// 设置缓存配置
	config.Auth.Cache.DefaultExpiration = 30 * time.Minute
	config.Auth.Cache.CleanupInterval = 10 * time.Minute

	return config
}

// loadFromFile 从文件加载配置
func loadFromFile(path string, config *Config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，使用默认配置
			return nil
		}
		return err
	}

	// 根据文件扩展名决定解析方式
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return json.Unmarshal(data, config)
	case ".yaml", ".yml":
		// 如果需要支持YAML，可以添加yaml包依赖
		return fmt.Errorf("暂不支持YAML格式配置文件")
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", ext)
	}
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *Config) {
	// 服务器配置
	if port := os.Getenv("APIHUB_SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Server.Port)
	}
	if host := os.Getenv("APIHUB_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}

	// 数据库配置
	if dbType := os.Getenv("APIHUB_DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dsn := os.Getenv("APIHUB_DB_DSN"); dsn != "" {
		config.Database.DSN = dsn
	}

	// 认证配置
	if jwtSecret := os.Getenv("APIHUB_JWT_SECRET"); jwtSecret != "" {
		config.Auth.JWT.Secret = jwtSecret
	}
	if apiKeySecret := os.Getenv("APIHUB_APIKEY_SECRET"); apiKeySecret != "" {
		config.Auth.APIKey.Secret = apiKeySecret
	}

	// 日志配置
	if logLevel := os.Getenv("APIHUB_LOG_LEVEL"); logLevel != "" {
		config.Log.Level = logLevel
	}
}

// loadSecretsFromDB 从数据库加载密钥
func loadSecretsFromDB(store store.Store, config *Config) error {
	ctx := context.Background()

	// 如果JWT密钥未设置，尝试从数据库加载
	if config.Auth.JWT.Secret == "" {
		jwtSecret, err := store.Configs().Get(ctx, model.ConfigKeyJWTSecret)
		if err == nil {
			config.Auth.JWT.Secret = jwtSecret
		} else {
			// 忽略未找到的错误
			// 其他错误需要处理
			if !isNotFoundError(err) {
				return fmt.Errorf("获取JWT密钥失败: %w", err)
			}
		}
	}

	// 如果APIKey密钥未设置，尝试从数据库加载
	if config.Auth.APIKey.Secret == "" {
		apiKeySecret, err := store.Configs().Get(ctx, model.ConfigKeyAPIKeySecret)
		if err == nil {
			config.Auth.APIKey.Secret = apiKeySecret
		} else {
			// 忽略未找到的错误
			// 其他错误需要处理
			if !isNotFoundError(err) {
				return fmt.Errorf("获取APIKey密钥失败: %w", err)
			}
		}
	}

	return nil
}

// isNotFoundError 检查是否为"未找到"错误
func isNotFoundError(err error) bool {
	// 这里需要根据实际的错误类型进行判断
	// 由于没有完整的错误类型定义，这里使用字符串匹配
	return strings.Contains(err.Error(), "not found")
}

// GenerateConfigFile 生成配置文件
func GenerateConfigFile(path string) error {
	config := defaultConfig()

	// 设置示例密钥
	config.Auth.JWT.Secret = "change-me-in-production"
	config.Auth.APIKey.Secret = "change-me-in-production-32-chars-key"

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 确保目标目录存在
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	return ioutil.WriteFile(path, data, 0644)
}
