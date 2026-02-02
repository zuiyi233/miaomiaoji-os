# 第二阶段实施方案 - 后端架构

## 1. 第二阶段目标与范围（In/Out）

### 1.1 目标
在现有 Go + Gin + GORM 后端基础上，补齐业务能力（会话工作流、SSE、排版/质量门禁、语料、文件、结算、插件等），并把系统提升到"可生产交付"的状态。

### 1.2 范围内（13个核心任务模块）
1. 卷管理（Volumes）
2. 文档/章节管理（Documents）
3. 实体/世界观卡管理（Entities）
4. AI模板管理（Templates）
5. 插件管理（Plugins）
6. 会话工作流（Sessions）
7. SSE实时推送（Server-Sent Events）
8. 结算系统（Settlements）
9. 语料库管理（Corpus）
10. 文件管理（Files）
11. 排版引擎（Formatting）
12. 质量门禁（Quality Gate）
13. 项目导出（Project Export）

### 1.3 范围外
- 前端开发（React/TypeScript）
- AI模型训练
- 第三方支付集成
- AI模型调用（由插件系统提供）

---

## 2. 现状盘点

### 2.1 已实现的功能清单
- **认证模块**：用户注册、登录、退出、Token刷新
- **用户模块**：获取/更新用户信息、每日签到、积分管理、用户列表（管理员）
- **项目模块**：项目CRUD、项目列表、项目导出（TODO）
- **健康检查**：/healthz、/ready
- **中间件**：JWT认证、CORS、日志、请求ID、限流
- **公共能力**：统一响应、错误码、日志、配置管理

### 2.2 待实现的模块
- **Volumes**：卷管理（Repository、Service、Handler、Router）
- **Documents**：文档/章节管理（Repository、Service、Handler、Router）
- **Entities**：实体/世界观卡管理（Repository、Service、Handler、Router）
- **Templates**：AI模板管理（Repository、Service、Handler、Router）
- **Plugins**：插件管理（Repository、Service、Handler、Router）
- **Sessions**：会话工作流（Repository、Service、Handler、Router）
- **Settlements**：结算系统（Repository、Service、Handler、Router）
- **Corpus**：语料库管理（Repository、Service、Handler、Router）
- **Files**：文件管理（Repository、Service、Handler、Router）
- **Formatting**：排版引擎（Service）
- **Quality Gate**：质量门禁（Service）
- **SSE**：实时推送（Handler、Middleware）

### 2.3 技术债务
- **ProjectService.GetByIDWithDetails**：TODO注释，需要加载关联数据（volumes、documents、entities、templates）
- **ProjectHandler.Export**：TODO注释，需要实现项目导出为JSON格式
- **错误码扩展**：需要为各模块添加专用错误码（40xxx会话、50xxx插件等）

---

### 2.4 接口响应规范
本项目接口默认采用统一响应包装（见 pkg/response/response.go）。下文各模块接口表中的“响应（data）”指 `data` 字段的结构（除 SSE/文件流等流式接口外）。

**统一响应（JSON）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

