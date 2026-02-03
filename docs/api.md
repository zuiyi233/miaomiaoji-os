// API文档
// 本文档记录所有API接口

## 基础信息

- **Base URL**: `/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token

## 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page_info": {
      "page": 1,
      "size": 10,
      "total": 100
    }
  }
}
```

### 错误响应
```json
{
  "code": 10001,
  "message": "请求参数错误",
  "data": null
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 10001 | 请求参数错误 |
| 10002 | 未授权访问 |
| 10003 | 禁止访问 |
| 10004 | 资源不存在 |
| 10005 | 资源已存在 |
| 10006 | 服务器内部错误 |
| 10007 | 数据库操作失败 |
| 10008 | 外部接口调用失败 |
| 10009 | 数据校验失败 |
| 10010 | 请求频率超限 |
| 10011 | 文件操作失败 |
| 10012 | 请求超时 |

---

## 健康检查

### 健康状态
- **URL**: `GET /healthz`
- **描述**: 检查服务是否正常运行
- **认证**: 否
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "message": "Service is healthy"
  }
}
```

### 就绪检查
- **URL**: `GET /ready`
- **描述**: 检查服务是否就绪（包括数据库连接）
- **认证**: 否
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ready",
    "database": "connected"
  }
}
```

---

## 认证接口

### 用户注册
- **URL**: `POST /api/v1/auth/register`
- **描述**: 用户注册
- **认证**: 否
- **请求体**:
```json
{
  "username": "string (required, min:3, max:50)",
  "password": "string (required, min:6, max:100)",
  "email": "string (optional, email)",
  "nickname": "string (optional, max:50)"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "string",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "string",
      "nickname": "string",
      "email": "string",
      "role": "user",
      "points": 0
    }
  }
}
```

### 用户登录
- **URL**: `POST /api/v1/auth/login`
- **描述**: 用户登录
- **认证**: 否
- **请求体**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```
- **响应**: 同注册接口

### 用户退出
- **URL**: `POST /api/v1/auth/logout`
- **描述**: 用户退出（JWT无状态，客户端删除token）
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success"
}
```

### 刷新Token
- **URL**: `GET /api/v1/auth/refresh`
- **描述**: 刷新JWT Token
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "string",
    "expires_in": 86400
  }
}
```

---

## 用户接口

### 获取当前用户信息
- **URL**: `GET /api/v1/users/profile`
- **描述**: 获取当前登录用户信息
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "string",
    "nickname": "string",
    "email": "string",
    "role": "user",
    "points": 0,
    "check_in_streak": 0
  }
}
```

### 更新用户信息
- **URL**: `PUT /api/v1/users/profile`
- **描述**: 更新当前用户信息
- **认证**: 是
- **请求体**:
```json
{
  "nickname": "string (optional, max:50)",
  "email": "string (optional, email)"
}
```
- **响应**: 同获取用户信息

### 每日签到
- **URL**: `POST /api/v1/users/check-in`
- **描述**: 用户每日签到
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "签到成功"
  }
}
```

### 获取积分详情
- **URL**: `GET /api/v1/users/points`
- **描述**: 获取用户积分和签到信息
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "points": 0,
    "check_in_streak": 0
  }
}
```

### 获取用户列表（管理员）
- **URL**: `GET /api/v1/users?page=1&size=10`
- **描述**: 获取用户列表（仅管理员）
- **认证**: 是（admin角色）
- **响应**: 分页响应

### 更新用户状态（管理员）
- **URL**: `PUT /api/v1/users/:id/status`
- **描述**: 禁用/启用用户（仅管理员）
- **认证**: 是（admin角色）
- **请求体**:
```json
{
  "status": 0 // 0=禁用, 1=启用
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success"
}
```

---

## 项目接口

### 获取项目列表
- **URL**: `GET /api/v1/projects?page=1&size=10`
- **描述**: 获取当前用户的项目列表
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "string",
        "genre": "string",
        "tags": [],
        "core_conflict": "string",
        "character_arc": "string",
        "ultimate_value": "string",
        "world_rules": "string",
        "ai_settings": {},
        "user_id": 1,
        "created_at": "2024-01-01 00:00:00",
        "updated_at": "2024-01-01 00:00:00"
      }
    ],
    "page_info": {
      "page": 1,
      "size": 10,
      "total": 1
    }
  }
}
```

### 创建项目
- **URL**: `POST /api/v1/projects`
- **描述**: 创建新项目
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (required, max:200)",
  "genre": "string (optional, max:50)",
  "tags": ["string"],
  "core_conflict": "string",
  "character_arc": "string",
  "ultimate_value": "string",
  "world_rules": "string",
  "ai_settings": {}
}
```
- **响应**: 项目详情

### 获取项目详情
- **URL**: `GET /api/v1/projects/:id`
- **描述**: 获取项目详情
- **认证**: 是
- **响应**: 项目详情

### 更新项目
- **URL**: `PUT /api/v1/projects/:id`
- **描述**: 更新项目信息
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (optional, max:200)",
  "genre": "string (optional, max:50)",
  "tags": ["string"],
  "core_conflict": "string",
  "character_arc": "string",
  "ultimate_value": "string",
  "world_rules": "string",
  "ai_settings": {}
}
```
- **响应**: 项目详情

### 删除项目
- **URL**: `DELETE /api/v1/projects/:id`
- **描述**: 删除项目（软删除）
- **认证**: 是
- **响应**:
```json
{
  "code": 0,
  "message": "success"
}
```

