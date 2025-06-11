# 认证API文档

## 概述

认证API提供用户登录、登出等功能，用于控制对Dashboard和API接口的访问。

## 基础信息

- **Base URL**: `/api/v1`
- **认证方式**: JWT Bearer Token
- **响应格式**: JSON

## 统一响应格式

所有API响应都采用统一格式：

```json
{
    "code": 0,           // 状态码，0表示成功
    "message": "ok",     // 状态信息
    "data": {}          // 响应数据
}
```

### 状态码说明

| 状态码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 禁止访问 |
| 1004 | 资源不存在 |
| 1005 | 内部错误 |
| 1007 | 凭据无效 |
| 1008 | Token过期 |
| 1009 | Token无效 |

## API接口

### 1. 用户登录

**接口**: `POST /auth/login`

**描述**: 用户登录获取JWT Token

**请求参数**:
```json
{
    "username": "admin",
    "password": "password123"
}
```

**参数说明**:
- `username` (string, required): 用户名，1-50字符
- `password` (string, required): 密码，6-100字符

**成功响应**:
```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_in": 86400,
        "token_type": "Bearer",
        "user": {
            "id": 1,
            "username": "admin",
            "email": "admin@example.com",
            "role": "admin",
            "status": 1
        }
    }
}
```

**错误响应**:
```json
{
    "code": 1007,
    "message": "用户名或密码错误",
    "data": null
}
```

### 2. 用户登出

**接口**: `POST /auth/logout`

**描述**: 用户登出，撤销JWT Token

**请求头**:
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应**:
```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "message": "登出成功"
    }
}
```

### 3. 获取用户信息

**接口**: `GET /auth/profile`

**描述**: 获取当前登录用户的详细信息

**请求头**:
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应**:
```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "role": "admin",
        "status": 1
    }
}
```

## 使用示例

### 登录流程

1. **用户登录**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

2. **使用Token访问受保护的接口**:
```bash
curl -X GET http://localhost:8080/api/v1/dashboard/profile \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

3. **用户登出**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 受保护的路由

### Dashboard路由 (需要JWT认证)

- `GET /api/v1/dashboard/*` - 所有Dashboard相关接口
- 认证方式：JWT Bearer Token
- 权限：需要有效的访问令牌

### API路由 (支持JWT和APIKey认证)

- `POST /api/v1/api/*` - 所有API相关接口
- 认证方式：JWT Bearer Token 或 APIKey
- 权限：根据具体接口要求

## 错误处理

### 常见错误

1. **参数错误** (1001):
```json
{
    "code": 1001,
    "message": "请求参数错误: Key: 'LoginRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
    "data": null
}
```

2. **未授权** (1002):
```json
{
    "code": 1002,
    "message": "缺少Authorization头",
    "data": null
}
```

3. **Token无效** (1009):
```json
{
    "code": 1009,
    "message": "Token无效",
    "data": null
}
```

## 安全注意事项

1. **Token安全**:
   - 访问令牌有效期为24小时
   - 登出后Token会被加入黑名单

2. **密码安全**:
   - 密码使用bcrypt加密存储
   - 登录失败不会泄露具体错误信息

3. **HTTPS**:
   - 生产环境必须使用HTTPS
   - 避免Token在传输过程中被截获 