**分页响应（JSON，data 为 PageResponse）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page_info": {
      "page": 1,
      "size": 20,
      "total": 0
    }
  }
}
```

**无数据成功响应（JSON）**：
```json
{
  "code": 0,
  "message": "success"
}
```

## 3. 模块级任务拆解

### 任务1：卷管理（Volumes）

#### 3.1.1 范围说明
实现卷的CRUD操作，支持按项目查询、排序、软删除。

#### 3.1.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/projects/:project_id/volumes | CreateVolumeRequest | VolumeResponse |
| GET | /api/v1/projects/:project_id/volumes | - | PageResponse(list: VolumeResponse[], page_info: PageInfo) |
| GET | /api/v1/volumes/:id | - | VolumeResponse |
| PUT | /api/v1/volumes/:id | UpdateVolumeRequest | VolumeResponse |
| DELETE | /api/v1/volumes/:id | - | Success |

**CreateVolumeRequest**:
```json
{
  "title": "string (required, max:200)",
  "order_index": "int (required)",
  "theme": "string",
  "core_goal": "string",
  "boundaries": "string",
  "chapter_linkage_logic": "string",
  "volume_specific_settings": "string",
  "plot_roadmap": "string"
}
```

**UpdateVolumeRequest**:
```json
{
  "title": "string (optional, max:200)",
  "order_index": "int (optional)",
  "theme": "string",
  "core_goal": "string",
  "boundaries": "string",
  "chapter_linkage_logic": "string",
  "volume_specific_settings": "string",
  "plot_roadmap": "string"
}
```

**VolumeResponse**:
```json
{
  "id": 1,
  "title": "string",
  "order_index": 1,
  "theme": "string",
  "core_goal": "string",
  "boundaries": "string",
  "chapter_linkage_logic": "string",
  "volume_specific_settings": "string",
  "plot_roadmap": "string",
  "project_id": 1,
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.1.3 数据表/模型
**表名**：volumes

**字段定义**：
- id (uint, PK, auto_increment)
- title (string, 200, not null)
- order_index (int, indexed)
- theme (text)
- core_goal (text)
- boundaries (text)
- chapter_linkage_logic (text)
- volume_specific_settings (text)
- plot_roadmap (text)
- project_id (uint, indexed, not null, FK→projects.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (order_index)
- INDEX (project_id)

**关联**：
- project_id → projects.id (Many-to-One)
- documents (One-to-Many)

#### 3.1.4 关键实现点
- **Repository层**：
  - Create(volume *model.Volume) error
  - FindByID(id uint) (*model.Volume, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error)
  - Update(volume *model.Volume) error
  - Delete(id uint) error
  - FindByProjectIDWithDocuments(projectID uint) ([]*model.Volume, error)

- **Service层**：
  - Create(projectID uint, title string, orderIndex int, theme, coreGoal, boundaries, chapterLinkageLogic, volumeSpecificSettings, plotRoadmap string) (*model.Volume, error)
  - GetByID(id uint) (*model.Volume, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.Volume, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Volume, error)
  - Delete(id uint) error

- **Handler层**：
  - Create(c *gin.Context)
  - List(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)

#### 3.1.5 幂等/并发控制点
- **创建卷**：使用 project_id + order_index 唯一约束防止重复
- **更新卷**：默认最后写入覆盖；可选基于 updated_at 做冲突检测
- **删除卷**：软删除，不级联删除关联文档

#### 3.1.6 安全控制点
- **权限验证**：用户只能操作自己项目的卷
- **参数验证**：title必填，order_index必填
- **SQL注入防护**：使用GORM参数化查询

#### 3.1.7 完成标准（DoD）
- [ ] VolumeRepository实现完整
- [ ] VolumeService实现完整
- [ ] VolumeHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.1.8 对 docs/api.md 的变更点
在"项目接口"章节后添加"卷接口"章节，包含上述5个接口的完整文档。

---

### 任务2：文档/章节管理（Documents）

#### 3.2.1 范围说明
实现文档/章节的CRUD操作，支持按项目/卷查询、排序、书签管理、实体关联。

#### 3.2.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/projects/:project_id/documents | CreateDocumentRequest | DocumentResponse |
| GET | /api/v1/projects/:project_id/documents | - | PageResponse(list: DocumentResponse[], page_info: PageInfo) |
| GET | /api/v1/volumes/:volume_id/documents | - | PageResponse(list: DocumentResponse[], page_info: PageInfo) |
| GET | /api/v1/documents/:id | - | DocumentResponse |
| PUT | /api/v1/documents/:id | UpdateDocumentRequest | DocumentResponse |
| DELETE | /api/v1/documents/:id | - | Success |
| POST | /api/v1/documents/:id/bookmarks | AddBookmarkRequest | DocumentResponse |
| DELETE | /api/v1/documents/:id/bookmarks/:index | - | DocumentResponse |
| POST | /api/v1/documents/:id/entities | LinkEntityRequest | DocumentResponse |
| DELETE | /api/v1/documents/:id/entities/:entity_id | - | DocumentResponse |

**CreateDocumentRequest**:
```json
{
  "title": "string (required, max:200)",
  "content": "string",
  "summary": "string",
  "status": "string (default:草稿)",
  "order_index": "int (required)",
  "time_node": "string",
  "duration": "string",
  "target_word_count": "int",
  "chapter_goal": "string",
  "core_plot": "string",
  "hook": "string",
  "cause_effect": "string",
  "foreshadowing_details": "string",
  "volume_id": "int"
}
```

**UpdateDocumentRequest**:
```json
{
  "title": "string (optional, max:200)",
  "content": "string",
  "summary": "string",
  "status": "string",
  "order_index": "int",
  "time_node": "string",
  "duration": "string",
  "target_word_count": "int",
  "chapter_goal": "string",
  "core_plot": "string",
  "hook": "string",
  "cause_effect": "string",
  "foreshadowing_details": "string",
  "volume_id": "int"
}
```

**AddBookmarkRequest**:
```json
{
  "title": "string (required)",
  "position": "int (required)",
  "note": "string"
}
```

**LinkEntityRequest**:
```json
{
  "entity_id": "int (required)",
  "ref_type": "string (default:mention)",
  "metadata": {}
}
```

**DocumentResponse**:
```json
{
  "id": 1,
  "title": "string",
  "content": "string",
  "summary": "string",
  "status": "string",
  "order_index": 1,
  "bookmarks": [],
  "time_node": "string",
  "duration": "string",
  "target_word_count": 2250,
  "chapter_goal": "string",
  "core_plot": "string",
  "hook": "string",
  "cause_effect": "string",
  "foreshadowing_details": "string",
  "project_id": 1,
  "volume_id": 1,
  "entity_refs": [],
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.2.3 数据表/模型
**表名**：documents

**字段定义**：
- id (uint, PK, auto_increment)
- title (string, 200, not null)
- content (text)
- summary (text)
- status (string, 20, default:草稿)
- order_index (int, indexed)
- bookmarks (JSON)
- time_node (string, 100)
- duration (string, 50)
- target_word_count (int)
- chapter_goal (text)
- core_plot (text)
- hook (text)
- cause_effect (text)
- foreshadowing_details (text)
- project_id (uint, indexed, not null, FK→projects.id)
- volume_id (uint, indexed, FK→volumes.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (order_index)
- INDEX (project_id)
- INDEX (volume_id)

**关联表**：document_entity_refs
- id (uint, PK)
- document_id (uint, indexed, not null, FK→documents.id)
- entity_id (uint, indexed, not null, FK→entities.id)
- ref_type (string, 20, default:mention)
- metadata (JSON)
- created_at (datetime)
- updated_at (datetime)

**索引**：
- PRIMARY KEY (id)
- INDEX (document_id)
- INDEX (entity_id)
- UNIQUE INDEX (document_id, entity_id)

#### 3.2.4 关键实现点
- **Repository层**：
  - Create(document *model.Document) error
  - FindByID(id uint) (*model.Document, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error)
  - FindByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error)
  - Update(document *model.Document) error
  - Delete(id uint) error
  - AddBookmark(documentID uint, bookmark model.Bookmark) error
  - RemoveBookmark(documentID uint, index int) error
  - LinkEntity(documentID, entityID uint, refType string, metadata map[string]interface{}) error
  - UnlinkEntity(documentID, entityID uint) error

- **Service层**：
  - Create(projectID uint, title string, content, summary, status string, orderIndex int, timeNode, duration string, targetWordCount int, chapterGoal, corePlot, hook, causeEffect, foreshadowingDetails string, volumeID uint) (*model.Document, error)
  - GetByID(id uint) (*model.Document, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.Document, int64, error)
  - ListByVolumeID(volumeID uint, page, size int) ([]*model.Document, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Document, error)
  - Delete(id uint) error
  - AddBookmark(id uint, title string, position int, note string) error
  - RemoveBookmark(id uint, index int) error
  - LinkEntity(id, entityID uint, refType string, metadata map[string]interface{}) error
  - UnlinkEntity(id, entityID uint) error

- **Handler层**：
  - Create(c *gin.Context)
  - ListByProject(c *gin.Context)
  - ListByVolume(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)
  - AddBookmark(c *gin.Context)
  - RemoveBookmark(c *gin.Context)
  - LinkEntity(c *gin.Context)
  - UnlinkEntity(c *gin.Context)

#### 3.2.5 幂等/并发控制点
- **创建文档**：使用 project_id + order_index 唯一约束防止重复
- **更新文档**：默认最后写入覆盖；可选基于 updated_at 做冲突检测
- **添加书签**：使用 document_id + position 唯一约束
- **关联实体**：使用 document_id + entity_id 唯一约束

#### 3.2.6 安全控制点
- **权限验证**：用户只能操作自己项目的文档
- **参数验证**：title必填，order_index必填
- **SQL注入防护**：使用GORM参数化查询
- **XSS防护**：content字段需要转义（前端处理）

#### 3.2.7 完成标准（DoD）
- [ ] DocumentRepository实现完整
- [ ] DocumentService实现完整
- [ ] DocumentHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.2.8 对 docs/api.md 的变更点
在"卷接口"章节后添加"文档接口"章节，包含上述10个接口的完整文档。

---

### 任务3：实体/世界观卡管理（Entities）

#### 3.3.1 范围说明
实现实体/世界观卡的CRUD操作，支持按项目查询、标签管理、实体关联、引用计数。

#### 3.3.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/projects/:project_id/entities | CreateEntityRequest | EntityResponse |
| GET | /api/v1/projects/:project_id/entities | - | PageResponse(list: EntityResponse[], page_info: PageInfo) |
| GET | /api/v1/entities/:id | - | EntityResponse |
| PUT | /api/v1/entities/:id | UpdateEntityRequest | EntityResponse |
| DELETE | /api/v1/entities/:id | - | Success |
| POST | /api/v1/entities/:id/tags | AddTagRequest | EntityResponse |
| DELETE | /api/v1/entities/:id/tags/:tag | - | EntityResponse |
| POST | /api/v1/entities/:id/links | CreateLinkRequest | EntityResponse |
| DELETE | /api/v1/entities/:id/links/:target_id | - | EntityResponse |

**CreateEntityRequest**:
```json
{
  "entity_type": "string (required, enum:character/setting/organization/item/magic/event)",
  "title": "string (required, max:200)",
  "subtitle": "string",
  "content": "string",
  "voice_style": "string",
  "importance": "string (default:secondary, enum:main/secondary/minor)",
  "custom_fields": []
}
```

**UpdateEntityRequest**:
```json
{
  "entity_type": "string",
  "title": "string",
  "subtitle": "string",
  "content": "string",
  "voice_style": "string",
  "importance": "string",
  "custom_fields": []
}
```

**AddTagRequest**:
```json
{
  "tag": "string (required, max:50)"
}
```

**CreateLinkRequest**:
```json
{
  "target_id": "int (required)",
  "type": "string",
  "relation_name": "string"
}
```

**EntityResponse**:
```json
{
  "id": 1,
  "entity_type": "string",
  "title": "string",
  "subtitle": "string",
  "content": "string",
  "voice_style": "string",
  "importance": "string",
  "custom_fields": [],
  "reference_count": 0,
  "project_id": 1,
  "tags": [],
  "links": [],
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.3.3 数据表/模型
**表名**：entities

**字段定义**：
- id (uint, PK, auto_increment)
- entity_type (string, 20, not null, indexed)
- title (string, 200, not null)
- subtitle (string, 200)
- content (text)
- voice_style (string, 100)
- importance (string, 20, default:secondary)
- custom_fields (JSON)
- reference_count (int, default:0)
- project_id (uint, indexed, not null, FK→projects.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (entity_type)
- INDEX (project_id)

**关联表**：entity_tags
- id (uint, PK)
- entity_id (uint, indexed, not null, FK→entities.id)
- tag (string, 50, indexed, not null)
- created_at (datetime)

**索引**：
- PRIMARY KEY (id)
- INDEX (entity_id)
- INDEX (tag)
- UNIQUE INDEX (entity_id, tag)

**关联表**：entity_links
- id (uint, PK)
- source_id (uint, indexed, not null, FK→entities.id)
- target_id (uint, indexed, not null, FK→entities.id)
- type (string, 20)
- relation_name (string, 50)
- created_at (datetime)
- updated_at (datetime)

**索引**：
- PRIMARY KEY (id)
- INDEX (source_id)
- INDEX (target_id)
- UNIQUE INDEX (source_id, target_id)

#### 3.3.4 关键实现点
- **Repository层**：
  - Create(entity *model.Entity) error
  - FindByID(id uint) (*model.Entity, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error)
  - FindByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error)
  - Update(entity *model.Entity) error
  - Delete(id uint) error
  - AddTag(entityID uint, tag string) error
  - RemoveTag(entityID uint, tag string) error
  - CreateLink(sourceID, targetID uint, linkType, relationName string) error
  - DeleteLink(sourceID, targetID uint) error
  - IncrementReferenceCount(entityID uint) error
  - DecrementReferenceCount(entityID uint) error

- **Service层**：
  - Create(projectID uint, entityType, title, subtitle, content, voiceStyle, importance string, customFields []model.EntityCustomField) (*model.Entity, error)
  - GetByID(id uint) (*model.Entity, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.Entity, int64, error)
  - ListByType(projectID uint, entityType string, page, size int) ([]*model.Entity, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Entity, error)
  - Delete(id uint) error
  - AddTag(id uint, tag string) error
  - RemoveTag(id uint, tag string) error
  - CreateLink(id, targetID uint, linkType, relationName string) error
  - DeleteLink(id, targetID uint) error

- **Handler层**：
  - Create(c *gin.Context)
  - List(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)
  - AddTag(c *gin.Context)
  - RemoveTag(c *gin.Context)
  - CreateLink(c *gin.Context)
  - DeleteLink(c *gin.Context)

#### 3.3.5 幂等/并发控制点
- **创建实体**：使用 project_id + entity_type + title 唯一约束防止重复
- **更新实体**：默认最后写入覆盖；可选基于 updated_at 做冲突检测
- **添加标签**：使用 entity_id + tag 唯一约束
- **创建关联**：使用 source_id + target_id 唯一约束
- **引用计数**：使用原子操作或数据库触发器

#### 3.3.6 安全控制点
- **权限验证**：用户只能操作自己项目的实体
- **参数验证**：entity_type必填且在枚举范围内，title必填
- **SQL注入防护**：使用GORM参数化查询
- **循环引用检测**：创建关联时检测是否形成循环

#### 3.3.7 完成标准（DoD）
- [ ] EntityRepository实现完整
- [ ] EntityService实现完整
- [ ] EntityHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.3.8 对 docs/api.md 的变更点
在"文档接口"章节后添加"实体接口"章节，包含上述9个接口的完整文档。

---

### 任务4：AI模板管理（Templates）

#### 3.4.1 范围说明
实现AI模板的CRUD操作，支持按项目查询、分类管理、系统模板与项目模板。

#### 3.4.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/projects/:project_id/templates | CreateTemplateRequest | TemplateResponse |
| GET | /api/v1/projects/:project_id/templates | - | PageResponse(list: TemplateResponse[], page_info: PageInfo) |
| GET | /api/v1/templates/system | - | PageResponse(list: TemplateResponse[], page_info: PageInfo) |
| GET | /api/v1/templates/:id | - | TemplateResponse |
| PUT | /api/v1/templates/:id | UpdateTemplateRequest | TemplateResponse |
| DELETE | /api/v1/templates/:id | - | Success |

**CreateTemplateRequest**:
```json
{
  "name": "string (required, max:100)",
  "description": "string",
  "category": "string (enum:logic/style/content/character)",
  "template": "string (required)"
}
```

**UpdateTemplateRequest**:
```json
{
  "name": "string",
  "description": "string",
  "category": "string",
  "template": "string"
}
```

**TemplateResponse**:
```json
{
  "id": 1,
  "name": "string",
  "description": "string",
  "category": "string",
  "template": "string",
  "project_id": 1,
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.4.3 数据表/模型
**表名**：templates

**字段定义**：
- id (uint, PK, auto_increment)
- name (string, 100, not null)
- description (text)
- category (string, 20)
- template (text, not null)
- project_id (uint, indexed, 0表示系统模板)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (project_id)
- INDEX (category)

#### 3.4.4 关键实现点
- **Repository层**：
  - Create(template *model.Template) error
  - FindByID(id uint) (*model.Template, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error)
  - FindSystemTemplates(page, size int) ([]*model.Template, int64, error)
  - FindByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error)
  - Update(template *model.Template) error
  - Delete(id uint) error

- **Service层**：
  - Create(projectID uint, name, description, category, template string) (*model.Template, error)
  - GetByID(id uint) (*model.Template, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.Template, int64, error)
  - ListSystemTemplates(page, size int) ([]*model.Template, int64, error)
  - ListByCategory(projectID uint, category string, page, size int) ([]*model.Template, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Template, error)
  - Delete(id uint) error

- **Handler层**：
  - Create(c *gin.Context)
  - ListByProject(c *gin.Context)
  - ListSystem(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)

#### 3.4.5 幂等/并发控制点
- **创建模板**：使用 project_id + name 唯一约束防止重复
- **更新模板**：默认最后写入覆盖；可选基于 updated_at 做冲突检测

#### 3.4.6 安全控制点
- **权限验证**：用户只能操作自己项目的模板，系统模板只能管理员操作
- **参数验证**：name必填，template必填，category在枚举范围内
- **SQL注入防护**：使用GORM参数化查询

#### 3.4.7 完成标准（DoD）
- [ ] TemplateRepository实现完整
- [ ] TemplateService实现完整
- [ ] TemplateHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.4.8 对 docs/api.md 的变更点
在"实体接口"章节后添加"模板接口"章节，包含上述6个接口的完整文档。

---

### 任务5：插件管理（Plugins）

#### 3.5.1 范围说明
实现插件的注册、发现、健康检查、能力查询、调用管理。

#### 3.5.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/plugins | RegisterPluginRequest | PluginResponse |
| GET | /api/v1/plugins | - | PageResponse(list: PluginResponse[], page_info: PageInfo) |
| GET | /api/v1/plugins/:id | - | PluginResponse |
| PUT | /api/v1/plugins/:id | UpdatePluginRequest | PluginResponse |
| DELETE | /api/v1/plugins/:id | - | Success |
| POST | /api/v1/plugins/:id/health-check | - | HealthCheckResponse |
| GET | /api/v1/plugins/:id/capabilities | - | CapabilitiesResponse |
| POST | /api/v1/plugins/:id/invoke | InvokePluginRequest | InvokePluginResponse |

**RegisterPluginRequest**:
```json
{
  "name": "string (required, max:100)",
  "version": "string",
  "author": "string",
  "description": "string",
  "endpoint": "string (required, max:500)",
  "config": {}
}
```

**UpdatePluginRequest**:
```json
{
  "name": "string",
  "version": "string",
  "author": "string",
  "description": "string",
  "endpoint": "string",
  "is_enabled": "bool",
  "config": {}
}
```

**HealthCheckResponse**:
```json
{
  "status": "string (enum:online/offline/error/unknown)",
  "latency_ms": 100,
  "last_ping": "2024-01-01 00:00:00"
}
```

**CapabilitiesResponse**:
```json
{
  "capabilities": [
    {
      "id": 1,
      "cap_id": "string",
      "name": "string",
      "type": "string",
      "description": "string",
      "icon": "string"
    }
  ]
}
```

**InvokePluginRequest**:
```json
{
  "cap_id": "string (required)",
  "input": {}
}
```

**InvokePluginResponse**:
```json
{
  "output": {},
  "error": "string"
}
```

**PluginResponse**:
```json
{
  "id": 1,
  "name": "string",
  "version": "string",
  "author": "string",
  "description": "string",
  "endpoint": "string",
  "is_enabled": true,
  "status": "string",
  "last_ping": "2024-01-01 00:00:00",
  "latency_ms": 100,
  "config": {},
  "capabilities": [],
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.5.3 数据表/模型
**表名**：plugins

**字段定义**：
- id (uint, PK, auto_increment)
- name (string, 100, not null)
- version (string, 20)
- author (string, 100)
- description (text)
- endpoint (string, 500, not null)
- is_enabled (bool, default:true)
- status (string, 20, default:unknown)
- last_ping (datetime, nullable)
- latency_ms (int)
- config (JSON)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (is_enabled)
- INDEX (status)

**关联表**：plugin_capabilities
- id (uint, PK)
- plugin_id (uint, indexed, not null, FK→plugins.id)
- cap_id (string, 50)
- name (string, 100)
- type (string, 30)
- description (text)
- icon (string, 100)
- created_at (datetime)
- updated_at (datetime)

**索引**：
- PRIMARY KEY (id)
- INDEX (plugin_id)
- INDEX (cap_id)

#### 3.5.4 关键实现点
- **Repository层**：
  - Create(plugin *model.Plugin) error
  - FindByID(id uint) (*model.Plugin, error)
  - FindAll(page, size int) ([]*model.Plugin, int64, error)
  - FindEnabled(page, size int) ([]*model.Plugin, int64, error)
  - Update(plugin *model.Plugin) error
  - Delete(id uint) error
  - UpdateStatus(id uint, status string, latencyMs int, lastPing time.Time) error
  - FindCapabilities(pluginID uint) ([]*model.PluginCapability, error)
  - CreateCapability(capability *model.PluginCapability) error
  - DeleteCapabilities(pluginID uint) error

- **Service层**：
  - Register(name, version, author, description, endpoint string, config map[string]interface{}) (*model.Plugin, error)
  - GetByID(id uint) (*model.Plugin, error)
  - List(page, size int) ([]*model.Plugin, int64, error)
  - ListEnabled(page, size int) ([]*model.Plugin, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Plugin, error)
  - Delete(id uint) error
  - HealthCheck(id uint) (*model.Plugin, error)
  - GetCapabilities(id uint) ([]*model.PluginCapability, error)
  - Invoke(id uint, capID string, input map[string]interface{}) (map[string]interface{}, error)

- **Handler层**：
  - Register(c *gin.Context)
  - List(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)
  - HealthCheck(c *gin.Context)
  - GetCapabilities(c *gin.Context)
  - Invoke(c *gin.Context)

- **HTTP客户端**：
  - 创建 pkg/httpclient 包，封装插件调用逻辑
  - 实现超时控制、重试机制、错误处理

#### 3.5.5 幂等/并发控制点
- **注册插件**：使用 endpoint 唯一约束防止重复
- **更新插件**：默认最后写入覆盖；可选基于 updated_at 做冲突检测
- **健康检查**：使用分布式锁防止并发检查
- **调用插件**：使用请求ID追踪，支持重试

#### 3.5.6 安全控制点
- **权限验证**：插件注册、更新、删除需要管理员权限
- **参数验证**：name必填，endpoint必填且为有效URL
- **SQL注入防护**：使用GORM参数化查询
- **插件调用安全**：
  - 验证插件状态（is_enabled=true, status=online）
  - 验证能力是否存在
  - 设置调用超时（从配置读取）
  - 记录调用日志

#### 3.5.7 完成标准（DoD）
- [ ] PluginRepository实现完整
- [ ] PluginService实现完整
- [ ] PluginHandler实现完整
- [ ] pkg/httpclient包实现
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.5.8 对 docs/api.md 的变更点
在"模板接口"章节后添加"插件接口"章节，包含上述8个接口的完整文档。

---

### 任务6：会话工作流（Sessions）

#### 3.6.1 范围说明
实现会话工作流的创建、管理、步骤记录、状态追踪，支持4种模式（Normal/Fusion/Single/Batch）。

#### 3.6.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/projects/:project_id/sessions | CreateSessionRequest | SessionResponse |
| GET | /api/v1/projects/:project_id/sessions | - | PageResponse(list: SessionResponse[], page_info: PageInfo) |
| GET | /api/v1/sessions/:id | - | SessionResponse |
| PUT | /api/v1/sessions/:id | UpdateSessionRequest | SessionResponse |
| DELETE | /api/v1/sessions/:id | - | Success |
| POST | /api/v1/sessions/:id/steps | CreateStepRequest | StepResponse |
| GET | /api/v1/sessions/:id/steps | - | PageResponse(list: StepResponse[], page_info: PageInfo) |
| PUT | /api/v1/sessions/:id/steps/:step_id | UpdateStepRequest | StepResponse |
| DELETE | /api/v1/sessions/:id/steps/:step_id | - | Success |

**CreateSessionRequest**:
```json
{
  "title": "string (required, max:200)",
  "mode": "string (required, enum:Normal/Fusion/Single/Batch)"
}
```

**UpdateSessionRequest**:
```json
{
  "title": "string"
}
```

**CreateStepRequest**:
```json
{
  "title": "string (required, max:200)",
  "content": "string",
  "format_type": "string",
  "order_index": "int (required)",
  "metadata": {}
}
```

**UpdateStepRequest**:
```json
{
  "title": "string",
  "content": "string",
  "format_type": "string",
  "metadata": {}
}
```

**SessionResponse**:
```json
{
  "id": 1,
  "title": "string",
  "mode": "string",
  "project_id": 1,
  "user_id": 1,
  "steps": [],
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

**StepResponse**:
```json
{
  "id": 1,
  "title": "string",
  "content": "string",
  "format_type": "string",
  "order_index": 1,
  "metadata": {},
  "session_id": 1,
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.6.3 数据表/模型
**表名**：sessions

**字段定义**：
- id (uint, PK, auto_increment)
- title (string, 200)
- mode (string, 50)
- project_id (uint, indexed, not null, FK→projects.id)
- user_id (uint, indexed, not null, FK→users.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (project_id)
- INDEX (user_id)

**关联表**：session_steps
- id (uint, PK, auto_increment)
- title (string, 200)
- content (text)
- format_type (string, 50)
- order_index (int, indexed)
- metadata (JSON)
- session_id (uint, indexed, not null, FK→sessions.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (order_index)
- INDEX (session_id)

#### 3.6.4 关键实现点
- **Repository层**：
  - Create(session *model.Session) error
  - FindByID(id uint) (*model.Session, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.Session, int64, error)
  - FindByUserID(userID uint, page, size int) ([]*model.Session, int64, error)
  - Update(session *model.Session) error
  - Delete(id uint) error
  - CreateStep(step *model.SessionStep) error
  - FindStepsBySessionID(sessionID uint, page, size int) ([]*model.SessionStep, int64, error)
  - UpdateStep(step *model.SessionStep) error
  - DeleteStep(id uint) error

- **Service层**：
  - Create(projectID, userID uint, title, mode string) (*model.Session, error)
  - GetByID(id uint) (*model.Session, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.Session, int64, error)
  - ListByUserID(userID uint, page, size int) ([]*model.Session, int64, error)
  - Update(id uint, updates map[string]interface{}) (*model.Session, error)
  - Delete(id uint) error
  - CreateStep(sessionID uint, title, content, formatType string, orderIndex int, metadata map[string]interface{}) (*model.SessionStep, error)
  - ListSteps(sessionID uint, page, size int) ([]*model.SessionStep, int64, error)
  - UpdateStep(id uint, updates map[string]interface{}) (*model.SessionStep, error)
  - DeleteStep(id uint) error

- **Handler层**：
  - Create(c *gin.Context)
  - List(c *gin.Context)
  - GetByID(c *gin.Context)
  - Update(c *gin.Context)
  - Delete(c *gin.Context)
  - CreateStep(c *gin.Context)
  - ListSteps(c *gin.Context)
  - UpdateStep(c *gin.Context)
  - DeleteStep(c *gin.Context)

#### 3.6.5 幂等/并发控制点
- **创建会话**：使用 project_id + user_id + title 唯一约束防止重复
- **创建步骤**：使用 session_id + order_index 唯一约束防止重复
- **更新会话/步骤**：默认最后写入覆盖；可选基于 updated_at 做冲突检测

#### 3.6.6 安全控制点
- **权限验证**：用户只能操作自己的会话
- **参数验证**：mode必填且在枚举范围内
- **SQL注入防护**：使用GORM参数化查询

#### 3.6.7 完成标准（DoD）
- [ ] SessionRepository实现完整
- [ ] SessionService实现完整
- [ ] SessionHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.6.8 对 docs/api.md 的变更点
在"插件接口"章节后添加"会话接口"章节，包含上述9个接口的完整文档。

---

### 任务7：SSE实时推送（Server-Sent Events）

#### 3.7.1 范围说明
实现SSE服务，支持会话步骤实时推送、进度更新、错误通知。

#### 3.7.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| GET | /api/v1/sessions/:id/stream | - | SSE流 |

**SSE事件格式**：
```json
{
  "event": "step_created|step_updated|session_completed|error",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 3.7.3 数据表/模型
无需额外数据表，使用内存管理SSE连接。

#### 3.7.4 关键实现点
- **SSE管理器**（pkg/sse）：
  - SSEManager结构体，管理所有SSE连接
  - Client结构体，表示单个SSE客户端
  - NewSSEManager() *SSEManager
  - AddClient(sessionID uint, client *Client)
  - RemoveClient(sessionID uint, clientID string)
  - Broadcast(sessionID uint, event string, data interface{})
  - SendToClient(sessionID uint, clientID string, event string, data interface{})

- **Handler层**：
  - StreamSession(c *gin.Context)

- **中间件**：
  - SSE中间件，设置正确的响应头
  - 心跳机制，定期发送keep-alive事件

#### 3.7.5 幂等/并发控制点
- **客户端连接**：使用 clientID 唯一标识
- **事件广播**：使用 channel 实现并发安全
- **心跳机制**：定期发送keep-alive，检测断开连接

#### 3.7.6 安全控制点
- **权限验证**：用户只能订阅自己项目的会话
- **连接限制**：限制单个用户的并发SSE连接数
- **超时控制**：设置SSE连接超时时间（从配置读取）
- **资源清理**：客户端断开时及时清理资源

#### 3.7.7 完成标准（DoD）
- [ ] pkg/sse包实现完整
- [ ] SSEHandler实现完整
- [ ] SSE中间件实现
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.7.8 对 docs/api.md 的变更点
在"会话接口"章节后添加"SSE接口"章节，包含上述接口的完整文档。

---

### 任务8：结算系统（Settlements）

#### 3.8.1 范围说明
实现结算记录的创建、查询、统计，支持按世界/章节/阶段查询。

#### 3.8.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/settlements | CreateSettlementRequest | SettlementResponse |
| GET | /api/v1/settlements | - | PageResponse(list: SettlementResponse[], page_info: PageInfo) |
| GET | /api/v1/settlements/world/:world_id | - | PageResponse(list: SettlementResponse[], page_info: PageInfo) |
| GET | /api/v1/settlements/chapter/:chapter_id | - | PageResponse(list: SettlementResponse[], page_info: PageInfo) |
| GET | /api/v1/settlements/stats | - | StatsResponse |

**CreateSettlementRequest**:
```json
{
  "world_id": "string (required, max:100)",
  "chapter_id": "string (required, max:100)",
  "loop_stage": "string (required, enum:planning/drafting/validation/refinement)",
  "points_delta": "int",
  "payload": {}
}
```

**SettlementResponse**:
```json
{
  "id": 1,
  "world_id": "string",
  "chapter_id": "string",
  "loop_stage": "string",
  "points_delta": 100,
  "payload": {},
  "user_id": 1,
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

**StatsResponse**:
```json
{
  "total_points": 1000,
  "total_entries": 10,
  "by_stage": {
    "planning": 200,
    "drafting": 300,
    "validation": 250,
    "refinement": 250
  }
}
```

#### 3.8.3 数据表/模型
**表名**：settlement_entries

**字段定义**：
- id (uint, PK, auto_increment)
- world_id (string, 100, indexed, not null)
- chapter_id (string, 100, indexed, not null)
- loop_stage (string, 30, indexed, not null)
- points_delta (int)
- payload (JSON)
- user_id (uint, indexed, not null, FK→users.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (world_id)
- INDEX (chapter_id)
- INDEX (loop_stage)
- INDEX (user_id)
- INDEX (world_id, chapter_id, loop_stage)

#### 3.8.4 关键实现点
- **Repository层**：
  - Create(entry *model.SettlementEntry) error
  - FindByID(id uint) (*model.SettlementEntry, error)
  - FindByUserID(userID uint, page, size int) ([]*model.SettlementEntry, int64, error)
  - FindByWorldID(worldID string, page, size int) ([]*model.SettlementEntry, int64, error)
  - FindByChapterID(chapterID string, page, size int) ([]*model.SettlementEntry, int64, error)
  - FindByLoopStage(loopStage string, page, size int) ([]*model.SettlementEntry, int64, error)
  - GetStats(userID uint) (*SettlementStats, error)

- **Service层**：
  - Create(userID uint, worldID, chapterID, loopStage string, pointsDelta int, payload map[string]interface{}) (*model.SettlementEntry, error)
  - GetByID(id uint) (*model.SettlementEntry, error)
  - ListByUserID(userID uint, page, size int) ([]*model.SettlementEntry, int64, error)
  - ListByWorldID(worldID string, page, size int) ([]*model.SettlementEntry, int64, error)
  - ListByChapterID(chapterID string, page, size int) ([]*model.SettlementEntry, int64, error)
  - ListByLoopStage(loopStage string, page, size int) ([]*model.SettlementEntry, int64, error)
  - GetStats(userID uint) (*SettlementStats, error)

- **Handler层**：
  - Create(c *gin.Context)
  - List(c *gin.Context)
  - ListByWorld(c *gin.Context)
  - ListByChapter(c *gin.Context)
  - GetStats(c *gin.Context)

#### 3.8.5 幂等/并发控制点
- **创建结算**：使用 world_id + chapter_id + loop_stage 唯一约束防止重复
- **更新结算**：默认最后写入覆盖；可选基于 updated_at 做冲突检测

#### 3.8.6 安全控制点
- **权限验证**：用户只能查询自己的结算记录
- **参数验证**：world_id必填，chapter_id必填，loop_stage必填且在枚举范围内
- **SQL注入防护**：使用GORM参数化查询

#### 3.8.7 完成标准（DoD）
- [ ] SettlementRepository实现完整
- [ ] SettlementService实现完整
- [ ] SettlementHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.8.8 对 docs/api.md 的变更点
在"SSE接口"章节后添加"结算接口"章节，包含上述5个接口的完整文档。

---

### 任务9：语料库管理（Corpus）

#### 3.9.1 范围说明
实现语料库故事的导入、查询、统计，支持按题材分类、字数统计。

#### 3.9.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/corpus/upload | UploadCorpusRequest | CorpusResponse |
| GET | /api/v1/corpus | - | PageResponse(list: CorpusResponse[], page_info: PageInfo) |
| GET | /api/v1/corpus/:id | - | CorpusResponse |
| GET | /api/v1/corpus/genre/:genre | - | PageResponse(list: CorpusResponse[], page_info: PageInfo) |
| GET | /api/v1/corpus/stats | - | CorpusStatsResponse |
| DELETE | /api/v1/corpus/:id | - | Success |

**UploadCorpusRequest**:
```json
{
  "title": "string (required, max:200)",
  "genre": "string (max:50)",
  "file_path": "string (required, max:500)",
  "metadata": {}
}
```

**CorpusResponse**:
```json
{
  "id": 1,
  "title": "string",
  "genre": "string",
  "file_path": "string",
  "file_size": 1024,
  "word_count": 10000,
  "metadata": {},
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

**CorpusStatsResponse**:
```json
{
  "total_stories": 100,
  "total_words": 1000000,
  "by_genre": {
    "都市": 50,
    "末世": 30,
    "恐怖": 20
  }
}
```

#### 3.9.3 数据表/模型
**表名**：corpus_stories

**字段定义**：
- id (uint, PK, auto_increment)
- title (string, 200, not null)
- genre (string, 50, indexed)
- file_path (string, 500, not null)
- file_size (int64)
- word_count (int)
- metadata (JSON)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (genre)

#### 3.9.4 关键实现点
- **Repository层**：
  - Create(story *model.CorpusStory) error
  - FindByID(id uint) (*model.CorpusStory, error)
  - FindAll(page, size int) ([]*model.CorpusStory, int64, error)
  - FindByGenre(genre string, page, size int) ([]*model.CorpusStory, int64, error)
  - Update(story *model.CorpusStory) error
  - Delete(id uint) error
  - GetStats() (*CorpusStats, error)

- **Service层**：
  - Upload(title, genre, filePath string, metadata map[string]interface{}) (*model.CorpusStory, error)
  - GetByID(id uint) (*model.CorpusStory, error)
  - List(page, size int) ([]*model.CorpusStory, int64, error)
  - ListByGenre(genre string, page, size int) ([]*model.CorpusStory, int64, error)
  - Delete(id uint) error
  - GetStats() (*CorpusStats, error)

- **Handler层**：
  - Upload(c *gin.Context)
  - List(c *gin.Context)
  - GetByID(c *gin.Context)
  - ListByGenre(c *gin.Context)
  - GetStats(c *gin.Context)
  - Delete(c *gin.Context)

- **文件处理**：
  - 创建 pkg/fileutil 包，封装文件读取、字数统计逻辑
  - 支持多种编码格式（UTF-8、GBK等）

#### 3.9.5 幂等/并发控制点
- **上传语料**：使用 file_path 唯一约束防止重复
- **更新语料**：默认最后写入覆盖；可选基于 updated_at 做冲突检测

#### 3.9.6 安全控制点
- **权限验证**：语料库管理需要管理员权限
- **参数验证**：title必填，file_path必填且为有效路径
- **SQL注入防护**：使用GORM参数化查询
- **文件安全**：验证文件路径在允许的目录内

#### 3.9.7 完成标准（DoD）
- [ ] CorpusRepository实现完整
- [ ] CorpusService实现完整
- [ ] CorpusHandler实现完整
- [ ] pkg/fileutil包实现
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.9.8 对 docs/api.md 的变更点
在"结算接口"章节后添加"语料库接口"章节，包含上述6个接口的完整文档。

---

### 任务10：文件管理（Files）

#### 3.10.1 范围说明
实现文件上传、下载、删除，支持文件元信息管理、SHA256校验。

#### 3.10.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/files/upload | multipart/form-data | FileResponse |
| GET | /api/v1/files/:id | - | FileResponse |
| GET | /api/v1/files/:id/download | - | 文件流 |
| DELETE | /api/v1/files/:id | - | Success |
| GET | /api/v1/files | - | PageResponse(list: FileResponse[], page_info: PageInfo) |

**FileResponse**:
```json
{
  "id": 1,
  "file_name": "string",
  "file_type": "string",
  "content_type": "string",
  "size_bytes": 1024,
  "storage_key": "string",
  "sha256": "string",
  "user_id": 1,
  "project_id": 1,
  "created_at": "2024-01-01 00:00:00",
  "updated_at": "2024-01-01 00:00:00"
}
```

#### 3.10.3 数据表/模型
**表名**：files

**字段定义**：
- id (uint, PK, auto_increment)
- file_name (string, 255, not null)
- file_type (string, 20, indexed, not null)
- content_type (string, 100)
- size_bytes (int64)
- storage_key (string, 500, uniqueIndex, not null)
- sha256 (string, 64, indexed)
- user_id (uint, indexed, not null, FK→users.id)
- project_id (uint, indexed, FK→projects.id)
- created_at (datetime)
- updated_at (datetime)
- deleted_at (datetime, soft delete)

**索引**：
- PRIMARY KEY (id)
- INDEX (file_type)
- INDEX (sha256)
- INDEX (user_id)
- INDEX (project_id)
- UNIQUE INDEX (storage_key)

#### 3.10.4 关键实现点
- **Repository层**：
  - Create(file *model.File) error
  - FindByID(id uint) (*model.File, error)
  - FindByStorageKey(storageKey string) (*model.File, error)
  - FindBySHA256(sha256 string) (*model.File, error)
  - FindByUserID(userID uint, page, size int) ([]*model.File, int64, error)
  - FindByProjectID(projectID uint, page, size int) ([]*model.File, int64, error)
  - Delete(id uint) error

- **Service层**：
  - Upload(userID uint, projectID *uint, fileType string, fileHeader *multipart.FileHeader) (*model.File, error)
  - GetByID(id uint) (*model.File, error)
  - Download(id uint) (string, io.ReadCloser, error)
  - Delete(id uint) error
  - ListByUserID(userID uint, page, size int) ([]*model.File, int64, error)
  - ListByProjectID(projectID uint, page, size int) ([]*model.File, int64, error)

- **Handler层**：
  - Upload(c *gin.Context)
  - GetByID(c *gin.Context)
  - Download(c *gin.Context)
  - Delete(c *gin.Context)
  - List(c *gin.Context)

- **存储服务**（pkg/storage）：
  - Storage接口，支持本地存储和云存储
  - LocalStorage实现
  - UploadFile(key string, reader io.Reader) error
  - DownloadFile(key string) (io.ReadCloser, error)
  - DeleteFile(key string) error
  - GetFileURL(key string) string

#### 3.10.5 幂等/并发控制点
- **上传文件**：使用 storage_key 唯一约束防止重复
- **SHA256去重**：相同SHA256的文件只存储一份

#### 3.10.6 安全控制点
- **权限验证**：用户只能操作自己的文件
- **参数验证**：file_type必填且在枚举范围内（upload/export）
- **文件类型验证**：验证content_type与文件扩展名匹配
- **文件大小限制**：从配置读取最大文件大小
- **路径遍历防护**：验证storage_key不包含路径遍历字符
- **SQL注入防护**：使用GORM参数化查询

#### 3.10.7 完成标准（DoD）
- [ ] FileRepository实现完整
- [ ] FileService实现完整
- [ ] FileHandler实现完整
- [ ] pkg/storage包实现
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.10.8 对 docs/api.md 的变更点
在"语料库接口"章节后添加"文件接口"章节，包含上述5个接口的完整文档。

---

### 任务11：排版引擎（Formatting）

#### 3.11.1 范围说明
实现排版引擎，支持缩进、段落间距、标题格式、标点全角化、字数检查。

#### 3.11.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/formatting/apply | ApplyFormattingRequest | FormattingResponse |
| POST | /api/v1/formatting/validate | ValidateFormattingRequest | ValidationResponse |

**ApplyFormattingRequest**:
```json
{
  "content": "string (required)",
  "rules": {
    "indent": "string (default:全角空格)",
    "paragraph_spacing": "int (default:2)",
    "title_format": "string (default:第X章)",
    "punctuation_fullwidth": "bool (default:true)",
    "target_word_count": "int (default:2250)",
    "tolerance": "int (default:250)"
  }
}
```

**FormattingResponse**:
```json
{
  "formatted_content": "string",
  "word_count": 2250,
  "changes": [
    {
      "type": "string",
      "position": 0,
      "original": "string",
      "formatted": "string"
    }
  ]
}
```

**ValidateFormattingRequest**:
```json
{
  "content": "string (required)",
  "rules": {}
}
```

**ValidationResponse**:
```json
{
  "is_valid": true,
  "word_count": 2250,
  "errors": [],
  "warnings": []
}
```

#### 3.11.3 数据表/模型
无需额外数据表，使用配置文件存储排版规则。

#### 3.11.4 关键实现点
- **排版服务**（pkg/formatting）：
  - FormattingEngine结构体
  - ApplyFormatting(content string, rules FormattingRules) (string, []FormattingChange, error)
  - ValidateFormatting(content string, rules FormattingRules) (*ValidationResult, error)
  - CountWords(content string) int
  - ConvertToFullwidthPunctuation(content string) string
  - ApplyIndent(content string, indent string) string
  - ApplyParagraphSpacing(content string, spacing int) string
  - FormatTitle(content string, format string) string

- **排版规则**（configs/formatting.yaml）：
  - indent: 全角空格
  - paragraph_spacing: 2
  - title_format: 第X章
  - punctuation_fullwidth: true
  - target_word_count: 2250
  - tolerance: 250

- **Handler层**：
  - ApplyFormatting(c *gin.Context)
  - ValidateFormatting(c *gin.Context)

#### 3.11.5 幂等/并发控制点
- **排版应用**：无状态操作，天然幂等
- **排版验证**：无状态操作，天然幂等

#### 3.11.6 安全控制点
- **参数验证**：content必填
- **SQL注入防护**：不涉及数据库操作
- **资源限制**：限制content最大长度（从配置读取）

#### 3.11.7 完成标准（DoD）
- [ ] pkg/formatting包实现完整
- [ ] configs/formatting.yaml配置文件
- [ ] FormattingHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.11.8 对 docs/api.md 的变更点
在"文件接口"章节后添加"排版接口"章节，包含上述2个接口的完整文档。

---

### 任务12：质量门禁（Quality Gate）

#### 3.12.1 范围说明
实现质量门禁，支持逻辑一致性、人物一致性、伏笔检查、风格一致性检查。

#### 3.12.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| POST | /api/v1/quality/check | CheckQualityRequest | QualityCheckResponse |
| POST | /api/v1/quality/alignment | CheckAlignmentRequest | AlignmentCheckResponse |

**CheckQualityRequest**:
```json
{
  "content": "string (required)",
  "project_id": "int (required)",
  "checks": {
    "logic_consistency": "bool (default:true)",
    "character_consistency": "bool (default:true)",
    "foreshadowing": "bool (default:true)",
    "style_consistency": "bool (default:true)"
  }
}
```

**QualityCheckResponse**:
```json
{
  "overall_score": 85,
  "passed": true,
  "results": {
    "logic_consistency": {
      "score": 90,
      "passed": true,
      "issues": []
    },
    "character_consistency": {
      "score": 80,
      "passed": true,
      "issues": []
    },
    "foreshadowing": {
      "score": 85,
      "passed": true,
      "issues": []
    },
    "style_consistency": {
      "score": 85,
      "passed": true,
      "issues": []
    }
  }
}
```

**CheckAlignmentRequest**:
```json
{
  "content": "string (required)",
  "project_id": "int (required)",
  "reference_content": "string (required)"
}
```

**AlignmentCheckResponse**:
```json
{
  "alignment_score": 85,
  "passed": true,
  "issues": []
}
```

#### 3.12.3 数据表/模型
无需额外数据表，使用配置文件存储质量规则。

#### 3.12.4 关键实现点
- **质量门禁服务**（pkg/quality）：
  - QualityGate结构体
  - CheckQuality(content string, projectID uint, checks QualityChecks) (*QualityResult, error)
  - CheckAlignment(content, referenceContent string, projectID uint) (*AlignmentResult, error)
  - CheckLogicConsistency(content string, projectID uint) (*CheckResult, error)
  - CheckCharacterConsistency(content string, projectID uint) (*CheckResult, error)
  - CheckForeshadowing(content string, projectID uint) (*CheckResult, error)
  - CheckStyleConsistency(content string, projectID uint) (*CheckResult, error)

- **质量规则**（configs/quality.yaml）：
  - logic_consistency:
    - enabled: true
    - threshold: 70
  - character_consistency:
    - enabled: true
    - threshold: 70
  - foreshadowing:
    - enabled: true
    - threshold: 70
  - style_consistency:
    - enabled: true
    - threshold: 70

- **Handler层**：
  - CheckQuality(c *gin.Context)
  - CheckAlignment(c *gin.Context)

- **插件集成**：
  - 通过插件系统调用AI能力进行质量检查
  - 支持自定义检查规则

#### 3.12.5 幂等/并发控制点
- **质量检查**：无状态操作，天然幂等
- **对齐检查**：无状态操作，天然幂等

#### 3.12.6 安全控制点
- **权限验证**：用户只能检查自己项目的质量
- **参数验证**：content必填，project_id必填
- **SQL注入防护**：使用GORM参数化查询
- **资源限制**：限制content最大长度（从配置读取）

#### 3.12.7 完成标准（DoD）
- [ ] pkg/quality包实现完整
- [ ] configs/quality.yaml配置文件
- [ ] QualityHandler实现完整
- [ ] 路由注册完成
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.12.8 对 docs/api.md 的变更点
在"排版接口"章节后添加"质量门禁接口"章节，包含上述2个接口的完整文档。

---

### 任务13：项目导出（Project Export）

#### 3.13.1 范围说明
实现项目导出功能，支持导出为JSON格式，包含项目、卷、文档、实体、模板等完整数据。

#### 3.13.2 接口清单
| HTTP方法 | 路径 | 请求体 | 响应（data） |
|---------|------|--------|------|
| GET | /api/v1/projects/:id/export | - | ProjectExportResponse |

**ProjectExportResponse**:
```json
{
  "project": {
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
  },
  "volumes": [],
  "documents": [],
  "entities": [],
  "templates": []
}
```

#### 3.13.3 数据表/模型
无需额外数据表，使用现有数据表。

#### 3.13.4 关键实现点
- **Service层**：
  - GetByIDWithDetails(id uint) (*model.Project, error)
  - 使用GORM的Preload加载关联数据：
    - Preload("Volumes")
    - Preload("Volumes.Documents")
    - Preload("Entities")
    - Preload("Entities.Tags")
    - Preload("Entities.Links")
    - Preload("Templates")

- **Handler层**：
  - Export(c *gin.Context)

- **导出格式**：
  - JSON格式
  - 包含所有关联数据
  - 保持数据结构完整性

#### 3.13.5 幂等/并发控制点
- **项目导出**：无状态操作，天然幂等

#### 3.13.6 安全控制点
- **权限验证**：用户只能导出自己的项目
- **SQL注入防护**：使用GORM参数化查询
- **数据脱敏**：导出时移除敏感信息（如用户密码）

#### 3.13.7 完成标准（DoD）
- [ ] ProjectService.GetByIDWithDetails实现完整
- [ ] ProjectHandler.Export实现完整
- [ ] docs/api.md更新
- [ ] 单元测试通过（可选）

#### 3.13.8 对 docs/api.md 的变更点
更新"项目接口"章节中的"导出项目"接口文档。

---

## 4. 里程碑与交付物清单

### 第1周：基础模块（Volumes、Documents）
- **交付物**：
  - VolumeRepository、VolumeService、VolumeHandler
  - DocumentRepository、DocumentService、DocumentHandler
  - 路由注册完成
  - docs/api.md更新

### 第2周：实体与模板（Entities、Templates）
- **交付物**：
  - EntityRepository、EntityService、EntityHandler
  - TemplateRepository、TemplateService、TemplateHandler
  - 路由注册完成
  - docs/api.md更新

### 第3周：插件与会话（Plugins、Sessions）
- **交付物**：
  - PluginRepository、PluginService、PluginHandler
  - pkg/httpclient包
  - SessionRepository、SessionService、SessionHandler
  - 路由注册完成
  - docs/api.md更新

### 第4周：SSE与结算（SSE、Settlements）
- **交付物**：
  - pkg/sse包
  - SSEHandler、SSE中间件
  - SettlementRepository、SettlementService、SettlementHandler
  - 路由注册完成
  - docs/api.md更新

### 第5周：语料与文件（Corpus、Files）
- **交付物**：
  - CorpusRepository、CorpusService、CorpusHandler
  - pkg/fileutil包
  - FileRepository、FileService、FileHandler
  - pkg/storage包
  - 路由注册完成
  - docs/api.md更新

### 第6周：排版与质量（Formatting、Quality Gate）
- **交付物**：
  - pkg/formatting包
  - configs/formatting.yaml
  - pkg/quality包
  - configs/quality.yaml
  - FormattingHandler、QualityHandler
  - 路由注册完成
  - docs/api.md更新

### 第7周：项目导出与集成测试
- **交付物**：
  - ProjectService.GetByIDWithDetails实现
  - ProjectHandler.Export实现
  - 集成测试
  - 性能测试
  - 文档完善

---

## 5. 风险清单与缓解策略

### 5.1 SSE风险
- **风险**：SSE连接数过多导致服务器资源耗尽
- **缓解策略**：
  - 限制单个用户的并发SSE连接数（从配置读取）
  - 实现心跳机制，及时清理断开连接
  - 使用连接池管理SSE连接
  - 监控SSE连接数，超过阈值时拒绝新连接

### 5.2 文件上传风险
- **风险**：大文件上传导致服务器资源耗尽
- **缓解策略**：
  - 限制单个文件最大大小（从配置读取，默认100MB）
  - 使用流式上传，避免内存占用过高
  - 实现上传进度监控
  - 支持断点续传（可选）

### 5.3 项目导出风险
- **风险**：大数据量导出导致超时或内存溢出
- **缓解策略**：
  - 实现分页导出
  - 使用流式响应
  - 设置导出超时时间（从配置读取）
  - 提供异步导出功能（可选）

### 5.4 插件调用风险
- **风险**：插件调用超时或失败影响系统稳定性
- **缓解策略**：
  - 设置插件调用超时时间（从配置读取，默认30秒）
  - 实现重试机制（最多3次）
  - 记录插件调用日志
  - 实现插件健康检查，自动禁用异常插件

### 5.5 并发幂等风险
- **风险**：并发请求导致数据重复或冲突
- **缓解策略**：
  - 使用数据库唯一约束
  - 实现更新冲突检测（基于 updated_at，可选）
  - 使用分布式锁（可选）
  - 记录操作日志，便于排查问题

### 5.6 数据库性能风险
- **风险**：大数据量查询导致性能下降
- **缓解策略**：
  - 合理设计索引
  - 使用分页查询
  - 实现查询缓存（可选）
  - 监控慢查询，优化SQL

---

## 6. 文档与契约更新策略

### 6.1 接口变更流程
1. **新增接口**：
   - 在 docs/api.md 中添加接口文档
   - 包含HTTP方法、路径、请求体、响应、错误码
   - 更新"数据模型"章节（如有新增模型）

2. **修改接口**：
   - 在 docs/api.md 中更新接口文档
   - 标注变更内容（使用"新增"、"修改"、"删除"标记）
   - 更新"数据模型"章节（如有模型变更）

3. **删除接口**：
   - 在 docs/api.md 中标记接口为"已废弃"
   - 保留文档至少3个月
   - 更新"数据模型"章节（如有模型删除）

### 6.2 文档更新时机
- **代码提交前**：必须更新 docs/api.md
- **代码审查时**：检查文档是否完整
- **发布前**：验证文档与代码一致性

### 6.3 文档格式规范
- 使用Markdown格式
- 接口文档包含：HTTP方法、路径、描述、认证、请求体、响应
- 数据模型包含：字段、类型、说明
- 错误码包含：错误码、说明

---

## 7. 最终 DoD

### 7.1 功能验收
- [ ] 13个核心模块全部实现
- [ ] 所有接口功能正常
- [ ] docs/api.md更新完整
- [ ] 配置文件完整（configs/formatting.yaml、configs/quality.yaml）

### 7.2 安全验收
- [ ] 所有接口权限验证正常
- [ ] SQL注入防护正常
- [ ] XSS防护正常
- [ ] 文件上传安全验证正常
- [ ] 插件调用安全验证正常

### 7.3 性能验收
- [ ] 接口响应时间<1秒（P95）
- [ ] 并发100请求时系统稳定
- [ ] 数据库查询优化完成
- [ ] SSE连接管理正常

### 7.4 可运维验收
- [ ] 日志记录完整
- [ ] 错误处理完整
- [ ] 监控指标完整
- [ ] 配置管理完整
- [ ] 部署文档完整

### 7.5 代码质量
- [ ] 代码符合Go规范
- [ ] 代码注释完整
- [ ] 代码复用性良好
- [ ] 代码可维护性良好

---

## 附录A：错误码扩展

### 会话模块（40xxx）
- 40001: 会话不存在
- 40002: 会话步骤不存在
- 40003: 会话模式无效
- 40004: 会话步骤顺序冲突

### 插件模块（50xxx）
- 50001: 插件不存在
- 50002: 插件已禁用
- 50003: 插件离线
- 50004: 插件能力不存在
- 50005: 插件调用超时
- 50006: 插件调用失败

### 文件模块（60xxx）
- 60001: 文件不存在
- 60002: 文件类型不支持
- 60003: 文件大小超限
- 60004: 文件上传失败
- 60005: 文件下载失败

### 排版模块（70xxx）
- 70001: 排版规则无效
- 70002: 排版失败

### 质量门禁模块（80xxx）
- 80001: 质量检查失败
- 80002: 对齐检查失败

---

## 附录B：配置文件示例

### configs/formatting.yaml
```yaml
formatting:
  indent: "　"  # 全角空格
  paragraph_spacing: 2
  title_format: "第{number}章"
  punctuation_fullwidth: true
  target_word_count: 2250
  tolerance: 250
```

### configs/quality.yaml
```yaml
quality:
  logic_consistency:
    enabled: true
    threshold: 70
  character_consistency:
    enabled: true
    threshold: 70
  foreshadowing:
    enabled: true
    threshold: 70
  style_consistency:
    enabled: true
    threshold: 70
```

### configs/config.yaml（新增配置）
```yaml
# ... existing config ...

file:
  max_size: 104857600  # 100MB
  allowed_types:
    - txt
    - json
    - md

sse:
  max_connections_per_user: 5
  heartbeat_interval: 30  # seconds
  connection_timeout: 300  # seconds

plugin:
  call_timeout: 30  # seconds
  max_retries: 3
  health_check_interval: 60  # seconds
```

---

## 附录C：目录结构

```
internal/
├── handler/
│   ├── volume_handler.go
│   ├── document_handler.go
│   ├── entity_handler.go
│   ├── template_handler.go
│   ├── plugin_handler.go
│   ├── session_handler.go
│   ├── settlement_handler.go
│   ├── corpus_handler.go
│   ├── file_handler.go
│   ├── formatting_handler.go
│   └── quality_handler.go
├── service/
│   ├── volume_service.go
│   ├── document_service.go
│   ├── entity_service.go
│   ├── template_service.go
│   ├── plugin_service.go
│   ├── session_service.go
│   ├── settlement_service.go
│   ├── corpus_service.go
│   └── file_service.go
├── repository/
│   ├── volume_repository.go
│   ├── document_repository.go
│   ├── entity_repository.go
│   ├── template_repository.go
│   ├── plugin_repository.go
│   ├── session_repository.go
│   ├── settlement_repository.go
│   ├── corpus_repository.go
│   └── file_repository.go
└── model/
    ├── volume.go (已存在)
    ├── document.go (已存在)
    ├── entity.go (已存在)
    ├── template.go (已存在)
    ├── plugin.go (已存在)
    ├── session.go (已存在)
    ├── settlement.go (已存在)
    ├── corpus.go (已存在)
    └── file.go (已存在)

pkg/
├── sse/
│   ├── manager.go
│   └── client.go
├── httpclient/
│   └── client.go
├── fileutil/
│   └── util.go
├── storage/
│   ├── storage.go
│   └── local.go
├── formatting/
│   ├── engine.go
│   └── rules.go
└── quality/
    ├── gate.go
    └── checker.go

configs/
├── config.yaml (更新)
├── formatting.yaml (新增)
└── quality.yaml (新增)

docs/
└── api.md (更新)
```

---

**文档版本**：v1.0
**创建日期**：2026-02-02
**最后更新**：2026-02-02
