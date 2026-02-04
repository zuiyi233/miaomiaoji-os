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
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "string",
    "expires_in": 86400,
    "must_change_password": false,
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

## AI 接口

### 获取模型列表
- **URL**: `GET /api/v1/ai/models?provider=xxx`
- **描述**: 获取指定供应商模型列表
- **认证**: 是
- **请求参数**:
  - `provider` string 必填（如: gemini/openai/proxy/local）
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "models": [
      { "id": "gemini-3-flash-preview", "name": "gemini-3-flash-preview", "provider": "gemini" }
    ]
  }
}
```

### 更新供应商配置（管理员）
- **URL**: `PUT /api/v1/ai/providers`
- **描述**: 更新供应商 BaseURL 与 API Key
- **认证**: 是（管理员）
- **请求体**:
```json
{
  "provider": "gemini",
  "base_url": "https://generativelanguage.googleapis.com",
  "api_key": "string"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success"
}
```

### 获取供应商配置（管理员）
- **URL**: `GET /api/v1/ai/providers?provider=xxx`
- **描述**: 获取供应商配置（API Key 为脱敏值）
- **认证**: 是（管理员）
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "provider": "gemini",
    "base_url": "https://generativelanguage.googleapis.com",
    "api_key": "abc***xyz"
  }
}
```

### 测试供应商连接（管理员）
- **URL**: `POST /api/v1/ai/providers/test`
- **描述**: 测试供应商连接是否可用
- **认证**: 是（管理员）
- **请求体**:
```json
{
  "provider": "gemini"
}
```
- **响应**: `success`

### AI 代理请求
- **URL**: `POST /api/v1/ai/proxy`
- **描述**: 代理调用第三方模型接口
- **认证**: 是（且需有效 AI 权限）
- **请求体**:
```json
{
  "provider": "gemini",
  "path": "v1beta/models/gemini-3-flash-preview:generateContent",
  "body": "{...}"
}
```
- **响应**: 透传上游 JSON

### AI 代理流式请求
- **URL**: `POST /api/v1/ai/proxy/stream`
- **描述**: 代理调用第三方模型流式接口
- **认证**: 是（且需有效 AI 权限）
- **请求体**:
```json
{
  "provider": "openai",
  "path": "v1/chat/completions",
  "body": "{...}"
}
```
- **响应**: `text/event-stream` 透传

---

## 兑换码接口

### 兑换码验证
- **URL**: `POST /api/v1/codes/redeem`
- **描述**: 验证兑换码并更新 AI 权限
- **认证**: 是
- **请求体**:
```json
{
  "request_id": "req_xxx",
  "idempotency_key": "redeem_xxx",
  "code": "XXXX-XXXX",
  "device_id": "string",
  "client_time": "2026-02-04T02:00:00Z",
  "app_id": "novel-agent-os",
  "platform": "web",
  "app_version": "1.0.0",
  "result_status": "success",
  "result_error_code": "",
  "entitlement_delta": {
    "entitlement_type": "ai_access",
    "grant_mode": "add_days",
    "start_at": "2026-02-04T02:00:00Z",
    "end_at": "2026-03-06T02:00:00Z",
    "plan_or_sku": "monthly"
  }
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "code": "XXXX-XXXX",
    "duration_days": 30,
    "ai_access_until": "2026-02-04T02:00:00Z",
    "used_count": 1,
    "status": "active"
  }
}
```

### 获取兑换码列表（管理员）
- **URL**: `GET /api/v1/codes?status=all&search=&page=1&size=20&sort=desc`
- **描述**: 获取兑换码列表
- **认证**: 是（管理员）
- **响应**: 分页响应

### 批量生成兑换码（管理员）
- **URL**: `POST /api/v1/codes/generate`
- **请求体**:
```json
{
  "prefix": "VIP_",
  "length": 8,
  "count": 100,
  "validity_days": 30,
  "max_uses": 1,
  "char_type": "alphanum",
  "tags": ["campaign"],
  "note": "2026春季活动",
  "source": "admin"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      { "code": "VIP_XXXX", "status": "active" }
    ]
  }
}
```

### 批量更新兑换码（管理员）
- **URL**: `PUT /api/v1/codes/batch`
- **请求体**:
```json
{
  "codes": ["VIP_XXXX"],
  "action": "disable",
  "value": 30
}
```
- **响应**: `success`

### 导出兑换码（管理员）
- **URL**: `GET /api/v1/codes/export?status=all&search=&sort=desc`
- **描述**: CSV 导出

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
    "check_in_streak": 0,
    "must_change_password": false,
    "ai_access_until": "2026-02-04T02:00:00Z"
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

### 修改密码
- **URL**: `PUT /api/v1/users/password`
- **描述**: 修改当前用户密码
- **认证**: 是
- **请求体**:
```json
{
  "new_password": "string (required, min:6, max:100)"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success"
}
```

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

