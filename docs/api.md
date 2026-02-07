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
- **描述**: 获取指定供应商模型列表（带缓存兜底机制）
- **认证**: 是
- **请求参数**:
  - `provider` string 必填（如: gemini/openai/proxy/local）
- **缓存策略**:
  - 优先返回有效缓存（默认 TTL: 3600秒）
  - 缓存过期时请求上游 API 并更新缓存
  - 上游失败时返回过期缓存（如果 `use_stale_cache_on_error=true`）
  - 无缓存且上游失败时返回错误
- **配置项**:
  - `ai.models_cache_ttl`: 缓存过期时间（秒），默认 3600
  - `ai.use_stale_cache_on_error`: 上游失败时是否使用过期缓存，默认 true
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
- **前端建议**:
  - 建议前端也实现本地缓存（localStorage），减少不必要的请求
  - 缓存键建议格式：`ai_models_${provider}_${timestamp}`
  - 前端缓存 TTL 可设置为 1800秒（30分钟）

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

## SSE 接口

### 建立 SSE 连接
- **URL**: `GET /api/v1/sse/stream?session_id=xxx`
- **描述**: 订阅指定会话的实时事件流（步骤追加/进度/完成/任务等）
- **认证**: 是（JWT）
- **响应**: `text/event-stream`

### SSE 事件类型

通用字段：

```json
{
  "type": "step.appended",
  "data": {},
  "timestamp": "2026-02-05T12:00:00Z"
}
```

- `step.appended`：会话步骤追加（data: step_id/title/content/timestamp）
- `step.chunk`：流式内容片段（data: session_id/step_id/chunk/is_final）
- `step.completed`：流式完成（data: session_id/step_id/content）
- `step.error`：流式错误（data: session_id/step_id/error）
- `progress.updated`：工作流进度更新（data: progress/message/timestamp）
- `workflow.done`：工作流完成（data: mode/document_id/timestamp）
- `job.*`：异步任务事件（job.created/job.started/job.progress/job.failed/job.succeeded/job.canceled）
- `error`：错误事件

---

## 工作流接口

### AgentWriter 写作代理

#### 启动写作任务
- **URL**: `POST /api/v1/agent-writer/start`
- **描述**: 启动 AgentWriter 写作任务，异步生成多个章节并实时推送进度
- **认证**: 是（且需有效 AI 权限）

请求体：
```json
{
  "project_id": 1,
  "document_id": 123,
  "prompt": "写一个科幻小说，主题是人工智能觉醒",
  "outline": [
    {
      "title": "第一章：觉醒",
      "description": "AI 系统首次产生自我意识"
    },
    {
      "title": "第二章：探索",
      "description": "AI 开始探索人类世界"
    }
  ],
  "provider": "gemini",
  "path": "v1beta/models/xxx:streamGenerateContent"
}
```

响应体：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 456,
    "status": "pending",
    "message": "写作任务已启动，请通过 SSE 监听进度"
  }
}
```

#### 取消写作任务
- **URL**: `POST /api/v1/agent-writer/cancel`
- **描述**: 取消正在执行的写作任务
- **认证**: 是

请求体：
```json
{
  "session_id": 456
}
```

响应体：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 456,
    "message": "写作任务已取消"
  }
}
```

#### 查询任务状态
- **URL**: `GET /api/v1/agent-writer/status/:session_id`
- **描述**: 查询写作任务的当前状态
- **认证**: 是

