# PicSlim 启动脚本
# 使用: .\scripts\dev.ps1

$ErrorActionPreference = "Stop"

Write-Host "Starting PicSlim development server..." -ForegroundColor Cyan

# 检查前端依赖
if (-not (Test-Path "frontend/node_modules")) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Yellow
    Push-Location frontend
    pnpm install
    Pop-Location
}

# 启动 Wails 开发模式
Write-Host "Launching Wails dev server..." -ForegroundColor Green
wails dev
