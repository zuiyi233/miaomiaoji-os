# AI Provider 配置安全指南

## 概述

本指南说明如何安全地配置 AI Provider 的 API 密钥，以及如何控制内网 HTTP 访问权限。

## API 密钥安全配置

### 1. 使用环境变量（推荐）

为了避免在配置文件中暴露明文 API 密钥，建议使用环境变量：

#### 配置文件格式

在 `configs/providers/<provider>.yaml` 中使用环境变量占位符：

```yaml
provider: zhipu
base_url: http://192.168.32.15:39999/v1
api_key: ${ZHIPU_API_KEY}
models_cache:
  - model-1
  - model-2
```

#### 设置环境变量

**Windows (PowerShell)**:
```powershell
$env:ZHIPU_API_KEY="your-actual-api-key-here"
```

**Windows (CMD)**:
```cmd
set ZHIPU_API_KEY=your-actual-api-key-here
```

**Linux/Mac**:
```bash
export ZHIPU_API_KEY="your-actual-api-key-here"
```

#### 环境变量优先级

系统会按以下顺序查找 API 密钥（以 zhipu 为例）：

1. `NOVEL_AGENT_OS_ZHIPU_API_KEY`
2. `ZHIPU_API_KEY`
3. `GOOGLE_API_KEY`（通用后备）
4. `GEMINI_API_KEY`（通用后备）
5. 配置文件中的 `api_key` 字段

### 2. 配置文件示例

项目提供了 `.example` 模板文件，不包含真实密钥：

- `configs/providers/zhipu.yaml.example` - 模板文件
- `configs/providers/zhipu.yaml` - 实际配置（不应提交到版本控制）

建议在 `.gitignore` 中添加：
```
configs/providers/*.yaml
!configs/providers/*.yaml.example
```

### 3. 明文密钥警告

如果配置文件中包含明文 API 密钥（非环境变量占位符），系统会在启动时输出警告日志：

```
[WARN] 配置文件中包含明文API密钥，建议使用环境变量 provider=zhipu file=configs/providers/zhipu.yaml
```

## 内网 HTTP 访问控制

### 环境判断机制

系统根据运行环境自动调整 HTTP 访问策略：

#### 开发环境 (`app.env: development`)

- 允许访问内网 HTTP 地址（如 `http://192.168.*`、`http://10.*`）
- 允许访问 localhost HTTP（`http://localhost`、`http://127.0.0.1`）
- 仍然推荐使用 HTTPS

#### 生产环境 (`app.env: production`)

- **仅允许** localhost HTTP（`http://localhost`、`http://127.0.0.1`、`http://[::1]`）
- 内网 HTTP 地址会被拒绝
- 强制使用 HTTPS 访问外部服务

### 配置项说明

在 `configs/config.yaml` 中：

```yaml
app:
  env: development  # 或 production

ai:
  default_provider: gemini
  providers_path: "./configs/providers"
  allow_insecure_http: false  # 显式允许不安全的HTTP（慎用）
```

#### `allow_insecure_http` 配置

- `false`（默认）：遵循环境判断机制
- `true`：即使在生产环境也允许内网 HTTP（**不推荐**）

### 内网地址判断

系统使用 Go 标准库的 `net.IP.IsPrivate()` 方法判断内网地址，包括：

- `10.0.0.0/8`
- `172.16.0.0/12`
- `192.168.0.0/16`
- `127.0.0.0/8`（localhost）
- `169.254.0.0/16`（链路本地）
- IPv6 私有地址

## 安全最佳实践

### 1. 密钥管理

- ✅ 使用环境变量存储 API 密钥
- ✅ 将 `*.yaml` 配置文件添加到 `.gitignore`
- ✅ 仅提交 `*.yaml.example` 模板文件
- ❌ 不要在配置文件中硬编码密钥
- ❌ 不要将包含密钥的配置文件提交到版本控制

### 2. 网络访问

- ✅ 生产环境使用 HTTPS
- ✅ 开发环境可以使用内网 HTTP
- ✅ 设置 `app.env: production` 启用严格模式
- ❌ 不要在生产环境设置 `allow_insecure_http: true`
- ❌ 不要将生产密钥用于开发环境

### 3. 配置审计

定期检查配置文件：

```bash
# 检查是否有明文密钥
grep -r "api_key:" configs/providers/*.yaml

# 检查环境变量占位符
grep -r "\${.*}" configs/providers/*.yaml
```

## 故障排查

### 问题：API 调用失败，提示 "invalid base url"

**原因**：生产环境禁止访问内网 HTTP 地址

**解决方案**：
1. 检查 `configs/config.yaml` 中的 `app.env` 设置
2. 如果是开发环境，设置为 `development`
3. 如果是生产环境，将 provider 的 `base_url` 改为 HTTPS

### 问题：环境变量未生效

**原因**：环境变量未正确设置或应用未重启

**解决方案**：
1. 确认环境变量已设置：`echo $ZHIPU_API_KEY`（Linux/Mac）或 `echo %ZHIPU_API_KEY%`（Windows）
2. 重启应用以加载新的环境变量
3. 检查环境变量名称是否正确（大写，下划线分隔）

### 问题：仍然看到明文密钥警告

**原因**：配置文件中仍有明文密钥

**解决方案**：
1. 将 `api_key: sk-xxx` 改为 `api_key: ${PROVIDER_API_KEY}`
2. 设置对应的环境变量
3. 重启应用

## 配置示例

### 完整的安全配置示例

**configs/config.yaml**:
```yaml
app:
  env: production
  name: miaomiaoji-os
  debug: false

ai:
  default_provider: gemini
  providers_path: "./configs/providers"
  allow_insecure_http: false
```

**configs/providers/zhipu.yaml**:
```yaml
provider: zhipu
base_url: https://api.zhipuai.cn/v1
api_key: ${ZHIPU_API_KEY}
models_cache:
  - glm-4.6
  - glm-4.7
updated_at: "2026-02-07T13:00:00Z"
```

**环境变量设置**:
```bash
export ZHIPU_API_KEY="your-production-api-key"
export NOVEL_AGENT_OS_APP_ENV="production"
```

## 相关文档

- [API 文档](./api.md)
- [配置文件说明](../configs/config.yaml)
- [Provider 配置模板](../configs/providers/)
