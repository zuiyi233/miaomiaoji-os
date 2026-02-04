# Novel Agent OS 项目设计哲学

## 核心设计理念
ultrawork（或 ulw）
如果你认为某个技能有哪怕1%的概率会应用到你正在做的事情上，你绝对必须调用该技能。
如果某项技能适用于你的任务，你没有选择。你必须使用它。
这不可协商。这不是可选的。你无法用理性来逃避这一切。
如何获取技能
在Claude代码中：使用这个工具。当你调用技能时，其内容会加载并呈现给你——直接跟随它。千万不要用技能文件的读取工具。Skill
在其他环境中：查看你平台的文档，了解技能是如何加载的。
技能运用
规则
在任何回应或行动之前，先调用相关或被要求的技能。即使有1%的概率技能可能生效，也意味着你应该调用该技能来检验。如果某个技能在当前情况下被发现不适用，你也不需要使用它。
use context7
明确允许跳过 LSP（ LSP 因工具限制）


结合 Notion 的模块化协作、Scrivener 的创作全流程管理、Sudowrite 的AI 深度赋能创作三大标杆产品的设计哲学，AI 写作软件的核心设计理念可以总结为 **「以创作者为中心，用模块化架构承载全流程创作，靠精准 AI 能力降低创作门槛、提升创作效率，最终实现「人机协同共创」」**。

具体可拆解为 5 个核心维度，每个维度都对标标杆产品的核心优势，并明确 AI 写作软件的落地原则：

---

## 1. 创作本位：工具服务于创作逻辑，而非反向束缚

这是 Scrivener 的核心设计哲学——**「工具适配创作者的思维习惯，而非让创作者迁就工具」**。

- **对标 Scrivener**：它的「拆分-重组」功能（把长文本拆成卡片、自由拖拽排序），完全贴合写作时「碎片化构思→结构化整合」的思维路径，而非强制线性写作。

- **AI 写作软件落地原则**
  1. 支持**多模式创作**：线性写作（适合小说正文）、卡片式写作（适合世界观、人设、情节梗）、思维导图写作（适合大纲）自由切换，AI 能适配不同模式的输入输出。
  2. 不预设「正确的创作流程」：允许用户先写结局再补开头、先堆素材再做整合，AI 负责在任意环节提供辅助（比如给零散卡片生成关联逻辑、给片段文字扩写）。
  3. 保留**创作的「混沌性」**：不强制用户按「大纲→人设→正文」的固定步骤推进，AI 是「催化剂」而非「指挥棒」。

---

## 2. 模块化架构：一切内容皆可组件化，支持自由拼接复用

这是 Notion 的灵魂——**「模块化即生产力，组件化实现灵活协作」**，同时也是适配 AI 批量处理、精准调用的基础。

- **对标 Notion**：页面、数据库、块元素（文本、列表、图片）的自由组合，让用户可以搭建个性化的工作区，本质是「用最小单元的组件，承载无限的创作场景」。

- **AI 写作软件落地原则**
  1. **内容原子化**：把小说的「人设」「世界观设定」「情节节点」「金句」都拆成独立的**可复用组件**，AI 能识别组件标签（比如「#仙侠·门派设定」「#悬疑·反转梗」），按需调取组合。
  2. **结构可视化**：用「看板视图」「大纲视图」「时间线视图」展示模块化内容，AI 可以基于视图结构生成内容（比如按时间线补全情节空缺）。
  3. **协作轻量化**：支持多人对同一组件的编辑（比如和编辑讨论人设修改），AI 能记录修改痕迹并生成优化建议。

---

## 3. AI 赋能：精准嵌入创作环节，做「协作者」而非「代笔者」

这是 Sudowrite 的核心优势——**「AI 懂创作，而非只是懂文字」**，区别于普通的文本生成工具。

- **对标 Sudowrite**：它的「Show, Don't Tell」功能（把直白叙述改成场景描写）、「Plot Generator」（基于人物动机生成情节），都是**针对写作痛点的精准解决方案**，而非泛泛的文本扩写。

- **AI 写作软件落地原则**
  1. **场景化 AI 能力**：拒绝「大一统的生成按钮」，而是把 AI 功能拆解为创作各环节的「小工具」：
     - 构思期：人设生成、世界观推演、情节梗脑暴；
     - 写作期：文字润色、对话生成、场景描写补全；
     - 优化期：逻辑自查、节奏调整、风格统一。
  2. **可控的「人机分工」**：用户掌握**创作主导权**，AI 提供「备选方案」而非「最终答案」——比如用户写了一段对话，AI 可以生成 3 种不同语气的版本供选择。
  3. **学习用户风格**：AI 能通过分析用户的历史文本，适配其语言风格、叙事节奏，避免生成「千人一面」的内容。

---

## 4. 全流程闭环：从构思到输出，一站式承载创作生命周期

这是 Scrivener 区别于普通文本编辑器的关键——**「一个工具搞定写作全流程，无需在多个软件间切换」**。

- **对标 Scrivener**：从素材收集、大纲撰写、正文写作，到格式排版、导出发布，全部在一个界面完成，减少创作过程中的「注意力中断」。

- **AI 写作软件落地原则**
  1. **无缝衔接的创作链路**：集成「素材库→大纲工具→正文编辑器→校对工具→导出功能」，AI 在每个环节自动衔接——比如素材库的历史资料，AI 可以自动提炼关键信息导入大纲；正文写完后，AI 自动检查错别字、逻辑漏洞。
  2. **数据本地化与云端同步兼顾**：支持本地存储（保护创作隐私）和云端同步（跨设备写作），AI 模型可以在本地轻量化运行，也可以调用云端大模型获得更强能力。
  3. **适配多平台输出**：AI 自动根据输出需求（网文平台、出版、自媒体）调整格式、风格，比如网文版增加对话密度，出版版优化叙事节奏。

---

## 5. 轻量化扩展：低门槛上手，高上限定制

这是 Notion 与 Scrivener 的共同特点——**「新手友好，高手可用」**，既满足普通用户的基础需求，也能支撑专业创作者的复杂场景。

- **对标 Notion**：新手可以直接用模板快速起步，高手可以用数据库、API 搭建个性化工作流；对标 Scrivener：基础写作功能简单直观，专业功能（比如手稿统计、目标字数追踪）隐藏在高级设置里。

- **AI 写作软件落地原则**
  1. **分层级功能设计**：基础层（文本编辑、AI 润色）满足新手需求；进阶层（模块化管理、风格训练）满足专业创作者；专家层（API 接口、自定义 AI 提示词模板）满足团队协作或深度定制需求。
  2. **模板生态化**：内置不同体裁的创作模板（小说、剧本、散文、网文），支持用户自定义模板并分享，AI 可以基于模板生成适配内容。

---

## 核心设计理念的一句话总结

**「以创作者思维为锚点，用模块化架构搭建创作容器，让 AI 精准嵌入创作全流程，实现「用户主导、AI 辅助、全流程闭环」的人机协同写作」。**

这个理念的核心是**「平衡」**——平衡创作的自由性与工具的实用性，平衡 AI 的赋能与用户的主导权，平衡新手的易用性与专家的扩展性。
任何任务开始之前都必须使用技能

# RuleBack - AI代码生成规范

> **重要**: 这是AI生成代码的主规则文件。在生成任何代码之前，必须完整阅读本文件。

---

## 第一章：核心行为规范

### 1.1 AI必须遵守的行为准则

| 规则 | 说明 |
|------|------|
| 不生成表情符号 | 代码、注释、文档中禁止出现任何emoji |
| 不自动生成文档 | 完成代码后不要自动创建README或说明文档 |
| 不使用硬编码 | 关键配置必须写入配置文件，禁止在代码中硬编码 |
| 不生成测试用例 | 除非用户明确要求，否则不生成测试代码 |
| 更新API文档 | 每次修改接口后，更新docs目录下的API文档（不创建新文件） |
| 注释使用中文 | 默认使用中文编写注释，保持简洁明了 |
| 注释不要过多 | 注释简洁扼要，不使用装饰性分隔线 |
| 复用现有实例 | 不重复创建实例，复用已存在的单例 |
| 使用成熟框架 | 使用Gin、GORM等成熟框架，禁止造轮子 |

### 1.2 实例复用原则（重要）

**禁止重复创建实例，必须复用现有实例：**

```go
// 禁止：每次都创建新实例
func NewUserHandler() *UserHandler {
    return &UserHandler{
        service: service.NewUserService(),  // 每次都new，禁止
    }
}

// 正确：使用单例模式
var (
    userServiceInstance *UserService
    userServiceOnce     sync.Once
)

func GetUserService() *UserService {
    userServiceOnce.Do(func() {
        userServiceInstance = &UserService{
            repo: GetUserRepository(),
        }
    })
    return userServiceInstance
}
```

**全局实例必须使用sync.Once保护：**

```go
var (
    globalDB *gorm.DB
    dbOnce   sync.Once
)

func GetDB() *gorm.DB {
    dbOnce.Do(func() {
        // 初始化
    })
    return globalDB
}
```

### 1.3 禁止造轮子

**必须使用成熟的第三方库：**

| 功能 | 使用的库 | 禁止行为 |
|------|---------|---------|
| Web框架 | `github.com/gin-gonic/gin` | 自己实现HTTP路由 |
| ORM | `gorm.io/gorm` | 自己拼接SQL |
| 日志 | `go.uber.org/zap` | 使用fmt.Println |
| 配置 | `github.com/spf13/viper` | 自己解析配置文件 |
| UUID | `github.com/google/uuid` | 自己拼接时间戳生成ID |
| 密码加密 | `golang.org/x/crypto/bcrypt` | 自己实现加密算法 |
| JWT | `github.com/golang-jwt/jwt` | 自己实现token |
| 依赖注入 | `github.com/google/wire` | 手动管理复杂依赖 |

### 1.4 注释规范

**禁止使用装饰性分隔线注释：**
```go
// 禁止这样的注释
// =============================================================================
// 初始化函数
// =============================================================================

// 正确的注释方式
// UserService 用户业务逻辑层
type UserService struct {
    repo *repository.UserRepository
}
```

### 1.5 运行模式

项目支持两种运行模式：

| 模式 | 配置值 | 特点 |
|------|--------|------|
| 开发模式 | `app.env: development` | 详细日志、debug信息 |
| 生产模式 | `app.env: production` | 精简日志、性能优化 |

### 1.6 配置管理原则

**关键配置必须写入配置文件，禁止硬编码：**

```go
// 禁止硬编码
const JWTSecret = "my-secret-key"  // 禁止

// 正确方式
secret := config.Get().JWT.Secret
```

---

## 第二章：需求处理流程

### 2.1 接到需求后的处理步骤

```
步骤1: 理解需求 → 分析用户需求，明确要实现的功能
步骤2: 拆分任务 → 将需求拆分为具体的实现步骤
步骤3: 检查复用 → 检查是否有可复用的现有代码和实例
步骤4: 按模块实现 → Model → Repository → Service → Handler → Router
步骤5: 更新文档 → 在docs目录更新API文档
```

### 2.2 任务拆分示例

**需求**: 添加订单管理功能

**拆分后的任务：**
1. 在 `internal/model/order.go` 创建Order模型
2. 在 `internal/repository/order_repository.go` 创建数据访问层
3. 在 `internal/service/order_service.go` 创建业务逻辑层
4. 在 `internal/handler/order_handler.go` 创建HTTP处理器
5. 在 `internal/router/router.go` 注册路由
6. 在 `cmd/server/main.go` 添加数据库迁移
7. 更新 `docs/api.md` 的订单接口文档

---

## 第三章：绝对禁止的行为

| 禁止行为 | 正确做法 |
|---------|---------|
| 使用 `fmt.Println` 记录日志 | 使用 `logger.Info/Debug/Error` |
| 使用 `c.JSON` 返回响应 | 使用 `response.SuccessWithData/Fail` |
| 硬编码错误码数字 | 使用 `errors.CodeInvalidParams` |
| 硬编码配置值 | 从 `config.Get()` 读取 |
| 在Handler中写业务逻辑 | 业务逻辑放在Service层 |
| 在Service中直接操作数据库 | 通过Repository操作数据库 |
| 每次都New新实例 | 使用单例模式复用实例 |
| 自己实现已有功能 | 使用成熟的第三方库 |
| 自己拼接时间戳做ID | 使用uuid库 |
| 明文存储密码 | 使用bcrypt加密 |
| 使用装饰性分隔线注释 | 使用简洁的单行注释 |
| 自动生成测试用例 | 除非用户明确要求 |
| 自动创建文档文件 | 只更新现有文档 |

---

## 第四章：项目架构

### 4.1 目录结构

```
ruleback/
├── cmd/server/main.go          # 程序入口
├── configs/config.yaml         # 配置文件
├── docs/api.md                 # API文档
├── internal/
│   ├── config/                # 配置管理
│   ├── model/                 # 数据模型
│   ├── repository/            # 数据访问
│   ├── service/               # 业务逻辑
│   ├── handler/               # HTTP处理
│   ├── middleware/            # 中间件
│   ├── router/                # 路由配置
│   └── wire/                  # 依赖注入（Wire）
└── pkg/                       # 公共包
    ├── response/              # 统一响应
    ├── errors/                # 错误处理
    ├── logger/                # 日志记录
    └── database/              # 数据库连接
```

### 4.2 代码分层

```
Router → Handler → Service → Repository → Model
```

每层职责：
- **Router**: 路由注册、中间件应用
- **Handler**: 接收请求、参数验证、调用Service、返回响应
- **Service**: 业务逻辑处理、错误转换
- **Repository**: 数据库CRUD操作
- **Model**: 数据结构定义

---

## 第五章：实例管理规范

### 5.1 Wire依赖注入（推荐）

项目使用Google Wire进行编译时依赖注入，在 `internal/wire/` 目录管理：

```go
// 1. 每个模块提供New构造函数
func NewUserRepository(base *BaseRepository) *UserRepository {
    return &UserRepository{BaseRepository: base}
}

func NewUserService(repo *repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{service: svc}
}

// 2. 在wire/providers.go定义Provider
func ProvideUserRepository(base *repository.BaseRepository) *repository.UserRepository {
    return repository.NewUserRepository(base)
}

// 3. 在bootstrap.go使用Wire初始化
handlers, _ := wire.InitializeHandlers(database.GetDB())
r := router.Setup(handlers)
```

**添加新模块后必须重新生成Wire代码：**
```bash
~/go/bin/wire ./internal/wire/...
```

### 5.2 单例模式（向后兼容）

保留Get*方法用于向后兼容和简单场景：

```go
var (
    userRepoInstance *UserRepository
    userRepoOnce     sync.Once
)

func GetUserRepository() *UserRepository {
    userRepoOnce.Do(func() {
        userRepoInstance = &UserRepository{
            BaseRepository: GetBaseRepository(),
        }
    })
    return userRepoInstance
}
```

### 5.3 在路由中使用

```go
// 推荐：使用Wire注入的Handler
func registerUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler) {
    users := rg.Group("/users")
    {
        users.GET("", userHandler.List)
        users.POST("", userHandler.Create)
    }
}

// 向后兼容：使用Get单例
func registerUserRoutesLegacy(rg *gin.RouterGroup) {
    h := handler.GetUserHandler()
    // ...
}
```

---

## 第六章：已存在的功能（直接使用）

### 6.1 响应函数 (pkg/response)

```go
response.Success(c)
response.SuccessWithData(c, data)
response.SuccessWithPage(c, list, total, page, size)
response.Fail(c, errors.CodeXxx, "错误消息")
```

### 6.2 错误码 (pkg/errors)

```go
errors.CodeSuccess         // 0
errors.CodeInvalidParams   // 10001
errors.CodeUnauthorized    // 10002
errors.CodeNotFound        // 10004
errors.CodeDatabaseError   // 10007
```

### 6.3 日志函数 (pkg/logger)

```go
logger.Debug("消息", logger.String("key", "value"))
logger.Info("消息", logger.Uint("user_id", 1))
logger.Error("消息", logger.Err(err))
```

### 6.4 配置获取 (internal/config)

```go
cfg := config.Get()
cfg.App.Env
cfg.Server.Port
cfg.Database.Host
cfg.JWT.Secret
```

---

## 第七章：第三方库使用规范

### 7.1 必须使用的库

| 用途 | 库 | 导入路径 |
|------|-----|---------|
| Web框架 | Gin | `github.com/gin-gonic/gin` |
| ORM | GORM | `gorm.io/gorm` |
| 日志 | Zap | `go.uber.org/zap` |
| 配置 | Viper | `github.com/spf13/viper` |
| UUID | Google UUID | `github.com/google/uuid` |
| 密码 | Bcrypt | `golang.org/x/crypto/bcrypt` |
| 依赖注入 | Wire | `github.com/google/wire` |

### 7.2 使用示例

```go
// UUID生成
import "github.com/google/uuid"
id := uuid.New().String()

// 密码加密
import "golang.org/x/crypto/bcrypt"
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 密码验证
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
```

---

## 第八章：规则文件索引

| 模块 | 规则文件 |
|------|---------|
| 配置 | `internal/config/RULE.md` |
| 模型 | `internal/model/RULE.md` |
| Repository | `internal/repository/RULE.md` |
| Service | `internal/service/RULE.md` |
| Handler | `internal/handler/RULE.md` |
| 中间件 | `internal/middleware/RULE.md` |
| 路由 | `internal/router/RULE.md` |
| Wire依赖注入 | `internal/wire/RULE.md` |
| 响应 | `pkg/response/RULE.md` |
| 错误 | `pkg/errors/RULE.md` |
| 日志 | `pkg/logger/RULE.md` |
| 文档 | `docs/RULE.md` |

---

## 第九章：检查清单

### 生成代码前

- [ ] 已理解用户需求
- [ ] 已拆分为具体步骤
- [ ] 已检查是否有可复用的代码
- [ ] 已检查是否有成熟的库可用

### 生成代码时

- [ ] 使用单例模式管理实例（Get而不是New）
- [ ] 使用成熟的第三方库
- [ ] 使用response包返回响应
- [ ] 使用errors包的错误码
- [ ] 使用logger包记录日志
- [ ] 从config读取配置
- [ ] 注释简洁，使用中文
- [ ] 不使用装饰性分隔线

### 生成代码后

- [ ] 更新docs/api.md
- [ ] 不自动创建新文档（除非用户要求）
- [ ] 不自动生成测试用例（除非用户要求）

---

**记住：复用、简洁、规范。**
