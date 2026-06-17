# BookNest 开发指南

## 快速启动

### 1. 启动数据库
```bash
./start-postgres.sh
```

这会启动一个 PostgreSQL 17 容器，监听在 `localhost:15432`。

### 2. 启动后端
```bash
./start-backend.sh
```

后端会运行在 `http://localhost:8080`，自动连接到测试数据库。

### 3. 启动前端
```bash
./start-frontend.sh
```

前端会运行在 `http://localhost:5173`，自动代理 `/api` 到后端。

### 停止服务
```bash
./stop-services.sh
```

### 清理测试数据
```bash
./clean-test.sh
```

这会停止所有服务并删除 `test/` 目录中的所有测试数据。

## 测试数据目录结构

所有测试相关的运行时数据都存储在 `test/` 目录中（已在 .gitignore 中忽略）：

```
test/
├── backend-data/          # 后端数据目录
│   ├── books/            # 上传的图书文件
│   ├── covers/           # 图书封面
│   ├── assets/           # 系统资源
│   └── temp/             # 临时文件
└── postgres-data/         # PostgreSQL 数据文件
```

## 数据库连接信息

- **Host**: localhost
- **Port**: 15432
- **Database**: booknest
- **User**: booknest
- **Password**: booknest_password
- **连接字符串**: `postgres://booknest:booknest_password@localhost:15432/booknest?sslmode=disable`

## 前端开发

### 安装依赖
```bash
cd frontend
npm install
```

### 开发命令
```bash
npm run dev      # 启动开发服务器
npm run build    # 构建生产版本
npm run preview  # 预览生产构建
```

## 后端开发

### 构建
```bash
cd backend
go build -o booknest-server ./cmd/server
```

### 运行测试
```bash
cd backend
go test ./...
```

## Docker 部署测试

启动完整的 Docker 环境（PostgreSQL + 后端）：

```bash
docker compose -f docker-compose.test.yml up -d
```

## 环境变量

后端支持以下环境变量（`start-backend.sh` 已配置默认值）：

- `DATABASE_DSN` - 数据库连接字符串
- `DATA_DIR` - 数据根目录
- `BOOKS_DIR` - 图书文件目录
- `COVERS_DIR` - 封面图片目录
- `ASSETS_DIR` - 系统资源目录
- `TEMP_DIR` - 临时文件目录
- `FRONTEND_DIST_DIR` - 前端构建产物目录
- `HTTP_ADDR` - HTTP 监听地址（默认 :8080）
- `JWT_SECRET` - JWT 密钥（测试环境）

## 目录结构

```
.
├── backend/                # Go 后端
│   ├── cmd/server/        # 主程序入口
│   ├── internal/          # 内部包
│   └── migrations/        # 数据库迁移
├── frontend/              # Vue 3 前端
│   ├── src/              # 源代码
│   └── dist/             # 构建产物（git ignored）
├── frontend-docs/         # 前端开发文档
├── backend-docs/          # 后端开发文档
├── test/                  # 测试数据目录（git ignored）
├── start-postgres.sh      # 启动数据库
├── start-backend.sh       # 启动后端
├── start-frontend.sh      # 启动前端
├── stop-services.sh       # 停止服务
├── clean-test.sh          # 清理测试数据
└── docker-compose.test.yml # Docker 测试配置
```

## 开发工作流

### 日常开发
1. 启动数据库：`./start-postgres.sh`（只需启动一次）
2. 启动后端：`./start-backend.sh`（新终端）
3. 启动前端：`./start-frontend.sh`（新终端）
4. 访问 http://localhost:5173 开始开发

### 清理重启
```bash
./clean-test.sh           # 清理所有测试数据
./start-postgres.sh       # 重新启动数据库（会重新初始化）
./start-backend.sh        # 启动后端（会运行迁移）
./start-frontend.sh       # 启动前端
```

## 常见问题

### 数据库连接失败
确保 PostgreSQL 容器已启动：
```bash
docker ps | grep booknest-test-postgres
```

如果没有运行，执行 `./start-postgres.sh`。

### 端口占用
- 前端默认端口：5173
- 后端默认端口：8080
- PostgreSQL 端口：15432

如果端口被占用，可以修改对应的启动脚本。

### 数据迁移问题
后端启动时会自动运行数据库迁移。如果迁移失败，检查数据库连接或清理测试数据重新开始。

## 生产部署

生产环境部署请参考：
- 后端文档：`backend-docs/`
- 前端文档：`frontend-docs/`
- API 契约：`reader-api-contract.md`
