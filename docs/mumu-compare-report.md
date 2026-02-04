# MuMuAINovel 对标分析与 Go 实现路线图（我方项目）

> 目的：对照 MuMuAINovel 的模块与流程，形成我方功能全景图、缺口清单与落地优先级路线图。
> 语言口径：对标项目为 Python 实现；我方目标是 **用 Go 复现功能**，以现有 Go 分层（Router → Handler → Service → Repository → Model）落地，不复用 Python 细节实现。

## 0. 使用方式与边界

本文是对标驱动的“功能梳理 + 缺口 + 路线图”，用于指导迭代拆分与优先级，不是最终的详细技术设计。

### 0.1 以谁为准（证据优先级）

1. 我方代码（`internal/router/router.go` 与各 Handler/Service/Repo/Model）为最终事实来源。
2. `docs/api.md` 为契约参考，但以代码实现为准。
3. MuMuAINovel 文档仅作为对标基准，不直接约束我方实现细节。

### 0.2 术语映射（对标到我方）

| MuMu 术语 | MuMu 含义 | 我方当前承载 | 我方建议统一口径 |
|---|---|---|---|
| Project | 项目容器（设定+章节） | Project | Project（保持一致） |
| Wizard Step | 立项向导步骤产物 | SessionStep（工作流步骤） | 统一用 Session/SessionStep 承载，`format_type` 区分 `wizard_*` |
| Chapter | 章节实体（生成/分析/重写） | Document（文档）+ Volume（卷） | Document 增加类型：`chapter`，并与 Volume 绑定 |
| Generation History / Task | 生成记录与任务状态 | Job + SessionStep | Job 负责异步与状态；SessionStep 负责可展示的过程与产物 |
| MCP Plugin / Tool | 可注入 AI 的工具 | Plugin + PluginCapability | Capability 定义为“工具”，Job 执行回填 SessionStep |
| Memories / Foreshadows | 记忆与伏笔 | Corpus（语料）+ Entity | P0 先规则检索/标签，P2 再向量检索 |

## 1. 对标基准表（MuMuAINovel）

> 证据来源：`MuMuAINovel 项目流程总结.md` 对应章节

| 模块名 | 用户价值 | 关键数据模型 | 关键接口/交互 | 关键技术点 | 证据引用（章节） |
|---|---|---|---|---|---|
| 立项与设定层 | 从立项到可写设定闭环 | Project、Worldview、Career、Character、Relationship、Organization、WritingStyle、PromptTemplate | 向导分步生成（SSE） | 流式 SSE、结构化 Prompt、上下文注入 | §1 总览；§3 工作流；§6.1 向导 |
| 章节生产层 | 章节生成/分析/重写/批量生产 | Chapter、生成历史、任务状态 | 章节生成/分析/重写/批量生成接口 | SSE 长链路与任务状态 | §1 总览；§6.2 章节生产线 |
| 工具增强层（MCP） | 外部工具增强 AI | MCP 插件模型 | 插件配置/管理、工具注入、工具调用 | MCP 门面、工具缓存、自动重连 | §4 插件功能 |
| 关系图谱（CP） | 角色关系网络 | RelationshipType、CharacterRelationship、Organization | 关系图谱 API | 图谱结构化数据 | §5 CP 功能 |
| 一致性约束模块 | 设定一致性 | PromptTemplate、WritingStyle、Memories、Foreshadows | 多模块 API | Prompt JSON 约束 | §3.2；§6.3 |
| 智能上下文构建 | 长篇一致性 | ChapterContextBuilder | 内部服务调用 | RTCO 分层上下文+向量检索 | §6.4 |
| SSE 机制 | 贯穿向导/章节生成 | SSE Response | EventSource | 心跳/进度/结果 | §3.1 |

---

## 2. 我方现状表（按相同结构）

> 证据必须来自我方代码；若不确定标“待确认”并给出查证路径。

| 模块名（对标） | 我方现状 | 关键数据模型 | 关键接口/交互 | 关键技术点 | 证据引用（文件/行号） |
|---|---|---|---|---|---|
| 立项与设定层 | **部分实现**：项目 CRUD + 统一实体模型（角色/组织/设定等） | Project、Entity、EntityLink | `/api/v1/projects`、`/api/v1/projects/:project_id/entities` | REST CRUD | `internal/router/router.go:L169-L234`；`internal/model/entity.go:L15-L69`；`internal/handler/entity_handler.go:L61-L218` |
| 向导（SSE） | **部分实现**：前端有向导，但非 SSE；后端有工作流 SSE | Session/SessionStep | `NovelWizard` 调 `generateNovelBlueprint`；SSE 订阅会话流 | 非向导链路；SSE 用于工作流输出 | `web/components/NovelWizard.tsx:L12-L101`；`internal/router/router.go:L297-L302`；`docs/api.md:L930-L1077` |
| 章节生产线 | **部分实现**：文档 CRUD；工作流仅“世界观/润色” | Document | `/api/v1/projects/:project_id/documents` | 无章节生成/分析/重写/批量接口 | `internal/handler/document_handler.go:L25-L257`；`internal/router/router.go:L188-L219` |
| 工具增强（插件/MCP） | **部分实现**：插件管理/调用存在；缺少工具注入到 AI | Plugin、PluginCapability | `/api/v1/plugins/*` + invoke/invoke-async | 插件调用 API | `internal/model/plugin.go:L9-L55`；`internal/handler/plugin_handler.go:L52-L337`；`internal/router/router.go:L246-L263` |
| 关系图谱（CP） | **部分实现**：实体关联 + 前端图谱 + 关系边 API | EntityLink | 图谱组件渲染实体与文档关联 | 前端图谱（本地数据） | `internal/model/entity.go:L54-L69`；`internal/handler/entity_handler.go:L266-L309`；`web/components/GraphVisualizer.tsx:L35-L149` |
| 一致性约束（模板/风格） | **部分实现**：模板 CRUD；前端 AI 提示词模板工具 | Template | `/api/v1/projects/:project_id/templates` | 模板驱动提示词 | `internal/handler/template_handler.go:L26-L188`；`web/components/AIAssistant.tsx:L75-L260` |
| 智能上下文构建 | **未实现**：未发现上下文构建/向量检索服务，仅有语料库模块 | CorpusStory | `/api/v1/corpus/*` | 语料库 CRUD | `internal/handler/corpus_handler.go:L23-L211` |
| SSE 机制 | **已实现**：SSE Hub + Stream | SSEEvent | `/api/v1/sse/stream` | step/quality/export 事件 | `pkg/sse/sse.go:L12-L175`；`internal/handler/sse_handler.go:L24-L109`；`internal/router/router.go:L290-L295` |
| 工作流（世界观/润色） | **已实现**：RunWorld/RunPolish | Session/SessionStep | `/api/v1/workflows/world` `/polish` | AI 代理 + SSE 广播 | `internal/handler/workflow_handler.go:L27-L105`；`internal/service/workflow_service.go:L55-L191` |
| AI 代理 | **已实现**：同步/流式代理 | AI Provider Config | `/api/v1/ai/proxy` `/proxy/stream` | 透传模型 API | `internal/handler/ai_proxy_handler.go:L27-L91`；`internal/handler/ai_proxy_stream_handler.go:L27-L105` |
| 会话/步骤 | **已实现**：会话与步骤 CRUD | Session/SessionStep | `/api/v1/sessions/*` | 工作流落地载体 | `internal/handler/session_handler.go:L23-L259`；`internal/router/router.go:L272-L287` |
| 语料库（类“记忆”） | **已实现**：Corpus CRUD | CorpusStory | `/api/v1/corpus/*` | 语料库管理 | `internal/handler/corpus_handler.go:L23-L211`；`internal/router/router.go:L316-L326` |
| 质量/排版 | **已实现**：质量门禁+排版接口 | 质量结果 | `/api/v1/quality/*` `/api/v1/formatting/*` | 质量门禁与排版能力 | `internal/router/router.go:L340-L351`；`internal/handler/formatting_handler.go:L21-L88` |
| 质量门禁 | **已实现**：质量检查接口 | 质量评估结果 | `/api/v1/quality/*` | 质量检查 | `internal/router/router.go:L347-L352`；`internal/service/quality_service.go:L29-L137` |
| 排版/格式化 | **已实现**：排版接口 | 文本样式 | `/api/v1/formatting/*` | 格式化 | `internal/router/router.go:L341-L345`；`internal/service/formatting_service.go:L25-L159` |
| 文件与导出 | **已实现**：文件 CRUD/下载 | File | `/api/v1/files/*` | 文件管理 | `internal/router/router.go:L329-L337`；`internal/handler/file_handler.go:L23-L260` |
| 任务系统 | **已实现**：任务查询/取消 | Job | `/api/v1/jobs/*` | 异步任务 | `internal/router/router.go:L265-L270`；`internal/handler/job_handler.go:L19-L65` |
| 写作体验（前端） | **部分实现**：编辑器、AI 助手、AgentWriter | Document | AI 续写/润色、插件动作 | 前端驱动 AI | `web/components/Editor.tsx:L18-L132`；`web/components/AgentWriter.tsx:L18-L223` |

---

## 3. 一对一映射与缺口清单

| 模块 | 状态 | 缺口描述 | 依赖项 | 复杂度 | 风险点 | 证据引用 |
|---|---|---|---|---|---|---|
| 立项向导（SSE分步） | 部分实现 | 我方无“分步SSE向导链路”，现为前端本地向导 | Workflow/SSE/SessionStep | 中 | 需定义步骤协议与数据结构 | MuMu §6.1；`NovelWizard.tsx:L12-L101` |
| 章节生产线 | 部分实现 | 缺章节生成/分析/重写/批量接口 | Workflow/AI 代理/SSE | 高 | 长链路稳定性 | MuMu §6.2；`document_handler.go:L25-L257` |
| MCP 工具注入 | 未实现 | 插件存在但未注入 AI 调用 | AIService/插件工具协议 | 高 | 跨模型工具协议适配 | MuMu §4.3；`plugin_handler.go:L52-L337` |
| 关系图谱（CP） | 部分实现 | 关系类型库缺失；关系边 API 已有（实体链接） | Entity 扩展 | 中 | 关系一致性 | MuMu §5；`internal/model/entity.go:L54-L69`；`internal/handler/entity_handler.go:L266-L309` |
| 一致性约束（记忆/伏笔） | 部分实现 | 模板有，记忆/伏笔未见 | 新增模块 | 中 | 数据结构变更 | MuMu §6.3；`template_handler.go:L26-L188` |
| 智能上下文构建 | 未实现 | 未发现上下文构建/向量检索服务 | 章节/记忆/向量检索 | 高 | 成本与复杂度 | MuMu §6.4；`internal/handler/corpus_handler.go:L23-L211` |
| 质量门禁链路 | 部分实现 | 质量检查未与章节生产线闭环（章节工作流尚未实现） | 章节生成链路 | 中 | 质量标准一致性 | `internal/router/router.go:L297-L302`；`internal/router/router.go:L347-L352` |
| 文件导出链路 | 部分实现 | 已有项目导出 JSON，但未见“生成→排版→导出”闭环与 SSE 导出事件 | 导出/排版 | 中 | 导出格式一致性 | `internal/router/router.go:L181-L181`；`internal/handler/project_handler.go:L410-L442`；`internal/router/router.go:L329-L345` |

---

## 4. 优先级路线图（P0/P1/P2）

> 原则：闭环优先（立项 → 写作 → 持续写作）

| 里程碑 | 交付能力 | 改动面 | 验收标准 | 证据引用 |
|---|---|---|---|---|
| P0 | 向导分步SSE（世界观/角色/大纲）+ 章节生成SSE + 会话追踪 | 后端：Workflow + SSE；前端：Wizard/SSE | 新建项目 → 向导完成 → 生成第一章闭环 | MuMu §6.1/§6.2；`workflow_handler.go:L27-L105` |
| P1 | 章节分析/重写、模板驱动一致性 | 后端：模板/分析；前端：分析/重写入口 | 章节可分析并重写；模板可驱动 | MuMu §6.2/§6.3；`template_handler.go:L26-L188` |
| P2 | RTCO上下文构建 + 记忆/伏笔 + 关系图谱增强 | 后端：上下文/向量检索；前端：图谱/伏笔管理 | 长篇一致性提升（50+章） | MuMu §6.4/§5；`GraphVisualizer.tsx:L35-L149` |

### 4.1 P0 前置修复清单（影响闭环的硬阻塞）

| 项 | 问题 | 影响 | 建议修复方向 | 证据引用 |
|---|---|---|---|---|
| 权限校验（Files） | `ListFilesByProject` 仅按 `project_id` 查询，无项目/用户归属校验 | 存在越权读取风险 | 按项目归属校验；或查询时限定 `user_id` | `internal/handler/file_handler.go:L186-L225`；`internal/service/file_service.go:L87-L88` |
| 权限校验（Documents） | `ListByProject`/`GetByID` 未校验项目/文档归属；Service 层仅校验“项目是否存在” | 项目文档可能被越权读取 | List/Create/Detail 统一校验 project.UserID==当前用户 | `internal/handler/document_handler.go:L120-L189`；`internal/service/document_service.go:L83-L100` |
| AI 代理 allowlist | 仅拦截 `..`，缺少路径白名单 | 可被滥用调用非预期上游路径 | 以 provider+path allowlist 约束（或映射成固定枚举） | `internal/handler/ai_proxy_handler.go:L34-L58`；`internal/handler/ai_proxy_stream_handler.go:L34-L58`；`internal/service/workflow_service.go:L98-L101` |
| 配置注入 | Quality/Formatting 用 `config.Config{}` 初始化 | 阈值与规则不稳定，环境不一致 | 统一从配置中心注入真实配置 | `internal/router/router.go:L101-L106` |
| 响应契约一致性 | `ProjectHandler.Export` 直接 `c.JSON` | 与统一 response 包不一致，前端易踩坑 | 统一使用 `response.SuccessWithData` | `internal/handler/project_handler.go:L404-L442` |

---

1. 将工作流扩展为“向导分步SSE”（世界观/角色/大纲）。
2. Wizard 改为调用 `/api/v1/workflows/*` + SSE 会话流。
3. 新增“章节生成SSE”工作流，落地 SessionStep。
4. 章节内容写回 Document，形成持续写作数据层。
5. 保留并打通“章节润色”工作流回写文档。
6. 模板管理接入向导/章节生成作为 Prompt 配置入口。
7. 实体（角色/组织）与章节关联基础链路（EntityLink）。
8. 会话详情页 SSE 实时追加 step 内容。
9. AI 代理统一调用与安全校验（path 白名单）。
10. 验收：新建项目 → 向导完成 → 生成第一章 → 润色 → 保存。

---

## 6. 待确认清单（已核对）

| 项 | 说明 | 证据引用 |
|---|---|---|
| 关系类型库/关系边 API | 未发现关系类型库；关系边 API 以实体链接形式已存在（entities/:id/links） | `internal/model/entity.go:L54-L69`；`internal/handler/entity_handler.go:L266-L309`；`internal/router/router.go:L224-L234` |
| 记忆/伏笔/上下文构建 | 未发现上下文构建/向量检索服务，当前仅有语料库模块 | `internal/handler/corpus_handler.go:L23-L211` |
| 前端路由结构 | 入口由 React Router 渲染 Layout，路由映射集中在 Layout 中 | `web/App.tsx:L9-L26`；`web/components/Layout.tsx:L19-L214` |

---

## 7. Go 实现口径说明（对标复现）

1. **只复现功能链路，不复用 Python 实现细节**。
2. **按我方 Go 分层落地**：Router → Handler → Service → Repository → Model。
3. **统一 SSE 事件协议**（当前为 step/quality/export），向导/章节生成复用同一事件规范。
4. **插件/MCP 能力**：以 Go 服务实现“工具注入 + 执行回填”的完整链路。

---

## 8. 代码审查附录（全栈 + 结构级 + 关键代码级）

### 8.1 结构级发现

| 主题 | 发现 | 风险/影响 | 证据引用 |
|---|---|---|---|
| 模块边界 | 工作流仅覆盖“世界观/润色”，未覆盖章节生成/分析/重写 | 闭环不完整 | `internal/router/router.go:L297-L302`；`internal/handler/workflow_handler.go:L37-L105` |
| 插件/MCP | 插件管理完整，但未见 AI 调用时的工具注入链路 | 无法复现 MuMu MCP 核心能力 | `internal/handler/plugin_handler.go:L52-L337`；`web/services/pluginService.ts:L127-L229` |
| 图谱 | 有实体关联与前端图谱，关系边 API 已有但缺关系类型库 | 关系图谱语义不足 | `internal/model/entity.go:L54-L69`；`internal/handler/entity_handler.go:L266-L309`；`web/components/GraphVisualizer.tsx:L35-L149` |
| 质量门禁 | 质量接口存在但未与章节生成链路闭环 | 生成质量不可控 | `internal/router/router.go:L347-L352` |
| 导出链路 | 文件/排版存在，项目导出 JSON 已有；但未见“生成→排版→导出”闭环 | 无完整发布链路 | `internal/router/router.go:L181-L181`；`internal/handler/project_handler.go:L410-L442`；`internal/router/router.go:L329-L345` |

### 8.2 关键代码级发现

| 位置 | 发现 | 风险/影响 | 证据引用 |
|---|---|---|---|
| AI 代理 | 仅校验 path 是否含 `..`，无 allowlist（包含 workflow 入口） | 可能被滥用调用非预期路径 | `internal/handler/ai_proxy_handler.go:L42-L69`；`internal/handler/ai_proxy_stream_handler.go:L42-L69`；`internal/service/workflow_service.go:L98-L101` |
| SSE Hub | 无心跳/超时治理；仅在 client 断开时移除 | 长连接资源占用不可控 | `pkg/sse/sse.go:L52-L125` |
| 向导生成 | 前端向导生成本地 ID（Date.now），与后端 ID 不一致 | 前后端数据脱节风险 | `web/components/NovelWizard.tsx:L41-L99` |
| 插件文档 | 前端插件文档内置 Python 示例 | Go 复现目标有误导 | `web/components/PluginManager.tsx:L246-L260` |
| 章节写作 | 编辑器 AI 生成走前端调用，后端工作流未统一 | 生成链路难治理 | `web/components/Editor.tsx:L62-L132` |
| 会话状态 | 会话状态由前端“更新时间”推断 | 无明确状态机 | `web/components/WorkflowSessions.tsx:L54-L67` |
| 质量门禁文案 | 质量提示使用 `string(rune(minLength))` 拼接 | 提示字符显示错误 | `internal/service/quality_service.go:L71-L91` |
| 质量阈值 | 阈值为硬编码常量，且未从配置读取 | 配置不生效、环境难统一 | `internal/service/quality_service.go:L63-L77`；`internal/service/quality_service.go:L363-L378` |
| 排版服务 | 仅支持 tomato/standard 两种样式 | 可扩展性不足 | `internal/service/formatting_service.go:L25-L60` |
| 文件存储 | storage key 由 `userID + fileName` 构成 | 同名文件冲突风险 | `internal/service/file_service.go:L39-L110` |
| 异步任务 | 任务 worker 进程内队列 | 服务重启任务丢失 | `internal/service/job_service.go:L32-L110` |
| 插件异步执行 | 任务回填为 `step.appended`，规范未统一 | 工作流产物难复用 | `internal/service/job_service.go:L200-L257` |
| 文件列表 | 项目文件列表仅按项目查询，无归属校验 | 可能越权读取 | `internal/handler/file_handler.go:L186-L225`；`internal/service/file_service.go:L87-L88` |
| 文档列表 | 项目文档列表仅校验项目存在，不校验归属 | 可能越权读取项目文档 | `internal/handler/document_handler.go:L120-L145`；`internal/service/document_service.go:L92-L100` |
| 文档详情 | 文档详情未校验归属 | 可能越权读取文档内容 | `internal/handler/document_handler.go:L173-L189`；`internal/service/document_service.go:L83-L89` |
| 配置注入 | Quality/Formatting 初始化为零值配置 | 规则与阈值不稳定 | `internal/router/router.go:L101-L106` |
| 导出响应 | 项目导出不走统一响应格式 | 前端处理分支增加 | `internal/handler/project_handler.go:L404-L442` |
| 质量提示文案 | 数值转字符串使用 `string(rune(...))` | 文案错误风险 | `internal/service/quality_service.go:L71-L91` |

---

## 9. Go 复现建议（与 Python 计数法差异说明）

1. **计数口径以功能链路为准**：不按 Python 模块拆分；按我方 Go 服务与 API 组合覆盖“向导/章节/图谱/MCP/上下文”。
2. **将 MuMu“向导/章节生产线”拆为 Go Workflow 模块**：统一入口 + 多 step 配置 + SSE 事件标准化。
3. **MCP 插件规范用 Go 明确接口与回填机制**：以 `PluginCapability` 定义工具规范；AI 调用时注入工具列表。
4. **上下文构建用 Go 实现服务层抽象**：先规则/检索 MVP，后扩展向量检索。

---

## 10. 对标功能在 Go 的复现逻辑（模块级）

> 目标：按 MuMu 功能链路，用 Go 复现核心逻辑（不依赖 Python 实现细节）。

| 对标模块 | MuMu 功能要点 | Go 实现逻辑（建议） | 关键接口 | 证据引用 |
|---|---|---|---|---|
| 向导 SSE | 分步生成世界观/角色/大纲 | 以 `WorkflowDefinition` 描述 steps，执行后写入 `SessionStep`；SSE 广播 `step.appended` | `POST /api/v1/workflows/wizard` | MuMu §6.1；`internal/handler/workflow_handler.go:L27-L105` |
| 设定管理 | 角色/组织/设定结构化 | 在 `StoryEntity` 基础上扩展 schema；关键类型加专表或字段 | `/api/v1/projects/:id/entities` | MuMu §1；`internal/model/entity.go:L15-L69` |
| 章节生产线 | 生成/分析/重写/批量 | 新增 `chapter_generate`/`chapter_analyze`/`chapter_rewrite` workflow；落地文档 & 质量门禁 | `/api/v1/workflows/chapters/*` | MuMu §6.2；`internal/handler/document_handler.go:L25-L257` |
| MCP 插件 | 工具注入 + 自动调用 | AI 调用前注入 `ToolRegistry`；工具调用由 `JobService` 执行，回填 `SessionStep` | `/api/v1/plugins/*` + tool calls | MuMu §4；`internal/service/job_service.go:L19-L257` |
| 关系图谱 | 关系类型库 + 关系边 | 关系边已存在（实体链接）；关系类型库未见 | `/api/v1/entities/:id/links` | MuMu §5；`internal/model/entity.go:L54-L69`；`internal/handler/entity_handler.go:L266-L309` |
| 一致性约束 | Prompt/风格/记忆 | PromptTemplate + WritingStyle 模块；workflow 注入上下文 | `/api/v1/templates/*` | MuMu §6.3；`internal/handler/template_handler.go:L26-L188` |
| 上下文构建 | RTCO 分层 | 未实现上下文构建/向量检索服务，仅有语料库模块 | 内部服务调用 | MuMu §6.4；`internal/handler/corpus_handler.go:L23-L211` |
| 质量门禁 | 生成后质量门禁 | `quality_service.CheckQuality` 作为 workflow 后置步骤 | `/api/v1/quality/check` | MuMu §6.2；`internal/service/quality_service.go:L46-L137` |
| 排版/导出 | 生成后导出 | `formatting_service` + `file_service` 生成导出文件 | `/api/v1/formatting/format` `/api/v1/files/*` | MuMu §1；`internal/handler/formatting_handler.go:L21-L88` |

---

## 11. 优先模块：Go 复现草案（组件/接口/数据模型）

> 目标：按优先级模块给出 Go 层的落地结构（Handler/Service/Repo/Model），并复用现有接口能力。

### 11.1 向导（SSE）

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | `WorkflowHandler` 增加 `Wizard` 入口；`WorkflowService` 扩展 `WizardDefinition` 与 step 执行器；`SessionService` 落地 `SessionStep` | `internal/handler/workflow_handler.go:L27-L105`；`internal/service/workflow_service.go:L55-L191` |
| SSE 事件 | 复用 `step.appended` 与 `job.*` 事件；向导每步生成结果写入 `SessionStep` | `pkg/sse/sse.go:L52-L125`；`internal/service/job_service.go:L95-L257` |
| 接口草案 | `POST /api/v1/workflows/wizard`、`GET /api/v1/sse/stream` | `internal/router/router.go:L290-L302` |
| 数据模型 | `Session`/`SessionStep` 承载向导步骤产物，必要时在 `SessionStep.FormatType` 区分 `wizard_world`/`wizard_character`/`wizard_outline` | `internal/handler/session_handler.go:L23-L259` |

