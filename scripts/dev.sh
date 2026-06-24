#!/bin/bash
# PicSlim 启动脚本

set -e

echo "Starting PicSlim development server..."

# 检查前端依赖
if [ ! -d "frontend/node_modules" ]; then
    echo "Installing frontend dependencies..."
    cd frontend
    pnpm install
    cd ..
fi

# 启动 Wails 开发模式
echo "Launching Wails dev server..."
wails dev
