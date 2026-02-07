# 检查配置文件和数据库中的配置

Write-Host "=== 配置文件内容 ===" -ForegroundColor Cyan
$configPath = "configs/providers/zhipu.yaml"
if (Test-Path $configPath) {
    Get-Content $configPath
} else {
    Write-Host "配置文件不存在!" -ForegroundColor Red
}

Write-Host "`n=== 环境变量 ===" -ForegroundColor Cyan
Get-ChildItem Env: | Where-Object { $_.Name -like '*ZHIPU*' -or $_.Name -like '*GOOGLE*' -or $_.Name -like '*API*' } | Select-Object Name, Value | Format-Table -AutoSize