### 11.2 章节生产线

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | 新增 `ChapterWorkflowHandler` + `ChapterWorkflowService`；复用 `DocumentService` 写回章节文本 | `internal/handler/document_handler.go:L25-L257` |
| 接口草案 | `POST /api/v1/workflows/chapters/generate`、`/analyze`、`/rewrite`、`/batch` | MuMu §6.2 |
| 质量门禁 | 生成后调用 `QualityGateService.CheckQuality`；失败则返回质量原因 | `internal/service/quality_service.go:L29-L137` |
| 数据模型 | 章节与 `Document` 对应；生成历史挂 `SessionStep` | `internal/handler/document_handler.go:L25-L257` |

#### 11.2.1 章节工作流接口契约草案（generate/analyze/rewrite/batch）

目标：让前端不需要“猜”后端写回规则与步骤结构，能够统一以 Session + SSE 接入，且明确哪些字段需要在请求体中提供。

统一约定：
- 所有章节工作流都必须在 `Session` 中落档，SSE 订阅以 `session_id` 为准。
- `Session.Mode` 作为“工作流大类”；`SessionStep.FormatType` 作为“步骤类型”。
- `Document` 作为章节正文承载；分析/重写/批量的产物通过“Document 写回 + SessionStep 归档”同时提供。

##### A) `POST /api/v1/workflows/chapters/generate`

用途：生成新章节（可选覆盖已有文档）并写回 `Document.Content`。

请求体（草案）：
```json
{
  "project_id": 1,
  "session_id": 0,
  "document_id": 0,
  "volume_id": 0,
  "title": "第1章",
  "order_index": 1,

  "provider": "openai|gemini|...",
  "path": "v1/chat/completions",
  "body": "{\"model\":\"...\",\"messages\":[]}",

  "write_back": {
    "mode": "create_or_overwrite",
    "set_status": "草稿",
    "set_summary": true
  }
}
```

字段说明：
- `session_id`：可选；传入则续写到同一会话，否则创建新会话（与现有 workflow 行为一致）。
- `document_id`：可选；为 0 表示创建新 `Document`，否则表示覆盖/更新目标文档。
- `volume_id`：可选；仅在创建新文档时生效（对应 `Document.VolumeID`）。
- `order_index`：仅在创建新文档时要求；为减少前端复杂度，建议后端支持不传则自动 `max+1`（当前文档 CRUD 还未提供该能力）。
- `provider/path/body`：MVP 复用现有工作流的 AI 调用方式（`WorkflowService.callAI`）。
- `write_back`：声明写回策略，避免前端误判“生成结果在哪”。

响应（草案）：
```json
{
  "session": { "id": 1, "mode": "chapter_generate", "project_id": 1 },
  "document": { "id": 10, "project_id": 1, "title": "第1章", "content": "..." },
  "steps": [
    { "id": 100, "format_type": "chapter.generate.prompt", "content": "..." },
    { "id": 101, "format_type": "chapter.generate.result", "content": "..." }
  ],
  "content": "...",
  "raw": {}
}
```

SSE（最小）：
- 同一 `session_id` 下持续广播 `step.appended`，`content` 可作为 chunk 追加；最后一步广播包含完整结果。

##### B) `POST /api/v1/workflows/chapters/analyze`

用途：对章节做结构化分析（节奏/矛盾/人物一致性等），默认不覆盖正文，仅写回 `Document.Summary` 或写入 `SessionStep.Metadata`。

请求体（草案）：
```json
{
  "project_id": 1,
  "session_id": 0,
  "document_id": 10,

  "provider": "openai|gemini|...",
  "path": "v1/chat/completions",
  "body": "{\"model\":\"...\",\"messages\":[]}",

  "write_back": {
    "set_summary": true,
    "set_status": ""
  }
}
```

响应（草案）：
```json
{
  "session": { "id": 2, "mode": "chapter_analyze" },
  "document": { "id": 10, "summary": "..." },
  "analysis": {
    "highlights": [],
    "issues": [],
    "score": 0
  }
}
```

##### C) `POST /api/v1/workflows/chapters/rewrite`

用途：在指定约束下重写章节，写回 `Document.Content`；建议同时把“旧正文”存入 `SessionStep.Metadata.prev_content`，保证可回退。

请求体（草案）：
```json
{
  "project_id": 1,
  "session_id": 0,
  "document_id": 10,
  "rewrite_mode": "polish|expand|shorten|style_transfer",

  "provider": "openai|gemini|...",
  "path": "v1/chat/completions",
  "body": "{\"model\":\"...\",\"messages\":[]}",

  "write_back": {
    "mode": "overwrite",
    "set_status": "修改中"
  }
}
```

响应（草案）：
```json
{
  "session": { "id": 3, "mode": "chapter_rewrite" },
  "document": { "id": 10, "content": "..." },
  "diff": { "type": "none|unified", "text": "" }
}
```

##### D) `POST /api/v1/workflows/chapters/batch`

用途：批量生成多章（或基于大纲列表生成），每章写回一个 `Document`，并在同一会话内归档步骤。

请求体（草案）：
```json
{
  "project_id": 1,
  "session_id": 0,
  "volume_id": 0,
  "items": [
    { "title": "第1章", "order_index": 1, "outline": "..." },
    { "title": "第2章", "order_index": 2, "outline": "..." }
  ],

  "provider": "openai|gemini|...",
  "path": "v1/chat/completions",
  "body_template": "{\"model\":\"...\",\"messages\":[]}",

  "write_back": {
    "set_status": "草稿",
    "set_summary": true
  }
}
```

