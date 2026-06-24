# PicSlim 构建脚本
# 使用: .\scripts\build.ps1

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot

Write-Host "Building PicSlim..." -ForegroundColor Cyan

# 清理缓存
Write-Host "Cleaning caches..." -ForegroundColor Yellow

# 前端 Vite 缓存
$ViteCache = Join-Path $ProjectRoot "frontend\node_modules\.vite"
if (Test-Path $ViteCache) {
    Write-Host "  Removing Vite cache: $ViteCache" -ForegroundColor Gray
    Remove-Item -Recurse -Force $ViteCache
}

# 前端 dist
$FrontendDist = Join-Path $ProjectRoot "frontend\dist"
if (Test-Path $FrontendDist) {
    Write-Host "  Removing frontend dist: $FrontendDist" -ForegroundColor Gray
    Remove-Item -Recurse -Force $FrontendDist
}

# 构建输出目录
$BuildOutput = Join-Path $ProjectRoot "build\bindows"
if (Test-Path $BuildOutput) {
    Write-Host "  Removing build output: $BuildOutput" -ForegroundColor Gray
    Remove-Item -Recurse -Force $BuildOutput
}

# Wails 缓存
$WailsCache = Join-Path $env:LOCALAPPDATA "\Wails\builds"
if (Test-Path $WailsCache) {
    Write-Host "  Removing Wails cache: $WailsCache" -ForegroundColor Gray
    Remove-Item -Recurse -Force $WailsCache
}

# 构建前端
Write-Host "Building frontend..." -ForegroundColor Green
Push-Location (Join-Path $ProjectRoot "frontend")
pnpm run build
Pop-Location

# 构建 Wails 应用
Write-Host "Building Wails application..." -ForegroundColor Green
Push-Location $ProjectRoot
wails build
Pop-Location

Write-Host "Build complete!" -ForegroundColor Cyan
Write-Host "Output: build\windows\" -ForegroundColor Gray