### 导出项目
- **URL**: `GET /api/v1/projects/:id/export`
- **描述**: 导出项目为JSON
- **认证**: 是
- **响应**: 项目详情（含关联数据）

---

## 插件接口

### 创建插件
- **URL**: `POST /api/v1/plugins`
- **描述**: 创建插件记录（默认禁用）
- **认证**: 是
- **请求体**:
```json
{
  "name": "string (required)",
  "version": "string",
  "author": "string",
  "description": "string",
  "endpoint": "string (插件服务URL，如 http://127.0.0.1:9000)",
  "entry_point": "string"
}
```
- **响应（data）**: Plugin

### 获取插件列表
- **URL**: `GET /api/v1/plugins?page=1&page_size=20`
- **描述**: 分页获取插件列表
- **认证**: 是
- **响应（data）**:
```json
{
  "plugins": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

### 获取插件详情
- **URL**: `GET /api/v1/plugins/:id`
- **描述**: 获取单个插件（含 capabilities）
- **认证**: 是
- **响应（data）**: Plugin

### 更新插件
- **URL**: `PUT /api/v1/plugins/:id`
- **描述**: 更新插件信息
- **认证**: 是
- **请求体**:
```json
{
  "name": "string",
  "version": "string",
  "author": "string",
  "description": "string",
  "endpoint": "string",
  "entry_point": "string",
  "is_enabled": true
}
```
- **响应（data）**: Plugin

### 删除插件
- **URL**: `DELETE /api/v1/plugins/:id`
- **描述**: 删除插件
- **认证**: 是
- **响应**: 成功响应

### 启用插件
- **URL**: `PUT /api/v1/plugins/:id/enable`
- **描述**: 启用插件
- **认证**: 是
- **响应**: 成功响应

### 禁用插件
- **URL**: `PUT /api/v1/plugins/:id/disable`
- **描述**: 禁用插件
- **认证**: 是
- **响应**: 成功响应

### Ping 插件
- **URL**: `POST /api/v1/plugins/:id/ping`
- **描述**: 更新插件最后一次 ping 时间（不做真实网络探测）
- **认证**: 是
- **响应**: 成功响应

### 获取插件能力列表
- **URL**: `GET /api/v1/plugins/:plugin_id/capabilities`
- **描述**: 获取插件能力列表
- **认证**: 是
- **响应（data）**: PluginCapability[]

### 添加插件能力
- **URL**: `POST /api/v1/plugins/:plugin_id/capabilities`
- **描述**: 添加插件能力
- **认证**: 是
- **请求体**:
```json
{
  "cap_id": "string (required)",
  "name": "string (required)",
  "type": "string (required)",
  "description": "string",
  "icon": "string"
}
```
- **响应（data）**: PluginCapability

### 删除插件能力
- **URL**: `DELETE /api/v1/plugins/capabilities/:id`
- **描述**: 删除插件能力
- **认证**: 是
- **响应**: 成功响应

### 调用插件
- **URL**: `POST /api/v1/plugins/:id/invoke`
- **描述**: 调用插件能力（服务端转发请求到插件 endpoint）
- **认证**: 是（会将 `Authorization` header 转发给插件）
- **请求体**:
```json
{
  "method": "string (required)",
  "payload": {}
}
```
- **转发规则**:
  - 若 plugin.endpoint 为 `http(s)://host[:port]`（无 path 或 path 为 `/`），则默认请求 `POST {endpoint}/invoke`
  - 若 plugin.endpoint 自带 path，则直接请求该 URL
  - 超时：30s
  - 会更新插件健康状态与延迟（healthy/latency_ms）
- **响应（data）**: 插件返回的 JSON 对象；若非 JSON，则返回 `{ "raw": "..." }`

### 异步调用插件
- **URL**: `POST /api/v1/plugins/:id/invoke-async`
- **描述**: 创建异步任务调用插件（长任务建议使用），返回 job_uuid，结果可轮询 jobs 接口或通过 SSE 订阅 session_id 获取进度
- **认证**: 是（会将 `Authorization` header 转发给插件；仅在本进程内存保存，不落库，服务重启可能导致未执行任务失败）
- **请求体**:
```json
{
  "session_id": 1,
  "method": "string (required)",
  "payload": {}
}
```
- **响应**: HTTP 202
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "job_uuid": "uuid-string",
    "status": "queued"
  }
}
```
- **响应头**:
  - `Location: /api/v1/jobs/{job_uuid}`

---

## 任务接口

### 获取任务详情
- **URL**: `GET /api/v1/jobs/:job_uuid`
- **描述**: 获取任务状态与结果
- **认证**: 是
- **响应（data）**: Job

### 取消任务
- **URL**: `POST /api/v1/jobs/:job_uuid/cancel`
- **描述**: 取消任务（若任务正在本进程内执行，会尽力取消；否则仅标记 canceled）
- **认证**: 是
- **响应（data）**: Job

---

## 数据模型

### User 用户
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 用户ID |
| username | string | 用户名 |
| nickname | string | 昵称 |
| email | string | 邮箱 |
| role | string | 角色(user/admin) |
| status | int | 状态(0=禁用,1=启用) |
| points | int | 积分 |
| check_in_streak | int | 连续签到天数 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### Project 项目
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 项目ID |
| title | string | 标题 |
| genre | string | 类型/题材 |
| tags | []string | 标签 |
| core_conflict | string | 核心冲突 |
| character_arc | string | 人物弧光 |
| ultimate_value | string | 终极价值 |
| world_rules | string | 世界观规则 |
| ai_settings | object | AI设置 |
| user_id | uint | 所属用户ID |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |
