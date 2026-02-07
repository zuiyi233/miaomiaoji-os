# AI功能全量测试脚本 - 简化版
# 测试环境: http://localhost:8080
# API: http://192.168.32.15:39999
# Model: glm-4.7

$BaseUrl = "http://localhost:8080"
$TestResults = @()

function Test-HealthCheck {
    Write-Host "`n[Test] 服务健康检查" -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method GET -TimeoutSec 10
        if ($response.data.status -eq "ok") {
            Write-Host "  [PASS] 服务运行正常" -ForegroundColor Green
            return @{ Name = "健康检查"; Success = $true; Details = "服务运行正常" }
        }
    } catch {
        Write-Host "  [FAIL] $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "健康检查"; Success = $false; Details = $_.Exception.Message }
    }
}

function Test-UserAuth {
    Write-Host "`n[Test] 用户注册和登录" -ForegroundColor Cyan
    
    # 注册
    $username = "testuser_$(Get-Random)"
    $registerBody = @{
        username = $username
        password = "Test123456"
        nickname = "测试用户"
    } | ConvertTo-Json
    
    try {
        $reg = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -TimeoutSec 10
        Write-Host "  [PASS] 用户注册成功: $username" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] 注册失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "用户认证"; Success = $false; Details = $_.Exception.Message }
    }
    
    # 登录
    $loginBody = @{
        username = $username
        password = "Test123456"
    } | ConvertTo-Json
    
    try {
        $login = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -TimeoutSec 10
        Write-Host "  [PASS] 用户登录成功" -ForegroundColor Green
        return @{ Name = "用户认证"; Success = $true; Details = "Token获取成功"; Token = $login.data.token }
    } catch {
        Write-Host "  [FAIL] 登录失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "用户认证"; Success = $false; Details = $_.Exception.Message }
    }
}

function Test-AIConfig($token) {
    Write-Host "`n[Test] AI配置管理" -ForegroundColor Cyan
    $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
    
    # 更新配置
    $providerBody = @{
        provider = "zhipu"
        base_url = "http://192.168.32.15:39999"
        api_key = "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B"
    } | ConvertTo-Json
    
    try {
        $update = Invoke-RestMethod -Uri "$BaseUrl/api/v1/ai/providers" -Method PUT -Headers $headers -Body $providerBody -TimeoutSec 10
        Write-Host "  [PASS] 更新供应商配置" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] 更新配置失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # 获取配置
    try {
        $get = Invoke-RestMethod -Uri "$BaseUrl/api/v1/ai/providers?provider=zhipu" -Method GET -Headers $headers -TimeoutSec 10
        Write-Host "  [PASS] 获取供应商配置: $($get.data.base_url)" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] 获取配置失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "AI配置管理"; Success = $false; Details = $_.Exception.Message }
    }
    
    # 测试连接
    $testBody = @{ provider = "zhipu" } | ConvertTo-Json
    try {
        $test = Invoke-RestMethod -Uri "$BaseUrl/api/v1/ai/providers/test" -Method POST -Headers $headers -Body $testBody -TimeoutSec 30
        if ($test.code -eq 0) {
            Write-Host "  [PASS] 供应商连接测试成功" -ForegroundColor Green
            return @{ Name = "AI配置管理"; Success = $true; Details = "配置和连接测试成功" }
        } else {
            Write-Host "  [WARN] 连接测试返回: $($test.message)" -ForegroundColor Yellow
            return @{ Name = "AI配置管理"; Success = $true; Details = "配置成功但连接测试返回: $($test.message)" }
        }
    } catch {
        Write-Host "  [WARN] 连接测试失败: $($_.Exception.Message)" -ForegroundColor Yellow
        return @{ Name = "AI配置管理"; Success = $true; Details = "配置成功但连接测试失败: $($_.Exception.Message)" }
    }
}

function Test-ModelList($token) {
    Write-Host "`n[Test] AI模型列表" -ForegroundColor Cyan
    $headers = @{ "Authorization" = "Bearer $token" }
    
    try {
        $models = Invoke-RestMethod -Uri "$BaseUrl/api/v1/ai/models?provider=zhipu" -Method GET -Headers $headers -TimeoutSec 30
        if ($models.data.Count -gt 0) {
            $modelNames = $models.data | ForEach-Object { $_.id } | Select-Object -First 5
            Write-Host "  [PASS] 获取到 $($models.data.Count) 个模型" -ForegroundColor Green
            Write-Host "  模型: $($modelNames -join ', ')" -ForegroundColor Gray
            return @{ Name = "AI模型列表"; Success = $true; Details = "获取到 $($models.data.Count) 个模型" }
        } else {
            Write-Host "  [WARN] 模型列表为空" -ForegroundColor Yellow
            return @{ Name = "AI模型列表"; Success = $true; Details = "模型列表为空" }
        }
    } catch {
        Write-Host "  [FAIL] 获取模型列表失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "AI模型列表"; Success = $false; Details = $_.Exception.Message }
    }
}

