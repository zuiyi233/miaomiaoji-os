$apiUrl = "http://192.168.32.15:39999"
$apiKey = "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B"

$headers = @{
    "Authorization" = "Bearer $apiKey"
    "Content-Type" = "application/json"
}

# Test 1: List models
Write-Host "=== Testing /v1/models ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$apiUrl/v1/models" -Method GET -Headers $headers -TimeoutSec 30
    Write-Host "Success! Models:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
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

# Test 2: Chat completions
Write-Host "`n=== Testing /v1/chat/completions ===" -ForegroundColor Cyan
$body = @{
    model = "glm-4.7"
    messages = @(
        @{ role = "user"; content = "Hello, how are you?" }
    )
    temperature = 0.7
    max_tokens = 100
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/v1/chat/completions" -Method POST -Headers $headers -Body $body -TimeoutSec 60
    Write-Host "Success! Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
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

# Test 3: Try without /v1 prefix
Write-Host "`n=== Testing without /v1 prefix ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$apiUrl/chat/completions" -Method POST -Headers $headers -Body $body -TimeoutSec 60
    Write-Host "Success! Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
}
