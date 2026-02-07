$baseUrl = "http://localhost:8080"
$token = "YOUR_TOKEN_HERE"

# First login to get token
$loginBody = @{
    username = "admin"
    password = "admin"
} | ConvertTo-Json

try {
    $login = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $login.data.token
    Write-Host "Login success, token obtained" -ForegroundColor Green
} catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test AI Proxy
Write-Host "`n=== Testing AI Proxy ===" -ForegroundColor Cyan

$chatBody = @{
    provider = "zhipu"
    path = "/v1/chat/completions"
    body = (@{
        model = "glm-4.7"
        messages = @(
            @{ role = "system"; content = "You are a helpful assistant." }
            @{ role = "user"; content = "Hello" }
        )
        temperature = 0.7
        max_tokens = 100
    } | ConvertTo-Json -Compress)
} | ConvertTo-Json -Compress

Write-Host "Request body: $chatBody"

try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/v1/ai/proxy" -Method POST -Headers $headers -Body $chatBody -TimeoutSec 120
    Write-Host "`nStatus Code: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response Body:" -ForegroundColor Yellow
    $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $reader.DiscardBufferedData()
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error Body: $errorBody" -ForegroundColor Yellow
    }
}

# Test Models endpoint
Write-Host "`n=== Testing Models Endpoint ===" -ForegroundColor Cyan
try {
    $models = Invoke-RestMethod -Uri "$baseUrl/api/v1/ai/models?provider=zhipu" -Method GET -Headers $headers
    Write-Host "Models response:" -ForegroundColor Green
    $models | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}
