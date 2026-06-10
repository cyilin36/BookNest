# BookNest

BookNest 是一个多用户在线图书管理与阅读系统，提供个人书架、公共图书馆、在线阅读、阅读进度、书签和后台管理等功能。

## 特性

- 支持 EPUB、PDF、TXT 图书上传与在线阅读
- 多用户账号体系，第一个注册用户自动成为管理员
- 个人书架与公共图书馆
- 阅读进度保存和书签管理
- 图书分类、标签和基础信息编辑
- 管理后台：用户、图书、分类、标签、系统设置
- 支持 Docker Compose 一键部署

## 技术栈

- 前端：Vue 3、Vite、Pinia、Naive UI
- 后端：Go、Gin、GORM
- 数据库：PostgreSQL
- 部署：Docker、Docker Compose

## 快速开始

### Docker Compose 部署

克隆项目后，在项目根目录执行：

```bash
docker compose up -d
```

访问：

```text
http://localhost:8080
```

默认会使用 GitHub Container Registry 上的镜像：

```text
ghcr.io/cyilin36/booknest:latest
```

更新镜像：

```bash
docker compose pull
docker compose up -d
```

### 自建镜像

如果希望在本机从源码构建镜像：

```bash
docker compose up --build -d
```

## 数据目录

默认数据会保存在项目根目录：

```text
app-data/        上传的图书、封面和临时文件
postgres-data/   PostgreSQL 数据
```

## 配置

生产部署前建议修改 `docker-compose.yml` 中的以下配置：

```yaml
POSTGRES_PASSWORD: booknest_password
DATABASE_DSN: postgres://booknest:booknest_password@postgres:5432/booknest?sslmode=disable
JWT_SECRET: change-this-secret-before-production
```

其中 `JWT_SECRET` 必须替换为强随机密钥。

常用环境变量：

| 变量 | 说明 |
| --- | --- |
| `HTTP_ADDR` | 服务监听地址，默认 `:8080` |
| `DATABASE_DSN` | PostgreSQL 连接字符串 |
| `JWT_SECRET` | JWT 密钥 |
| `DATA_DIR` | 应用数据目录，默认 `/data` |
| `FRONTEND_DIST_DIR` | 前端静态资源目录 |
| `MAX_UPLOAD_SIZE_MB` | 最大上传大小 |
| `ALLOW_REGISTRATION` | 是否允许注册 |
| `LIBRARY_REVIEW_REQUIRED` | 公共图书是否需要审核 |

## 本地开发

后端：

```bash
cd backend
go test ./...
go run ./cmd/server
```

前端：

```bash
cd frontend
npm ci
npm run dev
```

前端开发服务器默认运行在 `5173`，并会将 `/api` 代理到后端 `8080`。

## 项目结构

```text
backend/              Go 后端
frontend/             Vue 前端
Dockerfile            生产镜像构建文件
docker-compose.yml    Docker Compose 部署文件
```

## 说明

- 第一个注册成功的用户会自动成为管理员。
- 公网部署建议放在 HTTPS 反向代理后。
- 生产环境请妥善备份 `app-data/` 和 `postgres-data/`。