响应（草案）：
```json
{
  "session": { "id": 4, "mode": "chapter_batch" },
  "documents": [{ "id": 11 }, { "id": 12 }],
  "job_uuids": []
}
```

#### 11.2.2 Document 写回规则（建议固化为后端统一行为）

| 场景 | 写回字段 | 最小要求 | 说明 |
|---|---|---|---|
| generate | `Document.Content` | 必须写回 | 新章生成的“最终产物”应落 Document，避免只落 step 导致正文丢失 |
| analyze | `Document.Summary`（可选） | 至少落 Step | 若 Summary 用于列表摘要/检索，建议写回；否则只落 Step.Metadata 也可 |
| rewrite | `Document.Content` | 必须写回 | 建议将旧正文以 step 归档，避免无法回退 |
| batch | 多个 `Document` | 必须写回 | 每章都应落 Document；步骤用于过程追踪与复盘 |

#### 11.2.3 SessionStep / Job 归档结构（前端接入关键）

**Session.Mode 建议枚举：**
- `chapter_generate` / `chapter_analyze` / `chapter_rewrite` / `chapter_batch`

**SessionStep.FormatType 建议枚举（最小集合）：**
- `chapter.generate.prompt` / `chapter.generate.chunk` / `chapter.generate.result`
- `chapter.analyze.result`
- `chapter.rewrite.prompt` / `chapter.rewrite.chunk` / `chapter.rewrite.result`
- `chapter.batch.item.started` / `chapter.batch.item.result`

**SessionStep.Metadata（JSON）建议字段：**
```json
{
  "project_id": 1,
  "document_id": 10,
  "volume_id": 0,
  "provider": "openai",
  "path": "v1/chat/completions",
  "job_uuid": "",
  "progress": 0,
  "done": false
}
```

**Job 归档（适用于异步/工具调用/长链路）：**
- `Job.SessionID`：必须指向产生事件的会话（SSE 订阅维度）。
- `Job.ProjectID`：建议填充，用于审计/筛选。
- `Job.Payload`：建议存“工具参数/批次 item 参数”，不要存大段正文（正文落 Document/Step）。

### 11.3 MCP 插件

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | `PluginService` 提供 `ToolRegistry`（由 `PluginCapability` 组成）；AI 调用前注入工具列表 | `internal/model/plugin.go:L9-L55` |
| 任务执行 | 复用 `JobService` 进行异步调用与回填 `SessionStep` | `internal/service/job_service.go:L19-L257` |
| 接口草案 | `/api/v1/plugins`（配置） + `/api/v1/plugins/:id/invoke`（同步） + `/api/v1/jobs/*`（异步） | `internal/handler/plugin_handler.go:L52-L257`；`internal/handler/job_handler.go:L19-L65` |

#### 11.3.1 工具 Schema：PluginCapability → AI tools（最小闭环）

目标：将已存在的 `PluginCapability`（含 `input_schema/output_schema`）映射成模型可识别的“工具（tool/function）”，并可逆映射回 `(plugin_id, cap_id)` 以创建 Job。

**统一 Tool 命名规范（建议）**
- `tool_name = "plugin_{plugin_id}__{cap_id}"`
- 约束：只允许 `a-zA-Z0-9_`，避免不同模型对函数名的限制差异。

**ToolRegistry 返回结构（供 AIService 组装请求）**
```json
{
  "tools": [
    {
      "name": "plugin_12__summarize",
      "description": "对文本做摘要",
      "input_schema": { "type": "object", "properties": { "text": { "type": "string" } }, "required": ["text"] }
    }
  ],
  "index": {
    "plugin_12__summarize": { "plugin_id": 12, "cap_id": "summarize" }
  }
}
```

**不同模型的工具注入适配（仅口径，后端实现可渐进）**
- OpenAI 风格：`tools: [{type:"function", function:{name, description, parameters}}]`
- Gemini 风格：按其 function declarations 结构注入
- 统一：模型返回 `tool_calls` 后，后端按 `tool_name` 反查到 `(plugin_id, cap_id)` 并创建 Job 执行

#### 11.3.2 tool_calls 如何落 Job（对齐现有 JobService）

现有任务模型已经具备落盘字段（`jobs` 表 + SSE 广播 + 回填 Step）：
- `Job.Type = plugin_invoke`
- `Job.PluginID = plugin_id`
- `Job.Method = cap_id`（与插件 `/invoke` 的 `method` 对齐）
- `Job.Payload = tool_args`（直接作为插件 payload）
- `Job.SessionID = session_id`（SSE 订阅维度）
- `Job.ProjectID`：建议填充（若可从 Session 推导）

**工具调用创建 Job 的标准化 payload（建议）**
```json
{
  "tool_name": "plugin_12__summarize",
  "args": { "text": "..." },
  "context": {
    "project_id": 1,
    "document_id": 10,
    "session_id": 3,
    "step_id": 101
  }
}
```

