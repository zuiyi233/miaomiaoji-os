$apiUrl = "http://192.168.32.15:39999/v1"
$apiKey = "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B"

# Test with Authorization header
Write-Host "=== Testing with Authorization: Bearer ===" -ForegroundColor Cyan
$headers = @{
    "Authorization" = "Bearer $apiKey"
    "Content-Type" = "application/json"
}

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/models" -Method GET -Headers $headers -TimeoutSec 30
    Write-Host "Success! Found $($response.data.Count) models" -ForegroundColor Green
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $reader.DiscardBufferedData()
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error Body: $errorBody" -ForegroundColor Yellow
    }
}

# Test chat completions
Write-Host "`n=== Testing Chat Completions ===" -ForegroundColor Cyan
$body = @{
    model = "glm-4.7"
    messages = @(
        @{ role = "user"; content = "Hello" }
    )
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/chat/completions" -Method POST -Headers $headers -Body $body -TimeoutSec 60
    Write-Host "Success! Response: $($response.choices[0].message.content.Substring(0, [Math]::Min(50, $response.choices[0].message.content.Length)))..." -ForegroundColor Green
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
}
