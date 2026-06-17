#!/usr/bin/env bash
# 启动 PostgreSQL 测试数据库

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 启动 PostgreSQL 测试数据库..."
docker compose -f docker-compose.test.yml up -d postgres

echo "⏳ 等待数据库就绪..."
timeout 60 bash -c 'until docker compose -f docker-compose.test.yml exec -T postgres pg_isready -U booknest -d booknest > /dev/null 2>&1; do sleep 1; done'

echo "✅ PostgreSQL 已启动并就绪"
echo "📊 连接信息:"
echo "   Host: localhost"
echo "   Port: 15432"
echo "   Database: booknest"
echo "   User: booknest"
echo "   Password: booknest_password"
echo ""
echo "🔗 连接字符串: postgres://booknest:booknest_password@localhost:15432/booknest?sslmode=disable"
