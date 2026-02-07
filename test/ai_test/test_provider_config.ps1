$baseUrl = "http://localhost:8080"

# Login
$loginBody = @{
    username = "admin"
    password = "admin"
} | ConvertTo-Json

Write-Host "=== Step 1: Login ===" -ForegroundColor Cyan
try {
    $login = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $login.data.token
    Write-Host "Login success! Token obtained" -ForegroundColor Green
} catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# Get provider config
Write-Host "`n=== Step 2: Get Provider Config ===" -ForegroundColor Cyan
try {
    $config = Invoke-RestMethod -Uri "$baseUrl/api/v1/ai/providers?provider=zhipu" -Method GET -Headers $headers
    Write-Host "Config retrieved:" -ForegroundColor Green
    $config.data | ConvertTo-Json
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test provider connection
Write-Host "`n=== Step 3: Test Provider Connection ===" -ForegroundColor Cyan
$testBody = @{
    provider = "zhipu"
} | ConvertTo-Json

try {
    $test = Invoke-RestMethod -Uri "$baseUrl/api/v1/ai/providers/test" -Method POST -Headers $headers -Body $testBody -ContentType "application/json" -TimeoutSec 30
    Write-Host "Test result:" -ForegroundColor Green
    $test | ConvertTo-Json
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $reader.DiscardBufferedData()
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error: $errorBody" -ForegroundColor Yellow
    }
}