响应体：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 456,
    "workflow_type": "agent_writer",
    "workflow_status": "running",
    "workflow_config": {
      "project_id": 1,
      "document_id": 123,
      "prompt": "写一个科幻小说",
      "outline": [...],
      "current_chapter": 1,
      "total_chapters": 2,
      "provider": "gemini",
      "path": "v1beta/models/xxx:streamGenerateContent"
    },
    "title": "AgentWriter: 写一个科幻小说",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:05:00Z"
  }
}
```

#### SSE 事件类型

AgentWriter 通过 SSE 推送以下事件：

1. **chapter.start** - 章节开始生成
```json
{
  "type": "chapter.start",
  "data": {
    "session_id": 456,
    "chapter_index": 0,
    "chapter_title": "第一章：觉醒",
    "total_chapters": 2
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

2. **chapter.progress** - 章节生成进度（流式内容）
```json
{
  "type": "chapter.progress",
  "data": {
    "session_id": 456,
    "step_id": 789,
    "chapter_index": 0,
    "chunk": "在遥远的未来..."
  },
  "timestamp": "2024-01-01T00:00:01Z"
}
```

3. **chapter.completed** - 章节生成完成
```json
{
  "type": "chapter.completed",
  "data": {
    "session_id": 456,
    "chapter_index": 0,
    "chapter_title": "第一章：觉醒"
  },
  "timestamp": "2024-01-01T00:02:00Z"
}
```

4. **workflow.completed** - 整个工作流完成
```json
{
  "type": "workflow.completed",
  "data": {
    "session_id": 456,
    "total_chapters": 2,
    "document_id": 123
  },
  "timestamp": "2024-01-01T00:05:00Z"
}
```

5. **step.error** - 章节生成错误
```json
{
  "type": "step.error",
  "data": {
    "session_id": 456,
    "chapter_index": 1,
    "error": "AI 调用超时"
  },
  "timestamp": "2024-01-01T00:03:00Z"
}
```

#### 工作流程

1. 前端调用 `/api/v1/agent-writer/start` 启动任务
2. 后端创建 Session（workflow_type=agent_writer, workflow_status=pending）
3. 后端异步执行工作流，状态变为 running
4. 遍历 outline 中的每个章节：
   - 推送 `chapter.start` 事件
   - 调用 AI 流式生成章节内容
   - 每个 chunk 推送 `chapter.progress` 事件
   - 章节完成后推送 `chapter.completed` 事件
   - 自动保存到 Document
5. 所有章节完成后推送 `workflow.completed` 事件
6. 前端通过 `/api/v1/sse/stream?session_id=456` 监听所有事件

---

### 世界观生成
- **URL**: `POST /api/v1/workflows/world`
- **描述**: 运行世界观工作流（写入 session_steps 并支持 SSE 推送）
- **认证**: 是（且需有效 AI 权限）

### 向导·世界观
- **URL**: `POST /api/v1/workflows/wizard/world`
- **描述**: 向导分步：生成世界观蓝图（写入 session_steps 并支持 SSE 推送）
- **认证**: 是（且需有效 AI 权限）

### 向导·角色
- **URL**: `POST /api/v1/workflows/wizard/characters`
- **描述**: 向导分步：生成角色设定（写入 session_steps 并支持 SSE 推送）
- **认证**: 是（且需有效 AI 权限）

### 向导·大纲
- **URL**: `POST /api/v1/workflows/wizard/outline`
- **描述**: 向导分步：生成第一卷/第一章标题等大纲信息（写入 session_steps 并支持 SSE 推送）
- **认证**: 是（且需有效 AI 权限）

### 章节润色（旧接口）
- **URL**: `POST /api/v1/workflows/polish`
- **描述**: 运行润色工作流（写入 session_steps 并支持 SSE 推送）
- **认证**: 是（且需有效 AI 权限）

### 流式工作流执行
- **URL**: `POST /api/v1/workflows/stream`
- **描述**: 执行流式工作流，实时通过 SSE 推送生成内容
- **认证**: 是（且需有效 AI 权限）

请求体：
```json
{
  "session_id": 1,
  "step_title": "流式生成",
  "provider": "gemini",
  "path": "v1beta/models/xxx:streamGenerateContent",
  "body": "{...}"
}
```

响应体：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "step_id": 123,
    "session_id": 1,
    "message": "Stream started"
  }
}
```

说明：
- 接口立即返回 `step_id`，前端通过 SSE 监听 `session_id` 获取流式内容
- SSE 事件类型：
  - `step.chunk`：流式内容片段（data: {session_id, step_id, chunk, is_final}）
  - `step.completed`：流式完成（data: {session_id, step_id, content}）
  - `step.error`：流式错误（data: {session_id, step_id, error}）

### 章节生成
- **URL**: `POST /api/v1/workflows/chapters/generate`
- **描述**: 生成章节内容，并按 write_back 配置写回 documents 表
- **认证**: 是（且需有效 AI 权限）

请求体（示例）：
```json
{
  "project_id": 1,
  "session_id": 0,
  "document_id": 0,
  "volume_id": 1,
  "title": "第1章",
  "order_index": 1,
  "provider": "gemini",
  "path": "v1beta/models/xxx:generateContent",
  "body": "{...}",
  "write_back": { "set_status": "草稿", "set_summary": false }
}
```

### 章节分析
- **URL**: `POST /api/v1/workflows/chapters/analyze`
- **描述**: 分析章节内容，可按 write_back.set_summary 写回 documents.summary
- **认证**: 是（且需有效 AI 权限）

### 章节重写
- **URL**: `POST /api/v1/workflows/chapters/rewrite`
- **描述**: 重写章节内容，写回 documents.content
- **认证**: 是（且需有效 AI 权限）

### 批量生成章节
- **URL**: `POST /api/v1/workflows/chapters/batch`
- **描述**: 批量生成多章内容，按条目创建 documents，并通过 SSE 推送进度
- **认证**: 是（且需有效 AI 权限）

请求体补充字段：
- `items[].client_document_id`：前端本地章节 ID（字符串），用于后端回传精确映射

响应体补充字段：
- `results[]`：数组，包含 `client_document_id` 与对应生成的 `document`

---

## 插件接口

### 获取插件列表
- **URL**: `GET /api/v1/plugins?page=1&page_size=20`
- **描述**: 获取插件列表（包含 capabilities）
- **认证**: 是

### 创建插件
- **URL**: `POST /api/v1/plugins`
- **描述**: 创建插件记录
- **认证**: 是

### 启用 / 禁用插件
- **URL**: `PUT /api/v1/plugins/:plugin_id/enable`
- **URL**: `PUT /api/v1/plugins/:plugin_id/disable`
- **描述**: 切换插件启用状态
- **认证**: 是

### 添加插件能力
- **URL**: `POST /api/v1/plugins/:plugin_id/capabilities`
- **描述**: 为插件添加能力（用于工具注入 / tool_calls）
- **认证**: 是

请求体（示例）：
```json
{
  "cap_id": "extract_entities",
  "name": "提取实体",
  "type": "data_provider",
  "description": "从章节文本中提取人物/地点/组织",
  "icon": "",
  "input_schema": {"type":"object","properties":{"text":{"type":"string"}},"required":["text"]},
  "output_schema": {"type":"object","properties":{"entities":{"type":"array"}}}
}
```

### 同步调用插件
- **URL**: `POST /api/v1/plugins/:plugin_id/invoke`
- **描述**: 同步调用插件 /invoke
- **认证**: 是

### 异步调用插件（Job）
- **URL**: `POST /api/v1/plugins/:plugin_id/invoke/async`
- **描述**: 创建一个异步 Job 运行插件，结果会写入 jobs 表，并通过 SSE 推送（同时会追加 session_steps）
- **认证**: 是

---

## Job 接口

### 获取 Job
- **URL**: `GET /api/v1/jobs/:job_uuid`
- **描述**: 获取异步任务状态与结果
- **认证**: 是

### 取消 Job
- **URL**: `POST /api/v1/jobs/:job_uuid/cancel`
- **描述**: 取消异步任务
- **认证**: 是

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

### Function Calling 工作流
- **URL**: `POST /api/v1/workflows/function-calling`
- **描述**: 执行 Function Calling 多轮对话工作流，支持工具调用和结果回写
- **认证**: 是（且需有效 AI 权限）
- **请求体**:
```json
{
  "session_id": 1,
  "prompt": "用户输入的初始提示词",
  "max_turns": 5,
  "tools": [
    {
      "name": "search",
      "description": "搜索工具",
      "parameters": {
        "type": "object",
        "properties": {
          "query": {
            "type": "string",
            "description": "搜索关键词"
          }
        },
        "required": ["query"]
      }
    }
  ]
}
```
- **响应（data）**:
```json
{
  "session_id": 1,
  "message": "Function calling completed"
}
```

**说明**:
- `session_id`: 会话 ID（必填）
- `prompt`: 用户输入的初始提示词（必填）
- `max_turns`: 最大对话轮数，默认 5（可选）
- `tools`: 可用工具列表（可选）

**工作流程**:
1. 用户输入 → AI 调用（返回 tool_calls）
2. 解析 tool_calls → 创建 Jobs
3. 等待 Jobs 完成 → 收集 tool_results
4. 构建新提示词（包含 tool_results）→ AI 调用
5. 重复 2-4，直到：
   - AI 不再返回 tool_calls（正常结束）
   - 达到 max_turns（防止无限循环）
   - 发生错误

**SessionStep 类型**:
- `user`: 用户输入
- `assistant`: AI 响应（文本）
- `tool_call`: AI 请求调用工具
- `tool_result`: 工具执行结果

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
