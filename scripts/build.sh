#!/bin/bash
# PicSlim 构建脚本

set -e

echo "Building PicSlim..."

# 构建前端
echo "Building frontend..."
cd frontend
pnpm run build
cd ..

# 构建 Wails 应用
echo "Building Wails application..."
wails build

echo "Build complete!"
echo "Output: build/windows/"
