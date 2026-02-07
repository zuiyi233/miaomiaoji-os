# AI功能全量测试脚本
# 测试环境: http://localhost:8080
# API: http://192.168.32.15:39999
# Model: glm-4.7

$BaseUrl = "http://localhost:8080"
$TestResults = @()

function Write-TestHeader($title) {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "  $title" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
}

function Write-TestResult($name, $success, $details) {
    $status = if ($success) { "✓ PASS" } else { "✗ FAIL" }
    $color = if ($success) { "Green" } else { "Red" }
    Write-Host "[$status] $name" -ForegroundColor $color
    if ($details) {
        Write-Host "  $details" -ForegroundColor Gray
    }
    return @{
        Name = $name
        Success = $success
        Details = $details
        Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    }
}

function Invoke-ApiRequest($method, $endpoint, $body = $null, $token = $null) {
    $headers = @{ "Content-Type" = "application/json" }
    if ($token) {
        $headers["Authorization"] = "Bearer $token"
    }
    
    try {
        $params = @{
            Uri = "$BaseUrl$endpoint"
            Method = $method
            Headers = $headers
        }
        if ($body) {
            $params["Body"] = ($body | ConvertTo-Json -Depth 10)
        }
        
        $response = Invoke-RestMethod @params -TimeoutSec 120
        return @{ Success = $true; Data = $response }
    } catch {
        return @{ Success = $false; Error = $_.Exception.Message; StatusCode = $_.Exception.Response.StatusCode }
    }
}

# ============ 测试1: 服务健康检查 ============
Write-TestHeader "测试1: 服务健康检查"

$health = Invoke-ApiRequest -method "GET" -endpoint "/healthz"
if ($health.Success -and $health.Data.data.status -eq "ok") {
    $TestResults += Write-TestResult "健康检查" $true "服务运行正常"
} else {
    $TestResults += Write-TestResult "健康检查" $false $health.Error
}

# ============ 测试2: 用户注册和登录 ============
Write-TestHeader "测试2: 用户认证"

# 注册用户
$registerBody = @{
    username = "testuser_$(Get-Random)"
    password = "Test123456"
    nickname = "测试用户"
}

$register = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/auth/register" -body $registerBody
if ($register.Success) {
    $TestResults += Write-TestResult "用户注册" $true "用户名: $($registerBody.username)"
} else {
    $TestResults += Write-TestResult "用户注册" $false $register.Error
}

# 登录用户
$loginBody = @{
    username = $registerBody.username
    password = $registerBody.password
}

$login = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/auth/login" -body $loginBody
$token = $null
if ($login.Success -and $login.Data.data.token) {
    $token = $login.Data.data.token
    $TestResults += Write-TestResult "用户登录" $true "Token获取成功"
} else {
    $TestResults += Write-TestResult "用户登录" $false $login.Error
}

if (-not $token) {
    Write-Host "`n无法获取Token，终止测试" -ForegroundColor Red
    exit 1
}

# ============ 测试3: AI配置管理 ============
Write-TestHeader "测试3: AI配置管理"

# 更新供应商配置
$providerBody = @{
    provider = "zhipu"
    base_url = "http://192.168.32.15:39999"
    api_key = "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B"
}

$updateProvider = Invoke-ApiRequest -method "PUT" -endpoint "/api/v1/ai/providers" -body $providerBody -token $token
if ($updateProvider.Success -and $updateProvider.Data.code -eq 0) {
    $TestResults += Write-TestResult "更新供应商配置" $true "Provider: zhipu"
} else {
    $TestResults += Write-TestResult "更新供应商配置" $false $updateProvider.Error
}

# 获取供应商配置
$getProvider = Invoke-ApiRequest -method "GET" -endpoint "/api/v1/ai/providers?provider=zhipu" -token $token
if ($getProvider.Success -and $getProvider.Data.data.provider -eq "zhipu") {
    $TestResults += Write-TestResult "获取供应商配置" $true "BaseURL: $($getProvider.Data.data.base_url)"
} else {
    $TestResults += Write-TestResult "获取供应商配置" $false $getProvider.Error
}

# 测试供应商连接
$testBody = @{
    provider = "zhipu"
}

$testProvider = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/ai/providers/test" -body $testBody -token $token
if ($testProvider.Success -and $testProvider.Data.code -eq 0) {
    $TestResults += Write-TestResult "测试供应商连接" $true "连接成功"
} else {
    $TestResults += Write-TestResult "测试供应商连接" $false "$($testProvider.Error) / $($testProvider.Data.message)"
}

# ============ 测试4: AI模型列表 ============
Write-TestHeader "测试4: AI模型列表"

$models = Invoke-ApiRequest -method "GET" -endpoint "/api/v1/ai/models?provider=zhipu" -token $token
if ($models.Success -and $models.Data.data.Count -gt 0) {
    $modelList = $models.Data.data | ForEach-Object { $_.id } | Select-Object -First 5
    $TestResults += Write-TestResult "获取模型列表" $true "找到 $($models.Data.data.Count) 个模型"
    Write-Host "  模型列表: $($modelList -join ', ')" -ForegroundColor Gray
} else {
    $TestResults += Write-TestResult "获取模型列表" $false $models.Error
}

# ============ 测试5: AI代理普通请求 ============
Write-TestHeader "测试5: AI代理普通请求"

$chatBody = @{
    provider = "zhipu"
    path = "/v1/chat/completions"
    body = (@{
        model = "glm-4.7"
        messages = @(
            @{ role = "system"; content = "你是一个 helpful assistant." }
            @{ role = "user"; content = "你好，请简单介绍一下自己" }
        )
        temperature = 0.7
        max_tokens = 200
    } | ConvertTo-Json -Depth 10 -Compress)
}

