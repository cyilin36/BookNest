#!/usr/bin/env bash
# 启动后端开发服务器

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
DATA_DIR="$PROJECT_ROOT/test/backend-data"

# 创建必要的数据目录
mkdir -p "$DATA_DIR/books" "$DATA_DIR/covers" "$DATA_DIR/assets" "$DATA_DIR/temp"

# 复制默认头像到运行目录
mkdir -p "$DATA_DIR/assets/default"
cp -f "$BACKEND_DIR/assets/default"/*.svg "$DATA_DIR/assets/default/" 2>/dev/null || true

cd "$BACKEND_DIR"

# 设置环境变量
export DATABASE_DSN="${DATABASE_DSN:-postgres://booknest:booknest_password@localhost:15432/booknest?sslmode=disable}"
export DATA_DIR="$DATA_DIR"
export BOOKS_DIR="$DATA_DIR/books"
export COVERS_DIR="$DATA_DIR/covers"
export ASSETS_DIR="$DATA_DIR/assets"
export TEMP_DIR="$DATA_DIR/temp"
export FRONTEND_DIST_DIR="${FRONTEND_DIST_DIR:-$PROJECT_ROOT/frontend/dist}"
export HTTP_ADDR="${HTTP_ADDR:-:8080}"
export JWT_SECRET="${JWT_SECRET:-test-only-change-me}"

echo "🚀 启动后端开发服务器..."
echo "📁 数据目录: $DATA_DIR"
echo "🌐 服务地址: http://localhost:8080"
echo ""

exec go run ./cmd/server