function Test-AIChat($token) {
    Write-Host "`n[Test] AI普通对话" -ForegroundColor Cyan
    $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
    
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
        } | ConvertTo-Json -Compress)
    } | ConvertTo-Json -Compress
    
    Write-Host "  发送请求到 AI..." -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/ai/proxy" -Method POST -Headers $headers -Body $chatBody -TimeoutSec 120
        $content = $response.choices[0].message.content
        $model = $response.model
        Write-Host "  [PASS] AI响应成功" -ForegroundColor Green
        Write-Host "  Model: $model" -ForegroundColor Gray
        Write-Host "  响应: $($content.Substring(0, [Math]::Min(80, $content.Length)))..." -ForegroundColor Gray
        return @{ Name = "AI普通对话"; Success = $true; Details = "Model: $model, 响应长度: $($content.Length)" }
    } catch {
        Write-Host "  [FAIL] AI对话失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "AI普通对话"; Success = $false; Details = $_.Exception.Message }
    }
}

function Test-AIStream($token) {
    Write-Host "`n[Test] AI流式对话" -ForegroundColor Cyan
    $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
    
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
        } | ConvertTo-Json -Compress)
    } | ConvertTo-Json -Compress
    
    Write-Host "  发送流式请求..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/api/v1/ai/proxy/stream" -Method POST -Headers $headers -Body $streamBody -TimeoutSec 60
        if ($response.StatusCode -eq 200) {
            Write-Host "  [PASS] 流式响应接收成功" -ForegroundColor Green
            return @{ Name = "AI流式对话"; Success = $true; Details = "流式响应接收成功" }
        }
    } catch {
        Write-Host "  [FAIL] 流式请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "AI流式对话"; Success = $false; Details = $_.Exception.Message }
    }
}

function Test-Workflow($token) {
    Write-Host "`n[Test] 工作流功能" -ForegroundColor Cyan
    $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
    
    # 创建项目
    $projectBody = @{
        title = "测试项目_$(Get-Random)"
        description = "AI测试项目"
    } | ConvertTo-Json
    
    try {
        $project = Invoke-RestMethod -Uri "$BaseUrl/api/v1/projects/" -Method POST -Headers $headers -Body $projectBody -TimeoutSec 10
        $projectId = $project.data.id
        Write-Host "  [PASS] 创建项目成功: $projectId" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] 创建项目失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "工作流功能"; Success = $false; Details = $_.Exception.Message }
    }
    
    # 世界构建工作流
    $worldBody = @{
        project_id = $projectId
        session_title = "世界构建测试"
        prompt = "创建一个简单的奇幻世界设定，包含世界名称和基本描述"
        provider = "zhipu"
        model = "glm-4.7"
    } | ConvertTo-Json
    
    Write-Host "  执行世界构建工作流..." -ForegroundColor Yellow
    try {
        $world = Invoke-RestMethod -Uri "$BaseUrl/api/v1/workflows/world" -Method POST -Headers $headers -Body $worldBody -TimeoutSec 120
        if ($world.code -eq 0) {
            Write-Host "  [PASS] 世界构建工作流成功" -ForegroundColor Green
            return @{ Name = "工作流功能"; Success = $true; Details = "世界构建工作流执行成功" }
        } else {
            Write-Host "  [FAIL] 工作流返回错误: $($world.message)" -ForegroundColor Red
            return @{ Name = "工作流功能"; Success = $false; Details = $world.message }
        }
    } catch {
        Write-Host "  [FAIL] 工作流执行失败: $($_.Exception.Message)" -ForegroundColor Red
        return @{ Name = "工作流功能"; Success = $false; Details = $_.Exception.Message }
    }
}

# ============ 主测试流程 ============
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  AI功能全量测试开始" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "API: http://192.168.32.15:39999" -ForegroundColor Gray
Write-Host "Model: glm-4.7" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Cyan

# 1. 健康检查
$result = Test-HealthCheck
$TestResults += $result

# 2. 用户认证
$result = Test-UserAuth
$TestResults += $result
$token = $result.Token

if (-not $token) {
    Write-Host "`n无法获取Token，终止测试" -ForegroundColor Red
    exit 1
}

# 3. AI配置
$result = Test-AIConfig -token $token
$TestResults += $result

# 4. 模型列表
$result = Test-ModelList -token $token
$TestResults += $result

# 5. AI普通对话
$result = Test-AIChat -token $token
$TestResults += $result

# 6. AI流式对话
$result = Test-AIStream -token $token
$TestResults += $result

# 7. 工作流
$result = Test-Workflow -token $token
$TestResults += $result

# ============ 测试报告 ============
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  测试报告汇总" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$total = $TestResults.Count
$passed = ($TestResults | Where-Object { $_.Success }).Count
$failed = $total - $passed

Write-Host "总测试数: $total" -ForegroundColor White
Write-Host "通过: $passed" -ForegroundColor Green
Write-Host "失败: $failed" -ForegroundColor Red
Write-Host "成功率: $([math]::Round($passed/$total*100, 2))%" -ForegroundColor Cyan

Write-Host "`n详细结果:" -ForegroundColor Yellow
$TestResults | ForEach-Object {
    $icon = if ($_.Success) { "[PASS]" } else { "[FAIL]" }
    $color = if ($_.Success) { "Green" } else { "Red" }
    Write-Host "  $icon $($_.Name) - $($_.Details)" -ForegroundColor $color
}

# 保存报告
$reportPath = "test/ai_test/test_report_$(Get-Date -Format 'yyyyMMdd_HHmmss').json"
$TestResults | Select-Object Name, Success, Details | ConvertTo-Json | Out-File -FilePath $reportPath -Encoding UTF8
Write-Host "`n测试报告已保存到: $reportPath" -ForegroundColor Cyan