Write-Host "  发送请求到 /v1/chat/completions..." -ForegroundColor Yellow
$chatResponse = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/ai/proxy" -body $chatBody -token $token

if ($chatResponse.Success) {
    $content = $chatResponse.Data.choices[0].message.content
    $model = $chatResponse.Data.model
    $TestResults += Write-TestResult "AI普通对话" $true "Model: $model, 响应长度: $($content.Length)"
    Write-Host "  AI响应: $($content.Substring(0, [Math]::Min(100, $content.Length)))..." -ForegroundColor Gray
} else {
    $TestResults += Write-TestResult "AI普通对话" $false $chatResponse.Error
}

# ============ 测试6: AI代理流式请求 ============
Write-TestHeader "测试6: AI代理流式请求"

$streamBody = @{
    provider = "zhipu"
    path = "/v1/chat/completions"
    body = (@{
        model = "glm-4.7"
        messages = @(
            @{ role = "user"; content = "用一句话描述春天的美丽" }
        )
        stream = $true
        max_tokens = 100
    } | ConvertTo-Json -Depth 10 -Compress)
}

Write-Host "  发送流式请求..." -ForegroundColor Yellow
try {
    $headers = @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $token"
    }
    
    $response = Invoke-WebRequest -Uri "$BaseUrl/api/v1/ai/proxy/stream" -Method POST -Headers $headers -Body ($streamBody | ConvertTo-Json -Depth 10) -TimeoutSec 60
    
    if ($response.StatusCode -eq 200) {
        $TestResults += Write-TestResult "AI流式对话" $true "流式响应接收成功"
    } else {
        $TestResults += Write-TestResult "AI流式对话" $false "状态码: $($response.StatusCode)"
    }
} catch {
    $TestResults += Write-TestResult "AI流式对话" $false $_.Exception.Message
}

# ============ 测试7: 工作流测试 ============
Write-TestHeader "测试7: 工作流功能"

# 先创建一个项目
$projectBody = @{
    title = "测试项目_$(Get-Random)"
    description = "AI测试项目"
}

$project = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/projects/" -body $projectBody -token $token
$projectId = $null
if ($project.Success) {
    $projectId = $project.Data.data.id
    $TestResults += Write-TestResult "创建项目" $true "ProjectID: $projectId"
} else {
    $TestResults += Write-TestResult "创建项目" $false $project.Error
}

if ($projectId) {
    # 测试世界构建工作流
    $worldBody = @{
        project_id = $projectId
        session_title = "世界构建测试"
        prompt = "创建一个简单的奇幻世界设定，包含世界名称和基本描述"
        provider = "zhipu"
        model = "glm-4.7"
    }
    
    Write-Host "  执行世界构建工作流..." -ForegroundColor Yellow
    $worldFlow = Invoke-ApiRequest -method "POST" -endpoint "/api/v1/workflows/world" -body $worldBody -token $token
    
    if ($worldFlow.Success -and $worldFlow.Data.code -eq 0) {
        $TestResults += Write-TestResult "世界构建工作流" $true "SessionID: $($worldFlow.Data.data.session.id)"
    } else {
        $TestResults += Write-TestResult "世界构建工作流" $false "$($worldFlow.Error) / $($worldFlow.Data.message)"
    }
}

# ============ 测试8: 错误处理测试 ============
Write-TestHeader "测试8: 错误处理测试"

# 测试无效供应商
$invalidProvider = Invoke-ApiRequest -method "GET" -endpoint "/api/v1/ai/models?provider=nonexistent" -token $token
if (-not $invalidProvider.Success -or $invalidProvider.Data.code -ne 0) {
    $TestResults += Write-TestResult "无效供应商处理" $true "正确返回错误"
} else {
    $TestResults += Write-TestResult "无效供应商处理" $false "应该返回错误"
}

# 测试未授权访问
$unauthorized = Invoke-ApiRequest -method "GET" -endpoint "/api/v1/ai/models?provider=zhipu"
if (-not $unauthorized.Success) {
    $TestResults += Write-TestResult "未授权访问处理" $true "正确拒绝未授权请求"
} else {
    $TestResults += Write-TestResult "未授权访问处理" $false "应该拒绝未授权请求"
}

# ============ 测试报告 ============
Write-TestHeader "测试报告汇总"

$total = $TestResults.Count
$passed = ($TestResults | Where-Object { $_.Success }).Count
$failed = $total - $passed

Write-Host "总测试数: $total" -ForegroundColor White
Write-Host "通过: $passed" -ForegroundColor Green
Write-Host "失败: $failed" -ForegroundColor Red
Write-Host "成功率: $([math]::Round($passed/$total*100, 2))%" -ForegroundColor Cyan

Write-Host "`n详细结果:" -ForegroundColor Yellow
$TestResults | ForEach-Object {
    $icon = if ($_.Success) { "✓" } else { "✗" }
    $color = if ($_.Success) { "Green" } else { "Red" }
    Write-Host "  $icon $($_.Name) - $($_.Timestamp)" -ForegroundColor $color
}

# 保存测试报告
$reportPath = "test/ai_test/test_report_$(Get-Date -Format 'yyyyMMdd_HHmmss').json"
$TestResults | ConvertTo-Json -Depth 10 | Out-File -FilePath $reportPath -Encoding UTF8
Write-Host "`n测试报告已保存到: $reportPath" -ForegroundColor Cyan