### 用户心跳
- **URL**: `POST /api/v1/users/heartbeat`
- **描述**: 上报用户基础信息（不包含项目数据）
- **认证**: 是
- **请求体**:
```json
{
  "device_id": "string"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success"
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

### 同步项目快照
- **URL**: `POST /api/v1/projects/snapshot`
- **描述**: 同步本地项目快照到后端（用于持久化）
- **认证**: 是
- **请求体**:
```json
{
  "external_id": "string (required, max:64)",
  "title": "string (optional, max:200)",
  "ai_settings": {},
  "snapshot": {}
}
```
- **响应**: 项目详情

> 说明：线上纯网页模式请关闭项目同步，仅在本地/单机后端或允许云端存储时启用。

### 备份项目快照（本地文件）
- **URL**: `POST /api/v1/projects/backup`
- **描述**: 将项目快照写入本地文件存储，作为保底备份
- **认证**: 是
- **请求体**:
```json
{
  "external_id": "string (required, max:64)",
  "title": "string (optional, max:200)",
  "ai_settings": {},
  "snapshot": {}
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 1,
    "file_name": "project_xxx_20260204_170000.json",
    "storage_key": "backups/1/xxx/project_xxx_20260204_170000.json",
    "project_id": 1
  }
}
```

### 获取最新备份
- **URL**: `GET /api/v1/projects/:id/backup/latest`
- **描述**: 获取指定项目最新备份文件信息
- **认证**: 是
- **响应**: File 元信息

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

## 文件接口

### 获取项目文件列表
- **URL**: `GET /api/v1/files/project/:project_id`
- **描述**: 获取指定项目的文件列表
- **认证**: 是
- **请求参数**:
  - `page` 页码
  - `page_size` 每页数量
  - `file_type` 可选（如 backup）
- **响应（data）**:
```json
{
  "files": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

### 下载文件
- **URL**: `GET /api/v1/files/:id/download`
- **描述**: 下载指定文件
- **认证**: 是

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

## 会话接口（Sessions）

### 创建会话
- **URL**: `POST /api/v1/sessions`
- **描述**: 创建工作流会话
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (required, max:200)",
  "mode": "string (required, max:50)",
  "project_id": "uint (required)"
}
```
- **响应（data）**: Session

### 获取会话列表
- **URL**: `GET /api/v1/sessions?page=1&page_size=20`
- **描述**: 获取当前用户的会话列表（分页）
- **认证**: 是
- **响应（data）**:
```json
{
  "sessions": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

### 获取会话详情
- **URL**: `GET /api/v1/sessions/:session_id`
- **描述**: 获取会话详情
- **认证**: 是
- **响应（data）**: Session

### 更新会话
- **URL**: `PUT /api/v1/sessions/:session_id`
- **描述**: 更新会话（部分字段可选）
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (optional, max:200)",
  "mode": "string (optional, max:50)"
}
```
- **响应（data）**: Session

### 删除会话
- **URL**: `DELETE /api/v1/sessions/:session_id`
- **描述**: 删除会话
- **认证**: 是
- **响应**: 成功响应

---

## 工作流接口（Workflows）

### 世界观生成
- **URL**: `POST /api/v1/workflows/world`
- **描述**: 触发世界观生成工作流，自动创建会话与步骤，并通过 SSE 推送内容
- **认证**: 是（且需有效 AI 权限）
- **请求体**:
```json
{
  "project_id": 1,
  "session_id": 0,
  "title": "世界观生成",
  "step_title": "世界观生成",
  "provider": "gemini",
  "path": "v1beta/models/gemini-3-flash-preview:generateContent",
  "body": "{...}"
}
```
- **响应（data）**:
```json
{
  "session": {},
  "step": {},
  "content": "...",
  "raw": {}
}
```

### 章节润色
- **URL**: `POST /api/v1/workflows/polish`
- **描述**: 触发章节润色工作流，自动创建会话与步骤，并通过 SSE 推送内容
- **认证**: 是（且需有效 AI 权限）
- **请求体**:
```json
{
  "project_id": 1,
  "session_id": 0,
  "title": "章节润色",
  "step_title": "章节润色",
  "provider": "gemini",
  "path": "v1beta/models/gemini-3-flash-preview:generateContent",
  "body": "{...}"
}
```
- **响应（data）**:
```json
{
  "session": {},
  "step": {},
  "content": "...",
  "raw": {}
}
```

---

## 会话步骤接口（Session Steps）

### 创建步骤
- **URL**: `POST /api/v1/sessions/:session_id/steps`
- **描述**: 在指定会话下创建步骤
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (required, max:200)",
  "content": "string",
  "format_type": "string (max:50)",
  "order_index": "int"
}
```
- **响应（data）**: SessionStep

### 获取步骤列表
- **URL**: `GET /api/v1/sessions/:session_id/steps`
- **描述**: 获取指定会话下的步骤列表
- **认证**: 是
- **响应（data）**: SessionStep[]

### 获取步骤详情
- **URL**: `GET /api/v1/sessions/steps/:id`
- **描述**: 获取步骤详情
- **认证**: 是
- **响应（data）**: SessionStep

### 更新步骤
- **URL**: `PUT /api/v1/sessions/steps/:id`
- **描述**: 更新步骤（部分字段可选）
- **认证**: 是
- **请求体**:
```json
{
  "title": "string (optional, max:200)",
  "content": "string (optional)",
  "format_type": "string (optional, max:50)",
  "order_index": "int (optional)"
}
```
- **响应（data）**: SessionStep

### 删除步骤
- **URL**: `DELETE /api/v1/sessions/steps/:id`
- **描述**: 删除步骤
- **认证**: 是
- **响应**: 成功响应

---

## SSE 接口

### 订阅会话流
- **URL**: `GET /api/v1/sse/stream?session_id=1`
- **描述**: 订阅指定 session 的实时事件流（用于工作流进度/输出推送）
- **认证**: 是
- **响应**: `text/event-stream`
- **事件格式**:
```
event: <type>
data: <json>

```
- **事件类型（event: <type>）**:
  - `step.appended`
  - `quality.checked`
  - `export.ready`
  - `error`
- **data（JSON）结构**:
```json
{
  "type": "string",
  "data": {},
  "timestamp": "RFC3339 time"
}
```

### 广播测试事件
- **URL**: `POST /api/v1/sse/test?session_id=1&type=step`
- **描述**: 向指定 session 广播测试事件（用于联调/验证）
- **认证**: 是
- **说明**:
  - `type` 支持：`step` / `quality` / `export`
- **响应（data）**:
```json
{
  "message": "Event broadcasted",
  "session_id": "string",
  "event_type": "string"
}
```

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
