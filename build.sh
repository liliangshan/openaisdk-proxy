#!/bin/bash

# 构建脚本 - 编译前端并嵌入到后端可执行文件中

set -e

echo "=========================================="
echo "开始构建..."
echo "=========================================="

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 创建 bin 目录
mkdir -p bin

# ==========================================
# 1. 构建前端 (frontend)
# ==========================================
echo ""
echo ">>> 构建前端 (frontend)..."

cd frontend

# 安装依赖（如果需要）
npm ci 2>/dev/null || npm install

# 构建前端
npm run build

# 清理旧的构建
rm -rf "$SCRIPT_DIR/cmd/api/user"
mkdir -p "$SCRIPT_DIR/cmd/api/user"

# 将构建结果复制到 cmd/api/user 目录
cp -r dist/* "$SCRIPT_DIR/cmd/api/user/"

cd "$SCRIPT_DIR"

# ==========================================
# 2. 构建后端 (cmd/api)
# ==========================================
echo ""
echo ">>> 构建后端 (cmd/api)..."

cd cmd/api

# 安装依赖（如果需要）
go mod download 2>/dev/null || true

# 清理旧的构建
rm -f "$SCRIPT_DIR/bin/openaisdk-proxy" "$SCRIPT_DIR/bin/openaisdk-proxy.exe"

# 编译后端
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

if [ "$GOOS" = "windows" ]; then
    go build -o "$SCRIPT_DIR/bin/openaisdk-proxy.exe" .
    echo ">>> 后端已编译: bin/openaisdk-proxy.exe"
else
    go build -o "$SCRIPT_DIR/bin/openaisdk-proxy" .
    echo ">>> 后端已编译: bin/openaisdk-proxy"
fi

cd "$SCRIPT_DIR"



echo ""
echo "=========================================="
echo "构建完成!"
echo "=========================================="
echo "后端: openaisdk-proxy"
echo ""
echo "注意: 前端已嵌入到后端可执行文件中"
echo ""
