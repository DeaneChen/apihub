-- 重新创建服务定义表，包含新字段
CREATE TABLE IF NOT EXISTS service_definitions_new (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    default_limit INTEGER DEFAULT -1,
    status      INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    allow_anonymous BOOLEAN NOT NULL DEFAULT 0,
    rate_limit INTEGER NOT NULL DEFAULT 60,
    quota_cost INTEGER NOT NULL DEFAULT 1
);

-- 复制数据到新表
INSERT INTO service_definitions_new 
SELECT 
    id, 
    service_name, 
    description, 
    default_limit, 
    status, 
    created_at, 
    updated_at,
    0 AS allow_anonymous,
    60 AS rate_limit,
    1 AS quota_cost
FROM service_definitions;

-- 删除旧表
DROP TABLE service_definitions;

-- 重命名新表为正式表名
ALTER TABLE service_definitions_new RENAME TO service_definitions;

-- 重新创建索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_definitions_name ON service_definitions(service_name);

-- 更新现有服务定义的新字段值
UPDATE service_definitions SET allow_anonymous = 1, rate_limit = 60, quota_cost = 1 WHERE service_name = 'text_processing';
UPDATE service_definitions SET allow_anonymous = 0, rate_limit = 30, quota_cost = 2 WHERE service_name = 'image_processing';
UPDATE service_definitions SET allow_anonymous = 0, rate_limit = 20, quota_cost = 5 WHERE service_name = 'data_analysis';

-- 添加两个新的示例服务
INSERT OR IGNORE INTO service_definitions 
(service_name, description, default_limit, status, allow_anonymous, rate_limit, quota_cost) 
VALUES 
('echo', '回显服务，返回请求的内容', 1000, 1, 1, 60, 1),
('time', '时间服务，返回当前服务器时间', 1000, 1, 1, 60, 1); 