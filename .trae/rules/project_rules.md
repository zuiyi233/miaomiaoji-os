---
alwaysApply: false
---
**通用协作约束（3 个智能体都要遵守）**
- 架构分层：Router → Handler → Service → Repository → Model
- 技术栈与规范：Gin / GORM v2 / Viper / Wire / Zap；统一使用 `pkg/response` 返回；错误码用 `pkg/errors`；日志用 `pkg/logger`
- 不硬编码配置（JWT secret、路径、大小限制等都从 config 读）
- 不自动生成测试代码（除非我明确要求）
- 修改/新增接口后：更新已有 `docs/api.md`（不新建 md）
- 避免冲突：每个智能体只改自己负责的目录/文件清单；公共改动尽量集中到智能体 A

---

## 智能体 A：工程骨架 + 基础设施（Owner of “可运行服务”）
**任务方案**
- 从 `rulebacktest-main` 模板初始化新后端工程（目录、go.mod、配置加载、日志、DB、迁移、Wire 注入）
- 中间件与基础能力：RequestID、Recovery、Logger、CORS、RateLimit、JWT Auth/Role
- 建立基础路由分组 `/api/v1`，并把后续模块的路由注册入口预留好
- 输出：项目能启动、能连库、能 AutoMigrate、基础健康检查接口可用

**建议负责文件范围**
- `cmd/server/*`
- `internal/config/*`
- `pkg/*`（如果需要调整）
- `internal/middleware/*`
- `internal/router/*`
- `internal/wire/*`
- `pkg/database/*`（或现有位置）

**提示词（复制给智能体 A）**
> 你是 Go 后端工程搭建负责人。请在仓库中基于 rulebacktest-main 初始化 miaomiaoji-os 的可运行骨架：Gin+GORM+Viper+Wire+Zap。  
> 目标：服务可启动、可加载双环境配置、可连接 SQLite/Postgres、可 AutoMigrate、具备统一响应与错误码、具备 JWT 认证与角色校验中间件、具备 RequestID/Recovery/Logger/CORS/RateLimit。  
> 约束：禁止硬编码关键配置；响应必须走 pkg/response；错误码走 pkg/errors；日志走 pkg/logger；不写业务逻辑到 handler 之外；不生成测试代码；不新建文档文件；如新增/调整接口，更新已有 docs/api.md。  
> 输出要求：  
> 1) 给出你修改的文件清单；2) 给出可运行的启动命令；3) 启动后至少提供 /healthz 或等价健康检查接口；4) Wire 初始化可生成并通过编译。

---

## 智能体 B：核心资源 CRUD（Project/Volume/Document/Entity）
**任务方案**
- 按文档方案实现核心数据模型与 CRUD：Project、Volume、Document、Entity
- 按最新方案实现关系去 JSONB 化：`DocumentEntityRef`、`EntityTag`（包含必要索引）
- 完整打通 Router/Handler/Service/Repository（不把业务逻辑塞 Handler）
- 输出：前端“项目-卷-文档-实体”的主链路接口可用（分页/排序/reorder、过滤 type/tag、文档关联实体）

**建议负责文件范围**
- `internal/model/{project,volume,document,entity}*`
- `internal/repository/{project,volume,document,entity}*`
- `internal/service/{project,volume,document,entity}*`
- `internal/handler/{project,volume,document,entity}*`
- 仅在需要注册路由时改 `internal/router/*` 的对应子路由文件（尽量别动全局）

**提示词（复制给智能体 B）**
> 你是后端核心业务 CRUD 负责人。请基于现有工程骨架，实现项目/卷/文档/实体模块的完整 CRUD，并严格遵循 Router→Handler→Service→Repository→Model 分层。  
> 必须按方案实现：  
> - 移除 Document.linked_ids JSONB；新增 DocumentEntityRef（document_id/entity_id/ref_type/metadata）并提供文档关联实体的增删查接口  
> - 移除 Entity.tags JSONB；新增 EntityTag（entity_id/tag）并支持实体列表按 type/tag 过滤  
> - Volume reorder 批量调整顺序接口  
> 约束：统一响应用 pkg/response；错误码用 pkg/errors；日志用 pkg/logger；配置从 config 读取；不生成测试代码；不新建文档文件；如新增/调整接口，更新已有 docs/api.md。  
> 输出要求：  
> 1) 列出实现的接口与路由；2) 列出数据库表与关键索引；3) 确保编译通过。

---

## 智能体 C：工作流与“重服务”（Session/SSE/Formatting/Quality/Settlement/File）
**任务方案**
- Session + SSE：会话与步骤追加、导出流事件（先把事件通路打通）
- Formatting/Quality：把“番茄排版/质量门禁”作为服务接口落地（先实现最小可用版本，规则从配置读取）
- Settlement：按新建模实现 `SettlementEntry`（逐条记录、world_id/chapter_id 字符串、支持筛选/导出）
- File：按方案实现 `File` 元信息表 + 存储抽象 `Storage interface`，提供文件列表/下载/删除接口（上传/导出都能落库记录）

**建议负责文件范围**
- `internal/model/{session,settlement,file,corpus?}*`
- `internal/repository/{session,settlement,file}*`
- `internal/service/{session,formatting,quality,settlement,file}*`
- `internal/handler/{session,formatting,quality,settlement,file}*`
- `storage/*`（仅代码需要，不要做“说明文档”）

**提示词（复制给智能体 C）**
> 你是后端工作流与文件/结算/实时能力负责人。请实现：Session+SessionStep 的 CRUD 与 SSE stream（至少支持 step.appended、quality.checked、export.ready、error 事件）；实现 formatting 与 quality 两个服务接口（规则/阈值从配置读取）；实现 SettlementEntry（逐条落库，world_id/chapter_id 为 string，支持按 world_id/chapter_id/loop_stage 过滤与导出）；实现 File 元信息表与 Storage 接口抽象，并提供 /api/v1/files 列表/详情/下载/删除。  
> 约束：统一响应用 pkg/response；错误码用 pkg/errors；日志用 pkg/logger；关键配置从 config 读取（上传大小、路径、清理间隔等）；不生成测试代码；不新建文档文件；如新增/调整接口，更新已有 docs/api.md。  
> 输出要求：  
> 1) SSE 的事件协议与示例响应；2) File/Settlement 的表字段与索引；3) 给出最小可用的 Storage 本地实现（Put/Open/Delete）并与 File 表打通；4) 确保编译通过。

---

**并行落地的合并顺序（建议）**
- 先合并 A（骨架可跑）→ 再合并 B/C（业务模块并行）→ 最后统一把 `docs/api.md` 对齐一次（谁改接口谁更新，但最后做一次冲突清理）。