#!/usr/bin/env bash
# 启动前端开发服务器

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

cd "$FRONTEND_DIR"

echo "🚀 启动前端开发服务器..."
echo "🌐 服务地址: http://localhost:5173"
echo "🔗 API 代理: /api -> http://localhost:8080"
echo ""

exec npm run dev
