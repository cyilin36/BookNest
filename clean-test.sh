#!/usr/bin/env bash
# 清理所有测试数据

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

echo "🧹 清理测试数据..."

# 停止服务
./stop-services.sh

# 删除测试数据
if [ -d "test" ]; then
    echo "🗑️  删除 test/ 目录..."
    rm -rf test/
fi

echo "✅ 测试数据已清理"
