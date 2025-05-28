# 数据库设计文档

## 概述
本文档描述了UniAPI项目的数据库表结构设计。采用SQLite作为数据库，遵循轻量化原则。

## 表结构设计

### 1. 用户表 (users)
用于存储系统用户信息
```sql
CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    email       TEXT UNIQUE,
    role        TEXT NOT NULL DEFAULT 'user',  -- 'admin' or 'user'
    status      INTEGER NOT NULL DEFAULT 1,    -- 0: disabled, 1: active
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 2. API密钥表 (api_keys)
用于管理API访问密钥
```sql
CREATE TABLE api_keys (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,
    key_name    TEXT NOT NULL,
    api_key     TEXT NOT NULL UNIQUE,
    status      INTEGER NOT NULL DEFAULT 1,    -- 0: disabled, 1: active
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at  DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 3. 系统配置表 (system_configs)
存储系统全局配置
```sql
CREATE TABLE system_configs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key  TEXT NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 4. 访问日志表 (access_logs)
记录API调用历史
```sql
CREATE TABLE access_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id  INTEGER NOT NULL,
    user_id     INTEGER NOT NULL,              -- 添加用户ID便于统计
    service_name TEXT NOT NULL,                -- 内部服务名称
    endpoint    TEXT NOT NULL,
    status      INTEGER NOT NULL,
    cost        INTEGER DEFAULT 1,             -- API调用计费单位，默认1
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 5. 服务配额表 (service_quotas)
记录用户服务配额设置
```sql
CREATE TABLE service_quotas (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,              -- 用户ID
    service_name TEXT NOT NULL,                -- 内部服务名称
    time_window TEXT NOT NULL,                 -- 统计时间窗口（如：2024-03或2024-03-15）
    usage       INTEGER DEFAULT 0,             -- 当前使用量
    limit_value INTEGER DEFAULT -1,            -- -1表示无限制
    reset_time  DATETIME NOT NULL,             -- 下次重置时间
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE (user_id, service_name, time_window)
);
```

### 6. 服务定义表 (service_definitions)
系统支持的服务定义
```sql
CREATE TABLE service_definitions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name TEXT NOT NULL UNIQUE,         -- 服务名称，与代码中保持一致
    description TEXT NOT NULL,                 -- 服务描述
    default_limit INTEGER DEFAULT -1,          -- 默认限制值
    status      INTEGER NOT NULL DEFAULT 1,    -- 服务状态：1-启用，0-禁用
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 索引设计
- users表：username和email列创建唯一索引
- api_keys表：api_key列创建唯一索引
- system_configs表：config_key列创建唯一索引
- access_logs表：(user_id, service_name, created_at)创建组合索引
- service_quotas表：(user_id, service_name, time_window)创建唯一索引
- service_definitions表：service_name列创建唯一索引

## 注意事项
1. 所有时间字段采用UTC时间存储
2. 密码字段需要进行加密存储
3. API密钥需要进行加密存储
4. 配额统计支持按日、按月统计，time_window字段格式为YYYY-MM或YYYY-MM-DD
5. 服务名称必须与代码中的服务模块名称保持一致 