落 Job 时建议规则：
- `Job.Payload` 只保存 `args` 与少量 `context`，避免把全文/大 JSON 塞进 jobs（正文落 Document/Step）。
- 多个 tool_calls 触发多个 Job，`job.created/job.progress/job.succeeded/job.failed` 通过 SSE 推送。

#### 11.3.3 回填到 SessionStep 的消息规范（让前端“可展示/可复盘”）

现状：Job 成功后会追加 `SessionStep`（`FormatType: plugin_result`）并广播 `step.appended`（`job_service.go`）。

建议规范化成 3 类 Step（最小集合）：
- `tool.call`：记录模型发起的工具调用（tool_name + args 摘要）
- `tool.result`：记录工具调用结果（可展示为结构化面板/原始 JSON）
- `tool.error`：记录失败原因（error_message）

建议的 `SessionStep.Metadata`（JSON）字段：
```json
{
  "tool_name": "plugin_12__summarize",
  "plugin_id": 12,
  "cap_id": "summarize",
  "job_uuid": "c0b7... ",
  "status": "queued|running|succeeded|failed",
  "document_id": 10
}
```

前端展示建议（对应 MuMu 的“生成历史/任务状态”体验）：
- Session 详情页以 Step 时间线为主（step.appended 实时追加）。
- 需要追踪异步状态时，以 `job.*` 事件驱动进度条，并最终落 `tool.result` step 作为可复盘产物。

### 11.4 关系图谱

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | 新增 `RelationshipType`/`EntityRelation` 模型与 Repo；Graph API 输出节点与关系 | `internal/model/entity.go:L54-L69` |
| 接口草案 | `/api/v1/relations/types`、`/api/v1/relations`、`/api/v1/graphs/:project_id` | MuMu §5 |
| 关联展示 | 前端复用 `GraphVisualizer`，关系类型用于边样式 | `web/components/GraphVisualizer.tsx:L35-L149` |

### 11.5 上下文构建

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | 新增 `ContextBuilderService`：规则检索 → 语料/实体 → 章节历史 → 向量检索（后置） | MuMu §6.4 |
| 数据输入 | `CorpusStory` + `Entity` + `Document` + `SessionStep` | `internal/handler/corpus_handler.go:L23-L211`；`internal/model/entity.go:L15-L69` |
| 产物输出 | 统一为 `PromptContext`（JSON）供 workflow 调用 | MuMu §6.4 |

### 11.6 质量门禁

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | `QualityGateService` 作为 workflow 后置步骤 | `internal/service/quality_service.go:L29-L137` |
| 接口草案 | `/api/v1/quality/check`、`/api/v1/quality/thresholds` | `internal/handler/formatting_handler.go:L53-L88` |
| 关键风险 | 文案使用 `string(rune(...))`，应改为 `strconv.Itoa`（避免字符错误） | `internal/service/quality_service.go:L71-L91` |

### 11.7 导出链路

| 项 | Go 复现建议 | 证据引用 |
|---|---|---|
| 组件结构 | `FormattingService` 生成排版文本 → `FileService` 存储 → `FileHandler` 下载 | `internal/service/formatting_service.go:L25-L159`；`internal/service/file_service.go:L39-L110` |
| 接口草案 | `/api/v1/formatting/format` → `/api/v1/files` → `/api/v1/files/:id/download` | `internal/handler/formatting_handler.go:L21-L88`；`internal/handler/file_handler.go:L23-L260` |
| 关键风险 | `storageKey` 以文件名拼接可能冲突，需加入时间/UUID | `internal/service/file_service.go:L39-L110` |

---

## 12. SSE 协议对标补齐（为“向导/章节长链路”做准备）

MuMu 的 SSE 事件更偏“进度机 + chunk 输出 + result/done”（§3.1），我方当前 SSE 更偏“事件广播（step.appended/quality/export）”。如果要复现 MuMu 的向导/章节体验，需要补齐以下约定（可逐步演进，不要求一次到位）：

| 目标能力 | 我方现状 | 建议补齐方式 | 备注 |
|---|---|---|---|
| 进度事件（progress） | 无统一 progress 事件 | 增加 `progress.updated` 事件（或在 `step.appended.metadata` 带 progress） | 用于前端进度条/状态提示 |
| 流式 chunk | 通过 step.appended 叠加 content | 保持 step.appended 作为 chunk 通道 | step_id 可作为聚合 key |
| result/done | 依赖最后一步输出 | 增加 `workflow.done`（或最后一步约定 `metadata.done=true`） | 用于前端自动收尾、解锁按钮 |
| heartbeat | 无 | 后端定时广播 `: heartbeat` 或 `event: ping` | 配合连接治理与监控 |

## 13. P0 验收脚本（可手动走通）

1. 注册/登录（拿到 JWT）。
2. 创建项目（Project）。
3. 启动向导工作流（世界观→角色→大纲），打开 session 详情页订阅 SSE，确认持续收到 step.appended。
4. 向导完成后，基于大纲启动“章节生成”工作流，生成第一章写回 Document。
5. 触发一次润色工作流，确认文档内容被更新并保存。
6. 触发质量检查与排版，确认接口返回结构稳定。
7. 导出项目或导出文件可下载（若导出暂缺，至少确认 Formatting→Files→Download 闭环）。
