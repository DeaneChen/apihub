-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    email       TEXT UNIQUE,
    role        TEXT NOT NULL DEFAULT 'user',
    status      INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- API密钥表
CREATE TABLE IF NOT EXISTS api_keys (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,
    key_name    TEXT NOT NULL,
    api_key     TEXT NOT NULL UNIQUE,
    status      INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at  DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建API密钥表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(api_key);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key  TEXT NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统配置表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(config_key);

-- 访问日志表
CREATE TABLE IF NOT EXISTS access_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id  INTEGER NOT NULL,
    user_id     INTEGER NOT NULL,
    service_name TEXT NOT NULL,
    endpoint    TEXT NOT NULL,
    status      INTEGER NOT NULL,
    cost        INTEGER DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建访问日志表索引
CREATE INDEX IF NOT EXISTS idx_access_logs_user_service_time ON access_logs(user_id, service_name, created_at);
CREATE INDEX IF NOT EXISTS idx_access_logs_api_key_id ON access_logs(api_key_id);

-- 服务配额表
CREATE TABLE IF NOT EXISTS service_quotas (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,
    service_name TEXT NOT NULL,
    time_window TEXT NOT NULL,
    usage       INTEGER DEFAULT 0,
    limit_value INTEGER DEFAULT -1,
    reset_time  DATETIME NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, service_name, time_window)
);

-- 创建服务配额表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_quotas_user_service_window ON service_quotas(user_id, service_name, time_window);

-- 服务定义表
CREATE TABLE IF NOT EXISTS service_definitions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    default_limit INTEGER DEFAULT -1,
    status      INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建服务定义表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_definitions_name ON service_definitions(service_name);

-- 插入默认系统配置
INSERT OR IGNORE INTO system_configs (config_key, config_value) VALUES 
('system_initialized', 'false'),
('default_quota_limit', '1000'),
('system_title', 'APIHub'),
('system_description', '统一API业务服务框架'),
('registration_open', 'false');

-- 插入默认服务定义
INSERT OR IGNORE INTO service_definitions (service_name, description, default_limit, status) VALUES 
('text_processing', '文本处理服务', 1000, 1),
('image_processing', '图像处理服务', 500, 1),
('data_analysis', '数据分析服务', 200, 1); 