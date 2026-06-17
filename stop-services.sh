#!/usr/bin/env bash
# 停止所有测试服务

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

echo "🛑 停止所有测试服务..."
docker compose -f docker-compose.test.yml down

echo "✅ 所有服务已停止"
