$baseUrl = "http://localhost:8080"

# Test register
$registerBody = @{
    username = "testuser_$(Get-Random)"
    password = "Test123456"
    nickname = "测试用户"
} | ConvertTo-Json

Write-Host "Testing register with body: $registerBody"

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -TimeoutSec 10
    Write-Host "Register Response:"
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Register Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $reader.DiscardBufferedData()
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error Body: $errorBody"
    }
}
