# 阅读器后端最终开发文档

本文档是后续后端开发、构建、运行、测试和部署的主依据。`reader-api-contract.md` 仍是前后端接口契约的最高优先级；若接口契约调整，必须同步更新本文档。过时的后端构建文档已删除，避免维护第二套后端说明。

维护规则：

- 后端实现、接口语义、部署方式、测试方式发生变化时，必须同步更新本文档。
- 与前端协作有关的请求、响应、鉴权和资源 URL 规则，必须同步更新 `reader-api-contract.md`。
- 本文档同时记录目标设计和当前落地状态；若二者不同，以“当前实现状态”和代码为准，再决定是否继续补齐目标能力。

## 1. 项目定位

本项目是一个支持多用户的 Web 阅读器系统。后端负责认证、用户管理、图书上传、个人书架、公共图书馆、阅读文件访问、章节解析、阅读进度、书签、分类标签、系统设置和 Docker 部署支撑。

核心目标：

- 多用户注册、登录、退出、刷新登录状态。
- 第一个注册用户自动成为管理员。
- 管理员可以禁用用户、调整角色、上下架和物理删除公共图书。
- 用户可以上传私有图书到自己的书架。
- 用户可以上传公共图书到公共图书馆。
- 其他用户将公共图书加入书架时只建立引用，不复制文件。
- 支持 EPUB、PDF、TXT 的在线阅读。
- MVP 必须生成章节目录，前端优先按章节读取内容。
- PostgreSQL 持久化，本地磁盘 volume 存储文件。
- 后端最终同时托管前端静态资源。

第一期坚持轻量实现，不引入 Redis、Elasticsearch、MQ、对象存储、全文搜索服务或重型转码服务。

当前实现状态：

- 后端开发工作区已落地在 `dev/backend/`，使用 Go + Gin + GORM。
- 后端开发文档位于 `dev/backend-docs/reader-backend-development-doc.md`；API 契约位于 `dev/reader-api-contract.md`。
- 数据库使用 PostgreSQL，测试环境使用 `dev/backend/test/docker-compose.yml` 中的 PostgreSQL 17 容器。
- 服务启动时自动执行内嵌 SQL migration。
- 后端 HTTP 端口为 `8080`，测试 PostgreSQL 暴露到宿主机 `15432`。
- 前端开发服务通常运行在 `5173`，通过 Vite proxy 将 `/api` 转发到 `8080`。
- 当前 Docker 后端镜像只构建 Go 后端；前端生产构建托管可通过 `FRONTEND_DIST_DIR` 指向磁盘目录。
- 当前 EPUB 章节内容已经支持普通 `<img src>` 和 SVG `<image href/xlink:href>` 图片归一化。
- 当前 EPUB 内嵌资源接口支持 Authorization 访问，也支持章节 HTML 内后端签名 URL 访问。

## 2. 技术栈

固定后端技术栈：

- 语言：Go 1.25。
- HTTP 框架：Gin。
- ORM：GORM。
- 数据库：PostgreSQL 17，PostgreSQL 16 兼容。
- 认证：JWT Access Token + Refresh Token。
- 密码哈希：bcrypt。
- 日志：标准库 `log/slog`。
- 配置：环境变量。
- 文件存储：本地磁盘 `/data`。
- 数据库迁移：内嵌 SQL migration 文件，服务启动时自动执行。
- 部署：Docker 多阶段构建 + docker compose。

当前依赖：

- `github.com/gin-gonic/gin`
- `gorm.io/gorm`
- `gorm.io/driver/postgres`
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- `github.com/google/uuid`

正式 schema 管理使用 SQL migration，不使用 GORM AutoMigrate。当前项目使用 `dev/backend/migrations/embed.go` 将 `*.up.sql` 内嵌到后端二进制，并由 `internal/database/migrate.go` 在启动时按文件名顺序执行。

## 3. 后端目录结构和当前实现

后端开发目录统一放在 `dev/backend/`。`book-server/` 根目录用于稳定发布版本整理，日常后端需求和 bug 修复先在 `dev/backend/` 继续开发。

```text
dev/backend/
  cmd/
    server/
      main.go
  internal/
    app/
      admin.go
      app.go
      bookmarks.go
      books.go
      dto.go
      router.go
      static.go
    config/
      config.go
    database/
      db.go
      migrate.go
    middleware/
      auth.go
      admin.go
      cors.go
      request_id.go
      recover.go
      logger.go
      body_limit.go
    auth/
      token.go
      password.go
    common/
      errors.go
      pagination.go
      response.go
      time.go
    model/
      book.go
      constants.go
      user.go
    parser/
      parser.go
      parser_test.go
    storage/
      local.go
  migrations/
    embed.go
    000001_init.up.sql
    000001_init.down.sql
    000002_backfill_bookshelf_added_at.up.sql
    000002_backfill_bookshelf_added_at.down.sql
    000003_simplify_library_status.up.sql
    000003_simplify_library_status.down.sql
    000004_bookshelf_personal_info.up.sql
    000004_bookshelf_personal_info.down.sql
  Dockerfile
  go.mod
  go.sum
```

分层规则：

- 当前 MVP 将多数 HTTP handler 和业务编排集中在 `internal/app`，优先保持行为正确和接口稳定。
- 后续若单文件继续膨胀，可按 `auth`、`reader`、`bookshelf`、`library`、`admin` 拆分 handler/service/repository，但必须保持接口契约不变。
- `model` 定义数据库模型和枚举常量。
- `dto` 定义请求和响应结构。
- `storage` 是唯一允许生成真实文件路径的模块。
- `parser` 是唯一处理图书格式解析、章节定位和资源读取的模块。
- `migrations` 使用 `embed.FS` 打入二进制，服务启动时自动迁移。

后端测试和运行辅助目录：

```text
dev/backend/test/
  docker-compose.yml       # PostgreSQL + backend 测试编排
  run-backend.sh           # 本机启动后端，连接测试 PostgreSQL
  postgres/data/           # PostgreSQL PGDATA
  backend-data/            # backend 测试运行数据卷
  backend-build/           # 本地 go build 产物
  go-build-cache/          # Go 测试/构建缓存
  backend-gocache/         # 容器/本地辅助缓存
  tmp/                     # smoke test 临时文件
  smoke.sh
  终端输出.md
```

约束：后续后端测试脚本、测试 compose、数据库文件、运行数据、构建产物、缓存和临时文件都必须放在 `dev/backend/test/` 下，避免污染 `dev/backend/` 源码目录和稳定发布目录。

## 4. 配置项

后端只从环境变量读取启动配置。系统设置可在运行后存入数据库，并优先于环境变量影响业务行为。

```text
APP_ENV=development
HTTP_ADDR=:8080
PUBLIC_BASE_URL=

DATABASE_DSN=postgres://booknest:password@localhost:5432/booknest?sslmode=disable

DATA_DIR=/data
BOOKS_DIR=/data/books
COVERS_DIR=/data/covers
TEMP_DIR=/data/temp
PUID=1000
PGID=1000

JWT_SECRET=change-me
ACCESS_TOKEN_TTL=2h
REFRESH_TOKEN_TTL=720h

MAX_UPLOAD_SIZE_MB=100
REQUEST_BODY_LIMIT_MB=110
TXT_CHUNK_SIZE=65536

ALLOW_REGISTRATION=true
LIBRARY_REVIEW_REQUIRED=false
DEFAULT_USER_STORAGE_QUOTA_MB=10240

LOG_LEVEL=info
```

规则：

- `BOOKS_DIR`、`COVERS_DIR`、`TEMP_DIR` 可为空，由 `DATA_DIR` 推导。
- Docker 部署时 `PUID`、`PGID` 控制后端主进程运行身份和 `/data` 持久化文件属主，默认 `1000:1000`。
- 生产环境禁止使用默认 `JWT_SECRET=change-me`。
- `REQUEST_BODY_LIMIT_MB` 必须大于或等于 `MAX_UPLOAD_SIZE_MB`。
- `ALLOW_REGISTRATION=false` 时，系统已有用户后拒绝新注册。
- 系统没有任何用户时，仍允许创建第一个管理员，避免初始化死锁。
- `LIBRARY_REVIEW_REQUIRED` 保留为兼容配置，当前不再影响公共图书状态；普通用户上传公共图书默认 `approved`。

## 5. HTTP 统一约定

API 基础路径固定为：

```text
/api/v1
```

健康检查：

```http
GET /api/v1/health
```

成功响应：

```json
{
  "data": {},
  "request_id": "req_xxx"
}
```

分页响应：

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100
  },
  "request_id": "req_xxx"
}
```

错误响应：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "请求参数不合法"
  },
  "request_id": "req_xxx"
}
```

通用错误码：

```text
invalid_request
unauthorized
forbidden
not_found
conflict
validation_failed
payload_too_large
unsupported_media_type
rate_limited
internal_error
```

业务错误码：

```text
registration_disabled
username_exists
email_exists
invalid_credentials
user_disabled
refresh_token_invalid
book_not_accessible
book_not_found
book_already_in_bookshelf
book_format_not_supported
book_file_missing
book_file_delete_failed
library_book_not_approved
category_not_found
tag_not_found
storage_quota_exceeded
```

分页参数：

```text
page=1
page_size=20
keyword=xxx
sort=created_at
order=desc
```

约束：

- `page` 最小为 1。
- `page_size` 默认 20，最大 100。
- `sort` 必须在接口允许列表中白名单校验。
- `order` 只允许 `asc`、`desc`。

认证头：

```http
Authorization: Bearer <access_token>
```

Refresh Token 默认使用 HttpOnly Cookie：

- Cookie 名称建议：`refresh_token`。
- 属性：`HttpOnly`、`SameSite=Lax`。
- 生产 HTTPS 下启用 `Secure`。
- `/auth/refresh` 和 `/auth/logout` 默认从 Cookie 读取 refresh token。
- 开发和 API 调试可兼容 JSON body 传递 refresh token。

## 6. 数据库设计

所有表必须通过 migration 创建。所有时间字段使用 `TIMESTAMPTZ`。

### 6.1 users

```sql
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  email VARCHAR(255),
  password_hash TEXT NOT NULL,
  nickname VARCHAR(128),
  avatar_path TEXT,
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  storage_quota_bytes BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ,
  CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user')),
  CONSTRAINT chk_users_status CHECK (status IN ('active', 'disabled'))
);

CREATE UNIQUE INDEX idx_users_email_unique_not_null ON users(email) WHERE email IS NOT NULL;
```

规则：

- `username` 必填且唯一。
- `email` 可为空，非空唯一。
- 禁用用户不能登录、刷新 token 或访问受保护接口。
- `storage_quota_bytes` 为用户专属配额（字节）：为空时套用全局默认配额 `default_user_storage_quota_mb`；值为 `0` 表示不限制。仅管理员可修改。
- `avatar_path` 存储头像文件相对路径，可为空。默认头像路径格式为 `default/<avatar_name>.svg`，自定义头像路径格式为 `avatars/<user_id>/avatar-<uuid>.<ext>`。

### 6.2 refresh_tokens

```sql
CREATE TABLE refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  user_agent TEXT,
  ip_address VARCHAR(64),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
```

规则：

- 数据库只存 refresh token 的 SHA-256 hash。
- 刷新 token 时必须轮换：撤销旧 token，写入新 token。
- 用户禁用或修改密码时撤销该用户所有 refresh token。

### 6.3 books

```sql
CREATE TABLE books (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(512) NOT NULL,
  author VARCHAR(512),
  description TEXT,
  original_filename VARCHAR(512),
  format VARCHAR(32) NOT NULL,
  file_size BIGINT NOT NULL,
  file_hash VARCHAR(128) NOT NULL,
  file_path TEXT NOT NULL,
  cover_path TEXT,
  charset VARCHAR(64),
  visibility VARCHAR(32) NOT NULL,
  owner_user_id BIGINT NOT NULL REFERENCES users(id),
  library_status VARCHAR(32),
  parse_status VARCHAR(32) NOT NULL DEFAULT 'parsed',
  parse_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT chk_books_format CHECK (format IN ('epub', 'pdf', 'txt')),
  CONSTRAINT chk_books_visibility CHECK (visibility IN ('private', 'public')),
  CONSTRAINT chk_books_library_status CHECK (
    library_status IS NULL OR library_status IN ('approved', 'hidden', 'deleted')
  ),
  CONSTRAINT chk_books_parse_status CHECK (parse_status IN ('parsed', 'partial', 'failed'))
);

CREATE INDEX idx_books_visibility ON books(visibility);
CREATE INDEX idx_books_owner_user_id ON books(owner_user_id);
CREATE INDEX idx_books_file_hash ON books(file_hash);
CREATE INDEX idx_books_library_status ON books(library_status);
CREATE INDEX idx_books_deleted_at ON books(deleted_at);
```

规则：

- 私有图书：`visibility='private'`，`library_status=NULL`。
- 公共图书：`visibility='public'`。
- 公共图书上传成功后默认 `library_status='approved'`，表示已上架，可在图书馆展示、加入书架和阅读。
- `hidden` 表示已下架，不在公共图书馆普通列表展示，不能新加入书架；已有书架引用保留但不可阅读。
- `deleted` 只作为管理员物理删除流程中的内部标记；删除完成后图书记录、相关索引和物理文件都应被清理。
- `file_path` 和 `cover_path` 保存相对路径，不保存 `/data` 绝对路径。
- 上传成功后文件必须已经落盘。

### 6.4 bookshelves

```sql
CREATE TABLE bookshelves (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  source_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  favorite BOOLEAN NOT NULL DEFAULT false,
  pinned BOOLEAN NOT NULL DEFAULT false,
  personal_title VARCHAR(512),
  personal_author VARCHAR(512),
  personal_description TEXT,
  personal_cover_path TEXT,
  personal_category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
  added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_read_at TIMESTAMPTZ,
  removed_at TIMESTAMPTZ,
  UNIQUE(user_id, book_id),
  CONSTRAINT chk_bookshelves_source_type CHECK (source_type IN ('uploaded', 'library')),
  CONSTRAINT chk_bookshelves_status CHECK (status IN ('active', 'removed', 'unavailable'))
);

CREATE INDEX idx_bookshelves_user_id ON bookshelves(user_id);
CREATE INDEX idx_bookshelves_book_id ON bookshelves(book_id);
CREATE INDEX idx_bookshelves_last_read_at ON bookshelves(last_read_at);
```

规则：

- `personal_title`、`personal_author`、`personal_description`、`personal_cover_path` 是公共图书加入个人书架后的个人覆盖信息；公共图书本体后续编辑不会覆盖已有书架项。
- 私有上传成功后自动创建书架记录，`source_type='uploaded'`。
- 上传公共图书不自动加入上传者书架。
- 公共图书加入书架只创建 `bookshelves`，不复制文件。
- 公共图书被隐藏或删除后，已有书架记录可保留，但阅读时应返回不可读。
- 书架个人分类使用 `bookshelves.personal_category_id`。
- 书架个人封面使用 `bookshelves.personal_cover_path`，保存 `COVERS_DIR` 下相对路径。
- 书架个人标签使用 `bookshelf_tags`。
- 书架个人封面、分类和标签只影响当前用户的书架展示/组织，不修改公共图书元数据。

### 6.5 categories、tags 和关联表

```sql
CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT,
  scope VARCHAR(32) NOT NULL,
  owner_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_categories_scope CHECK (scope IN ('system', 'user'))
);

CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  scope VARCHAR(32) NOT NULL,
  owner_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_tags_scope CHECK (scope IN ('system', 'user'))
);

CREATE TABLE book_categories (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, category_id)
);

CREATE TABLE book_tags (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, tag_id)
);

CREATE TABLE bookshelf_tags (
  bookshelf_id BIGINT NOT NULL REFERENCES bookshelves(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(bookshelf_id, tag_id)
);

CREATE UNIQUE INDEX idx_categories_system_slug ON categories(slug) WHERE scope = 'system';
CREATE UNIQUE INDEX idx_tags_system_slug ON tags(slug) WHERE scope = 'system';
```

MVP 只开放系统分类和系统标签。普通用户不能创建分类或标签，只能从管理员维护的集合中选择。

分类和标签职责：

- `book_categories`、`book_tags` 表示图书级元数据分类和标签，主要用于公共图书馆展示、筛选和管理员维护。
- `bookshelves.personal_category_id`、`bookshelf_tags` 表示当前用户对自己书架条目的个人整理。
- 私有上传创建书架记录时，前端提交的个人分类和标签写入 `bookshelves.personal_category_id` 和 `bookshelf_tags`。
- 公共上传提交的分类和标签写入 `book_categories` 和 `book_tags`，作为公共图书馆元数据。
- 个人书架筛选使用 `bookshelves.personal_category_id` 和 `bookshelf_tags`。
- 公共图书馆筛选使用 `book_categories` 和 `book_tags`。

### 6.6 book_chapters

```sql
CREATE TABLE book_chapters (
  id BIGSERIAL PRIMARY KEY,
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  chapter_index INTEGER NOT NULL,
  title VARCHAR(512) NOT NULL,
  locator TEXT NOT NULL,
  start_offset BIGINT,
  end_offset BIGINT,
  href TEXT,
  is_volume BOOLEAN NOT NULL DEFAULT false,
  word_count BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(book_id, chapter_index)
);

CREATE INDEX idx_book_chapters_book_id ON book_chapters(book_id);
```

定位规则：

- `chapter_index` 从 0 开始。
- TXT 使用 `start_offset`、`end_offset` 保存字节范围。
- 当前实现的 EPUB 章节定位主要使用 `href` 和 `locator`；`start_fragment_id`、`end_fragment_id`、`next_locator` 属于后续增强目标字段，当前 migration 尚未落库。
- PDF 使用 `locator='pdf_{index}'`，每 10 页一个章节。
- `is_volume=true` 表示卷标题或无正文目录节点。

### 6.7 reading_progress

```sql
CREATE TABLE reading_progress (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE SET NULL,
  format VARCHAR(32) NOT NULL,
  progress_type VARCHAR(32) NOT NULL,
  progress_value TEXT NOT NULL,
  percentage NUMERIC(6, 3),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, book_id),
  CONSTRAINT chk_reading_progress_type CHECK (progress_type IN ('epub_cfi', 'pdf_page', 'txt_offset')),
  CONSTRAINT chk_reading_progress_percentage CHECK (percentage IS NULL OR (percentage >= 0 AND percentage <= 100))
);

CREATE INDEX idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX idx_reading_progress_book_id ON reading_progress(book_id);
```

规则：

- 一个用户对一本书只有一条阅读进度。
- 保存进度时同步更新对应书架 `last_read_at`。
- `bookshelf_id` 可空，避免公共书架关系变化后历史进度丢失。

### 6.8 bookmarks

```sql
CREATE TABLE bookmarks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE SET NULL,
  title VARCHAR(255),
  position_type VARCHAR(32) NOT NULL,
  position_value TEXT NOT NULL,
  percentage NUMERIC(6, 3),
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_bookmarks_position_type CHECK (position_type IN ('epub_cfi', 'pdf_page', 'txt_offset'))
);

CREATE INDEX idx_bookmarks_user_book ON bookmarks(user_id, book_id);
```

MVP 可以先实现后端 CRUD，前端后续接入。

### 6.9 system_settings

```sql
CREATE TABLE system_settings (
  key VARCHAR(128) PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

初始化键：

```text
allow_registration=true
library_review_required=false
site_name=BookNest
max_upload_size_mb=100
default_user_storage_quota_mb=10240
```

## 7. 枚举常量

后端代码必须集中定义枚举常量，禁止散落魔法字符串。

```text
RoleAdmin=admin
RoleUser=user

UserStatusActive=active
UserStatusDisabled=disabled

BookVisibilityPrivate=private
BookVisibilityPublic=public

BookFormatEPUB=epub
BookFormatPDF=pdf
BookFormatTXT=txt

LibraryStatusApproved=approved
LibraryStatusHidden=hidden
LibraryStatusDeleted=deleted

BookshelfSourceUploaded=uploaded
BookshelfSourceLibrary=library
BookshelfStatusActive=active
BookshelfStatusRemoved=removed
BookshelfStatusUnavailable=unavailable

ProgressTypeEpubCFI=epub_cfi
ProgressTypePDFPage=pdf_page
ProgressTypeTXTOffset=txt_offset

ParseStatusParsed=parsed
ParseStatusPartial=partial
ParseStatusFailed=failed
```

## 8. 启动流程

服务启动顺序：

1. 加载配置。
2. 初始化 `slog`。
3. 校验生产环境必要配置。
4. 创建数据目录：`/data/books/private`、`/data/books/public`、`/data/covers`、`/data/temp`。
5. 连接 PostgreSQL。
6. 执行 SQL migrations。
7. 初始化系统设置默认值。
8. 初始化 Gin router 和中间件。
9. 注册 API 路由。
10. 注册前端静态资源路由。
11. 启动 HTTP server。
12. 接收到退出信号时优雅关闭。

优雅关闭要求：

- 停止接收新请求。
- 等待正在处理的请求完成，最多 10 到 30 秒。
- 关闭数据库连接池。

## 9. 中间件

必须实现：

- `RequestID`：生成或透传 `X-Request-ID`，响应头也返回。
- `Logger`：记录 method、path、status、latency、request_id、user_id、client_ip、user_agent。
- `Recover`：捕获 panic，记录堆栈，返回 `internal_error`。
- `AuthRequired`：校验 access token，查询用户存在且 active，写入上下文。
- `AdminRequired`：要求当前用户 role 为 admin。
- `BodyLimit`：上传接口限制请求体大小。

Access Token claims：

```json
{
  "sub": "user_id",
  "role": "user",
  "typ": "access",
  "exp": 1234567890,
  "iat": 1234567890
}
```

日志禁止记录密码、token 原文和上传文件内容。

## 10. 认证和用户模块

### 10.1 注册

```http
POST /api/v1/auth/register
```

请求：

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "password123",
  "nickname": "Alice"
}
```

规则：

- `username` 长度 3 到 64，只允许字母、数字、下划线、短横线。
- `email` 可空，非空时必须是合法邮箱。
- 当前实现要求 `password` 至少 6 位；若后续提高强度，必须先同步修改代码和 API 契约。
- `nickname` 可空，最大 128。
- 注册逻辑必须在事务中执行。
- 事务内使用 `SELECT pg_advisory_xact_lock(10001)` 避免并发首用户竞争。
- 加锁后统计 `users` 表，用户数为 0 时新用户为 `admin`，否则为 `user`。

### 10.2 登录、刷新、退出

```http
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

规则：

- 登录名可以是 username 或 email。
- 用户不存在或密码错误统一返回 `invalid_credentials`。
- 禁用用户返回 `user_disabled`。
- 登录成功更新 `last_login_at`。
- Refresh token 原文只返回一次。
- 刷新成功后撤销旧 refresh token 并签发新 token。
- 退出只撤销 refresh token，access token 等待自然过期。

### 10.3 当前用户和管理员用户管理

```http
GET   /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password

POST  /api/v1/users/me/avatar/upload
POST  /api/v1/users/me/avatar/default
GET   /api/v1/users/:userId/avatar
HEAD  /api/v1/users/:userId/avatar

GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
DELETE /api/v1/admin/users/:id
```

规则：

- 用户可修改 `nickname`、`email`。
- 修改密码需要旧密码。
- 修改密码后撤销该用户全部 refresh token。
- 管理员不能禁用自己。
- 第一期开发表现直接禁止管理员修改自己的角色。
- 管理员不能删除自己。
- 禁用用户时撤销该用户全部 refresh token。
- 删除用户时清除该用户账号、refresh token、书架、阅读进度、书签、拥有图书及图书文件/封面文件；若该用户拥有公共图书，同时清除其他用户引用这些图书产生的书架项、阅读进度和书签。
- `/users/me` 返回当前用户基础信息时必须包含 `storage_quota_bytes`、`storage_used_bytes` 和 `effective_storage_quota_bytes`。
- `storage_used_bytes` 按当前用户私有上传（`visibility = 'private'`）且未软删除的 `books.file_size` 汇总，不包含公共图书，也不包含引自公共图书馆的引用。
- `effective_storage_quota_bytes` 为只读计算字段，表示该用户实际生效的配额上限（字节）；`null` 表示不限制（管理员，或生效配额 ≤ 0）。详见 10.5 节。

### 10.4 用户头像

#### 数据库设计

`users.avatar_path` (VARCHAR(512), nullable)：存储头像文件相对路径。

- 默认头像：`default/default1.svg` 到 `default/default6.svg`
- 自定义头像：`avatars/<user_id>/avatar-<uuid>.<ext>`

迁移文件：

- `000006_add_user_avatar.up.sql`：添加 `avatar_path` 字段
- `000006_add_user_avatar.down.sql`：回滚迁移

#### 默认头像

系统预置 6 个简单 SVG 头像（不同颜色的圆形剪影）：

- `default1.svg` - 粉色 (#E8B4B8)
- `default2.svg` - 绿色 (#A8D5BA)
- `default3.svg` - 蓝色 (#A3C4E8)
- `default4.svg` - 橙色 (#F4D4A8)
- `default5.svg` - 紫色 (#D4A8E8)
- `default6.svg` - 米黄色 (#E8D4A8)

源码位置：`backend/assets/default/*.svg`

部署处理：

- 开发环境：`start-backend.sh` 启动时自动复制到 `$DATA_DIR/assets/default/`
- 生产环境：`Dockerfile` 将 `assets/` 打包到镜像，`docker-entrypoint.sh` 容器启动时自动复制到数据卷

#### 上传自定义头像

```http
POST /api/v1/users/me/avatar/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

规则：

- 支持格式：PNG、JPEG、WebP、GIF
- 最大 5 MB（通过 `middleware.BodyLimit(5*1024*1024)` 限制）
- 三重验证：扩展名、Content-Type、魔数（magic number）
- 保存到 `$DATA_DIR/assets/avatars/<user_id>/avatar-<uuid>.<ext>`
- 更新 `users.avatar_path` 字段
- 自动删除旧的自定义头像文件（默认头像不删除）

错误码：

- `avatar_format_not_supported`：不支持的格式
- `avatar_content_invalid`：文件内容与声明格式不匹配
- `avatar_too_large`：文件超过 5MB

实现：

- 处理函数：`uploadAvatar()`
- 存储方法：`storage.SaveUserAvatar()`
- 辅助函数：`avatarFormat()` 验证格式，`validAvatarBytes()` 验证魔数

#### 设置默认头像

```http
POST /api/v1/users/me/avatar/default
Authorization: Bearer <token>
Content-Type: application/json
```

请求体：

```json
{
  "avatar_name": "default1"  // default1 到 default6
}
```

规则：

- `avatar_name` 必须在允许列表 `[default1, default2, default3, default4, default5, default6]` 中
- 更新 `users.avatar_path` 为 `default/<avatar_name>.svg`
- 自动删除旧的自定义头像文件（默认头像不删除）

错误码：

- `invalid_avatar_name`：无效的默认头像名称

实现：

- 处理函数：`setDefaultAvatar()`

#### 获取用户头像

```http
GET /api/v1/users/:userId/avatar
HEAD /api/v1/users/:userId/avatar
```

权限：**无需认证**（公开访问）。

规则：

- 任何人都可以访问（便于前端 `<img>` 标签直接加载）
- 根据文件扩展名返回正确的 `Content-Type`（`image/svg+xml`、`image/png` 等）
- 响应头包含 `Cache-Control: public, max-age=3600`
- 未设置头像时返回 404
- 注册在公开 `api` 路由组，不在 `authRoutes` 下

实现：

- 处理函数：`userAvatar()`
- 使用 `c.File()` 返回文件流
- `http.DetectContentType()` + 扩展名映射确定 Content-Type

#### 用户模型变更

`internal/model/user.go`：

- `AvatarPath *string`：数据库字段，JSON 序列化时忽略（`json:"-"`）
- `AvatarURL *string`：计算字段，不存数据库（`gorm:"-"`）

`FindUserByID()` 填充 `avatar_url`：

```go
u.AvatarURL = avatarURL(u.ID, u.AvatarPath)
```

`avatarURL()` 辅助函数：

```go
func avatarURL(userID int64, avatarPath *string) *string {
    if avatarPath == nil || strings.TrimSpace(*avatarPath) == "" {
        return nil
    }
    v := fmt.Sprintf("/api/v1/users/%d/avatar", userID)
    return &v
}
```

#### 安全特性

1. **文件格式验证**：三重验证（扩展名、Content-Type、魔数）防止恶意文件上传
2. **大小限制**：5MB 硬限制，通过中间件和代码双重校验
3. **路径安全**：所有路径通过 `storage` 模块生成，防止路径穿越
4. **权限控制**：上传/设置头像只能修改自己的，查看头像公开访问
5. **文件隔离**：每个用户的自定义头像存储在独立目录 `avatars/<user_id>/`
6. **旧文件清理**：上传新头像或设置默认头像时自动删除旧的自定义头像

### 10.5 用户存储配额

#### 数据库字段

- `users.storage_quota_bytes` (BIGINT, nullable)：用户专属配额（字节），由管理员设置。

#### 配额规则

存储配额只针对**私人书籍**（`visibility = private`）。公共图书（`visibility = public`）既不计入用户已用空间，上传时也不校验配额。

配额生效逻辑：

1. **管理员不受配额限制**。管理员上传私人书籍时跳过所有配额校验。
2. 普通用户上传私人书籍时校验：`已用空间 + 本次文件大小 > 生效配额` 则拒绝，返回 `storage_quota_exceeded`（HTTP 403）。
3. 生效配额取值：
   - 用户设置了专属配额（`storage_quota_bytes` 非空）则使用专属值；
   - 否则套用全局默认配额 `default_user_storage_quota_mb`（默认 10240 MB = 10 GB）。
4. 生效配额 `<= 0`（包括专属配额或全局默认被设为 `0`）表示**不限制**。

#### 已用空间统计

`storage_used_bytes` 只统计当前用户 `visibility = private` 且未软删除的 `books.file_size` 汇总，不包含引自公共图书馆的引用，也不包含该用户上传的公共图书。

```go
func (s *Server) storageUsedBytes(userID int64) int64 {
    var total int64
    _ = s.db.Model(&model.Book{}).
        Where("owner_user_id = ? AND visibility = ? AND deleted_at IS NULL", userID, model.BookVisibilityPrivate).
        Select("COALESCE(SUM(file_size), 0)").Scan(&total).Error
    return total
}
```

#### 用户模型计算字段

`User` 响应包含三个存储相关字段：

- `storage_quota_bytes`（数据库字段）：专属配额，`null` 表示未设置。
- `storage_used_bytes`（计算字段）：已用空间（仅私人书籍）。
- `effective_storage_quota_bytes`（计算字段）：实际生效的配额上限。`null` 表示不限制（管理员，或生效配额 `<= 0`）；否则为具体字节上限。前端可直接用此字段展示配额和计算使用比例，无需自行判断角色或查询全局设置。

三个辅助函数：

```go
// 是否对该用户执行配额限制（管理员不受限）
func (s *Server) storageQuotaEnforced(u *model.User) bool {
    if u == nil {
        return false
    }
    return u.Role != model.UserRoleAdmin
}

// 该用户实际生效的配额（字节），<= 0 表示不限制
func (s *Server) effectiveStorageQuotaBytes(u *model.User) int64 {
    if u != nil && u.StorageQuotaBytes != nil {
        return *u.StorageQuotaBytes
    }
    return int64(s.loadSettings().DefaultUserStorageQuotaMB) * 1024 * 1024
}

// 填充 User 计算字段：已用空间、生效配额、头像 URL
func (s *Server) fillUserComputed(u *model.User) {
    if u == nil {
        return
    }
    u.StorageUsedBytes = s.storageUsedBytes(u.ID)
    u.AvatarURL = avatarURL(u.ID, u.AvatarPath)
    if !s.storageQuotaEnforced(u) {
        u.EffectiveStorageQuotaBytes = nil
        return
    }
    quota := s.effectiveStorageQuotaBytes(u)
    if quota <= 0 {
        u.EffectiveStorageQuotaBytes = nil
        return
    }
    u.EffectiveStorageQuotaBytes = &quota
}
```

#### 管理员调整配额

管理员通过 `PATCH /api/v1/admin/users/:id` 修改单个用户的 `storage_quota_bytes`：

- 传具体数值（`>= 0`）设置专属配额，`0` 表示该用户不限制。
- 传 `null` 清除专属配额，该用户回退到全局默认配额。
- 负数返回 `validation_failed`。

管理员通过 `PUT /api/v1/admin/system/settings` 修改全局默认配额 `default_user_storage_quota_mb`（`>= 0`，`0` 表示默认不限制）。

## 11. 文件存储和上传

真实目录：

```text
/data/
  books/
    private/
      user-{user_id}/
        {book_id}.{ext}
    public/
      {book_id}.{ext}
  covers/
    {book_id}/cover.jpg
    books/{book_id}/cover-{uuid}.jpg
    bookshelves/{bookshelf_id}/cover-{uuid}.jpg
  assets/
    site/
      icon.{ext}
      login-background.{ext}
    default/
      default1.svg
      default2.svg
      default3.svg
      default4.svg
      default5.svg
      default6.svg
    avatars/
      {user_id}/
        avatar-{uuid}.{ext}
  temp/
    upload-{uuid}.tmp
```

数据库保存相对路径：

```text
books/private/user-1/100.epub
books/public/200.pdf
covers/100/cover.jpg
covers/books/100/cover-uuid.jpg
covers/bookshelves/20/cover-uuid.jpg
site/icon.png
site/login-background.jpg
default/default1.svg
avatars/123/avatar-uuid.jpg
```

规则：

- 不使用原始文件名作为真实文件名。
- 原始文件名只存 `books.original_filename`。
- 所有路径生成必须由 `storage` 模块完成。
- 禁止将用户输入直接拼接为路径。
- 最终路径必须 `filepath.Clean` 并做目录前缀检查。
- 所有文件访问必须经过后端鉴权。

支持格式：

```text
epub
pdf
txt
```

上传校验：

- 扩展名白名单。
- MIME 仅作辅助，不作为唯一依据。
- PDF 文件头必须以 `%PDF-` 开头。
- EPUB 是 ZIP 包，至少校验 ZIP 头和扩展名；后续增强为校验 `mimetype` 或 `META-INF/container.xml`。
- TXT 不强依赖魔数，但应拒绝高概率二进制文件。
- 空文件拒绝。
- 文件大小不超过 `MAX_UPLOAD_SIZE_MB`。

上传流程：

1. 认证用户发起 multipart upload。
2. Body limit 拦截超大请求。
3. 创建 temp 文件。
4. 流式读取上传内容并写入 temp 文件。
5. 同时计算 SHA-256 和文件大小。
6. 校验扩展名、格式、大小；私人书籍且非管理员时校验用户配额（公共图书和管理员跳过配额校验，详见 10.5 节）。
7. 开启数据库事务。
8. 插入 `books` 记录，先写空路径或临时路径。
9. 根据 `book_id` 生成正式相对路径。
10. 移动 temp 文件到正式路径。
11. 解析元数据、封面和章节。
12. 更新 `books.file_path`、元数据、封面路径和解析状态。
13. 私有上传创建 `bookshelves` 记录；公共上传不创建书架记录。
14. 提交事务。

文件操作无法被数据库事务自动回滚，涉及文件的 service 必须写补偿逻辑：

- temp 写入失败：删除残留 temp。
- 数据库插入失败：删除 temp。
- 文件移动失败：回滚事务，删除 temp。
- 事务提交失败：删除正式文件和封面文件。

## 12. 图书解析

本项目的图书解析模块不是从零设计一套“普通文件解析器”，而是以后端 Go 实现复用和移植 Legado 本地书籍解析规则为目标。开发解析模块前必须先阅读 `legado-local-book-parsing.md`；遇到本文档没有展开的格式细节，以 `legado-local-book-parsing.md` 中整理的 Legado 行为为准。

### 12.0 Legado 解析复用原则

后端解析职责：

- 文件鉴权、文件流读取和真实路径隔离。
- 书籍元数据抽取：书名、作者、简介、语言、出版社、封面。
- 章节列表生成和章节定位信息保存。
- 章节正文读取和正文归一化。
- 图片、封面、内嵌资源读取。
- TXT 编码识别、目录正则选择、字节偏移分章。
- EPUB 目录解析、fragment 截取、跨 XHTML 拼接和图片资源解析。
- PDF 按 Legado 思路每 10 页分段，并保留后续页面图片式阅读扩展点。
- 后续 MOBI / AZW3 / AZW / UMD 等浏览器端支持较弱格式也应落在后端解析。

前端职责：

- 阅读器 UI、目录面板、翻页或滚动交互。
- 字号、主题、行距等阅读设置。
- 请求后端章节列表、章节正文和资源图片。
- 节流上报阅读进度。

必须复用的 Legado 核心模型：

- `Book` 对应本项目 `books` 表和解析元数据。
- `BookChapter` 对应本项目 `book_chapters` 表。
- `BookChapter.index` 对应 `chapter_index`。
- `BookChapter.title` 对应 `title`。
- `BookChapter.url` 对应 `locator`。
- `BookChapter.start/end` 对应 TXT 的 `start_offset/end_offset`。
- `BookChapter.startFragmentId/endFragmentId` 对应 EPUB 的 fragment 截取边界。
- `BookChapter.nextUrl` 对应 `next_locator`。
- `BookChapter.isVolume` 对应 `is_volume`。
- `BookChapter.wordCount` 对应 `word_count`。

Legado 源码参考位置：

- `legado/app/src/main/java/io/legado/app/model/localBook/LocalBook.kt`：格式分派和章节后处理。
- `legado/app/src/main/java/io/legado/app/model/localBook/BaseLocalBookParse.kt`：本地解析器统一接口。
- `legado/app/src/main/java/io/legado/app/model/localBook/TextFile.kt`：TXT 编码、目录规则、字节偏移分章、正文读取。
- `legado/app/src/main/assets/defaultData/txtTocRule.json`：TXT 默认目录规则。
- `legado/app/src/main/java/io/legado/app/model/localBook/EpubFile.kt`：EPUB 元数据、目录、fragment 截取、图片资源。
- `legado/app/src/main/java/io/legado/app/model/localBook/PdfFile.kt`：PDF 每 10 页分段、页面渲染为图片。
- `legado/app/src/main/java/io/legado/app/model/localBook/MobiFile.kt`：MOBI/AZW/AZW3 目录、正文、图片入口。
- `legado/app/src/main/java/io/legado/app/lib/mobi/`：MOBI/KF6/KF8 底层解析。
- `legado/app/src/main/java/io/legado/app/model/localBook/UmdFile.kt`：UMD 解析。
- `legado/app/src/main/java/io/legado/app/utils/HtmlFormatter.kt`：HTML 正文归一化。

MVP 必须进行基础元数据、封面和章节目录解析。实现时应优先做到“输出结构和阅读行为接近 Legado”，而不是只满足格式库能打开文件。

统一解析接口建议：

```go
type LocalBookParser interface {
    ReadMetadata(file BookFile) (*BookMetadata, error)
    ReadChapters(file BookFile) ([]Chapter, error)
    ReadContent(file BookFile, chapter Chapter) (*Content, error)
    ReadResource(file BookFile, href string) (io.ReadCloser, string, error)
}
```

统一中间模型：

- `BookMetadata`：`title`、`author`、`description`、`publisher`、`language`、`charset`、`coverBytes`。
- `Chapter`：`index`、`title`、`locator`、`startOffset`、`endOffset`、`href`、`startFragmentId`、`endFragmentId`、`nextLocator`、`isVolume`、`wordCount`。
- `Content`：`contentType` 和 `content`。正文使用安全 HTML 子集，允许纯文本换行和 `<img src="...">`。

上传元数据优先级：

1. 用户显式传入的 `title`、`author`、`description`。
2. 后端解析出的元数据。
3. 原始文件名去掉扩展名。

解析失败策略：

- 文件格式不可读或不符合声明格式时，上传失败。
- 元数据或封面解析失败不阻断上传。
- 章节解析失败时必须创建一个全书兜底章节。
- 解析部分失败时 `books.parse_status='partial'`，完整失败但有兜底章节时为 `failed`。

### 12.1 TXT 解析，按 Legado `TextFile.kt` 移植

TXT 必须由后端解析，不允许前端一次性加载整本书。TXT 解析的目标是复现 Legado `TextFile.kt` 的目录识别、字节偏移分章和正文读取行为。

MVP 要求：

- 编码识别至少支持 UTF-8、UTF-8 BOM、GBK、GB18030。
- 首次自动识别读取前 `512000` 字节。
- 检测 UTF-8 BOM，正文偏移从 3 开始。
- 移植 Legado `txtTocRule.json` 默认目录规则中的启用规则。
- 自动选择最佳目录规则。
- 章节保存字节级 `start_offset` 和 `end_offset`。
- 章节内容读取必须按字节范围读取。
- 无目录命中时按固定大小分段兜底。

默认启用目录规则至少包含：

- `目录(去空白)`
- `目录`
- `数字 分隔符 标题名称`
- `大写数字 分隔符 标题名称`
- `正文 标题/序号`
- `Chapter/Section/Part/Episode 序号 标题`
- `特殊符号 序号 标题`
- `特殊符号 标题(单个)`
- `章/卷 序号 标题`
- `书名 括号 序号`
- `书名 序号`
- `字数分割 分节阅读`

自动选择目录规则：

1. 读取前 `512000` 字节为 `blockContent`。
2. 遍历启用规则，使用 multiline 正则匹配。
3. 对每条规则统计有效匹配数 `csNum` 和疑似误匹配数 `numE`。
4. 只有 `csNum >= numE * 3` 且比当前最佳规则多超过 2 个匹配时，替换最佳规则。
5. 最佳匹配数超过 70 时提前停止。
6. 未选中规则时走无目录分段。

按规则分章：

- 分块读取文件，每块 `512000` 字节。
- 块刚好读满时，从末尾向前找换行字节 `0x0a`，避免标题行被截断。
- 正则匹配位置是字符串下标，但章节 `start/end` 必须换算为文件字节偏移。
- 首个标题前有内容时创建 `前言` 章节。
- 两个标题之间正文为空时，将上一个章节标记为 `is_volume=true`。
- 单章过长时后续可拆分，MVP 可先不做自动长章拆分，但不得一次性读整本书。

无规则分章：

- 默认按约 10KB 到 64KB 字节段拆分，具体值可配置。
- 标题为 `第 N 段`。
- 最后一段过小时合并到上一段。

正文读取：

- 只读 `start_offset` 到 `end_offset` 字节范围。
- 按书籍 `charset` 解码。
- 开头连续空白和换行可归一化为全角缩进。
- 不允许只保存字符串下标；所有章节定位必须可回到原始文件字节偏移。

### 12.2 EPUB 解析，按 Legado `EpubFile.kt` 移植

EPUB 解析目标是复现 Legado `EpubFile.kt` 的 metadata、TOC/spine、fragment 截取、跨 XHTML 拼接和图片资源读取行为。

MVP 要求：

- 解析 OPF metadata 获取标题、作者、简介、语言、出版社。
- 封面优先从 EPUB cover image 提取，保存为 JPEG。
- 目录优先使用 nav/ncx TOC。
- TOC 为空或失败时退回 spine。
- 支持章节 `href` 和 fragment。
- 有 children 的 TOC 节点标记为 `is_volume=true`。
- 上一章记录下一章 locator，用于正文截断。

正文读取规则：

- 章节内容返回清洗后的 HTML 子集。
- 删除 `script`、`style`、`title`、隐藏元素。
- 将 EPUB 的 SVG `image xlink:href` 归一成 `<img src="...">`。
- 图片相对路径用当前资源 href 解析成规范 href。
- 必须支持同一 XHTML 内多章的 fragment 截取。
- 必须支持一章跨多个 XHTML 的拼接，直到下一章资源或 fragment。

EPUB 文件接口仍保留，前端可用 `epub.js` 直接读取完整文件；但章节目录接口和章节内容接口也必须可用。

### 12.3 PDF 解析，按 Legado `PdfFile.kt` 分段

MVP PDF 以文件流 + 前端 `pdf.js` 渲染为主，后端必须支持 Range。章节生成按 Legado `PdfFile.kt` 的 `PAGE_SIZE = 10` 思路实现。

章节规则：

- 每 10 页作为一个章节。
- `locator='pdf_{index}'`。
- 标题为 `分段_{index + 1}`。
- 章节内容可返回页面占位 HTML，例如 `<img src="0">` 到 `<img src="9">`，或返回页面范围结构。

元数据和封面：

- 尽量读取 PDF metadata。
- 封面可渲染第一页为 JPEG；如果实现成本较高，封面失败不阻断上传。

后续如果采用 Legado 图片式阅读，则增加页面图片资源接口。

### 12.4 后续格式，按 Legado 解析路线扩展

MVP 只开放 EPUB、PDF、TXT 上传，但解析模块目录和接口必须为后续 MOBI / AZW3 / AZW / UMD 预留。

MOBI / AZW3 / AZW：

- 参考 `MobiFile.kt` 和 `legado/app/src/main/java/io/legado/app/lib/mobi/`。
- 不得只按扩展名判断内部结构，应读取 MOBI header，区分 KF6 和 KF8。
- AZW3 基本按 KF8 路径处理。
- 元数据、封面、TOC、正文和图片资源都应通过后端解析后输出统一 `Chapter` 和 HTML 子集正文。

UMD：

- 参考 `UmdFile.kt`。
- 元数据来自 UMD header。
- 目录来自 chapters titles。
- 正文按 chapter index 读取。

压缩包导入：

- 后续支持 ZIP/RAR/7Z 时，参考 Legado 的压缩包导入行为。
- 解压后只导入符合 `txt|epub|umd|pdf|mobi|azw3|azw` 的书籍文件。
- 临时文件丢失时可根据原压缩包和 `originName` 重新导入。

HTML 正文归一化：

- 参考 `HtmlFormatter.kt`。
- `&nbsp;`、`&ensp;`、`&emsp;` 转空格。
- 删除零宽字符和不可打印字符。
- 块级标签转为换行。
- 删除普通 HTML 标签，但 `formatKeepImg` 场景保留并规范化 `<img src="...">`。
- 最终 EPUB/MOBI/AZW3 正文形态应为普通文本、换行缩进和 `<img src="...">`。

## 13. 个人书架模块

### 13.1 私有图书上传

```http
POST /api/v1/bookshelf/upload
```

multipart 字段：

```text
file=<book file>
cover=<optional image file>
title=<optional>
author=<optional>
description=<optional>
category_ids=<optional comma separated>
tag_ids=<optional comma separated>
```

创建结果：

- `books.visibility='private'`
- `books.owner_user_id=current_user_id`
- `bookshelves.user_id=current_user_id`
- `bookshelves.source_type='uploaded'`
- 文件路径：`books/private/user-{user_id}/{book_id}.{ext}`
- `cover` 可选，支持 `png`、`jpg/jpeg`、`webp`，最大 5MB。
- 传入 `cover` 时优先使用上传封面；未传入时 EPUB 会尝试自动提取封面。

### 13.2 我的书架

```http
GET    /api/v1/bookshelf
GET    /api/v1/bookshelf/:id
GET    /api/v1/bookshelf/:id/cover
GET    /api/v1/bookshelf/:id/download
PATCH  /api/v1/bookshelf/:id
PUT    /api/v1/bookshelf/:id/cover
DELETE /api/v1/bookshelf/:id
```

列表支持：

```text
keyword
format
category_id
tag_id
source_type
favorite
page
page_size
sort=last_read_at|added_at|title
order=asc|desc
```

默认排序：

```text
pinned desc, last_read_at desc nulls last, added_at desc
```

返回中必须包含：

- `book_id`
- `title`
- `author`
- `format`
- `cover_url`
- `source_type`
- `visibility`
- `library_status`
- `favorite`
- `pinned`
- `last_read_at`
- `added_at`
- `readable`
- `unreadable_reason`

`unreadable_reason` 规则：

- `readable=true` 时为空。
- 公共图书被隐藏时为 `library_hidden`。
- 公共图书处于删除流程或被标记删除时为 `library_deleted`。
- 图书文件丢失时为 `file_missing`。
- 其他权限变化导致不可读时为 `permission_denied`。

编辑书架只能改个性化字段：

- `personal_title`
- `personal_author`
- `personal_description`
- `personal_category_id`
- `favorite`
- `pinned`
- 个人书架标签集合

书架封面规则：

- `BookshelfItem.cover_url` 返回当前书架项有效封面地址。
- `GET /api/v1/bookshelf/:id/cover` 只能读取当前用户自己的 active 书架项封面。
- 私人上传图书的书架封面来自 `books.cover_path`。
- 公共图书引用优先使用 `bookshelves.personal_cover_path`，为空时回退 `books.cover_path`。
- `PUT /api/v1/bookshelf/:id/cover` 使用 multipart 字段 `cover`，支持 `png`、`jpg/jpeg`、`webp`，最大 5MB。
- 私人上传图书更新封面时写入 `books.cover_path`。
- 公共图书引用更新封面时写入 `bookshelves.personal_cover_path`，只影响当前用户当前书架项，不修改公共图书本体。
- 新封面写入成功后删除旧封面物理文件。

删除规则：

- 私有上传图书：先写入 `books.deleted_at` 删除标记，再物理删除真实文件和封面，最后删除该书的书架记录、书架标签、阅读进度、书签、章节和 `books` 图书记录。
- 公共图书：只删除当前用户书架记录，不删除 `books` 和真实文件。
- 私有图书删除成功后，数据库不再保留该图书条目和相关个人数据，效果等同从未添加过该图书。
- 私有文件物理删除失败时，撤销 `books.deleted_at` 删除标记，并返回 `book_file_delete_failed`，前端应提示用户稍后重试。

下载规则：

- `GET /api/v1/bookshelf/:id/download` 只能下载当前用户自己的 active 书架项。
- 私有上传图书只要求当前用户拥有该书架项。
- 公共图书馆引用要求该书仍为 `library_status='approved'`、`deleted_at IS NULL`。
- 下载内容是原始图书文件，不重新转码或修改文件内容。
- 响应头使用 `Content-Disposition: attachment`。
- 下载文件名使用书架上展示的标题；若存在 `personal_title` 则优先使用，否则使用图书标题。
- 扩展名使用图书 `format`。
- 文件名中的路径分隔符、控制字符和常见非法文件名字符会被替换或移除。

### 13.3 从公共图书加入书架

```http
POST /api/v1/bookshelf/from-library/:bookId
POST /api/v1/library/books/:id/add-to-bookshelf
```

两个接口路由到同一个 service。

规则：

- 目标书籍必须存在。
- `visibility='public'`。
- `library_status='approved'`。
- 未软删除。
- 当前用户没有同一本书的书架记录。
- 只创建书架引用，不复制文件。

## 14. 公共图书馆模块

### 14.1 普通用户接口

```http
GET  /api/v1/library/books
GET  /api/v1/library/books/:id
GET  /api/v1/library/books/:id/download
POST /api/v1/library/books/upload
PUT  /api/v1/library/books/:id/cover
POST /api/v1/library/books/:id/hide
POST /api/v1/library/books/:id/show
```

普通用户默认只能看到：

- `visibility='public'`
- `library_status='approved'`
- `deleted_at IS NULL`

例外：

- `GET /api/v1/library/books/:id` 允许上传者查看自己上传的 `approved` 或 `hidden` 公共图书。
- `GET /api/v1/library/books` 默认只返回 `approved` 公共图书。
- 上传者可通过 `mine=true&status=approved|hidden` 查询自己上传的公共图书。
- 其他普通用户不可查看非本人上传的 `hidden` 公共图书。
- `deleted` 不对普通用户查询开放。
- 管理员物理删除后的公共图书已经删除 `books` 记录，不会出现在上传者的 `mine=true` 列表中，效果等同从未上传过该书。

列表支持：

```text
keyword
format
category_id
tag_id
mine=true|false
status=approved|hidden
page
page_size
sort=created_at|title
order=asc|desc
```

返回中必须包含：

- 当前用户是否已加入书架：`in_bookshelf`。
- 对应书架 id：`bookshelf_id`，未加入为空。
- 公共图书当前分类和标签 id：`category_ids`、`tag_ids`。

上传公共图书：

- `books.visibility='public'`
- `books.owner_user_id=current_user_id`
- `books.library_status='approved'`
- 文件路径：`books/public/{book_id}.{ext}`。
- 不自动给上传者创建书架记录。
- `category_ids`、`tag_ids` 写入 `book_categories`、`book_tags`，作为公共图书馆元数据。
- `cover` 可选，支持 `png`、`jpg/jpeg`、`webp`，最大 5MB。
- 传入 `cover` 时优先使用上传封面；未传入时 EPUB 会尝试自动提取封面。

编辑自己上传的公共图书：

- 接口：`PATCH /api/v1/library/books/:id`。
- 只能由上传者调用，且图书必须未被删除。
- 可编辑 `title`、`author`、`description`、`category_ids`、`tag_ids`。
- 公共封面通过 `PUT /api/v1/library/books/:id/cover` 单独更新，multipart 字段为 `cover`。
- 更新公共封面写入 `books.cover_path`，影响公共图书馆展示和之后新加入书架时的默认封面。
- 公共封面更新不影响已经设置了 `bookshelves.personal_cover_path` 的私人书架项。
- `title` 非空；`author`、`description` 可为空或 `null`。
- `category_ids`、`tag_ids` 传入数组时整体替换公共图书的分类和标签，传入 `null` 或空数组表示清空。
- `category_ids`、`tag_ids` 必须全部来自系统分类/标签，不存在时返回 `category_not_found` 或 `tag_not_found`。
- 分类和标签修改影响公共图书馆筛选和展示，不影响用户已加入私人书架后的个人分类/标签。

下载公共图书馆图书：

- 只允许下载 `visibility='public'`、`library_status='approved'`、`deleted_at IS NULL` 的公共图书。
- `hidden`、`deleted` 或已被管理员物理删除的公共图书不可下载。
- 下载内容是原始图书文件，不重新转码或修改文件内容。
- 响应头使用 `Content-Disposition: attachment`。
- 下载文件名使用公共图书馆展示标题，扩展名使用图书 `format`。
- 文件名中的路径分隔符、控制字符和常见非法文件名字符会被替换或移除。

下架自己上传的公共图书：

- 只能由上传者调用。
- 只把 `library_status` 更新为 `hidden`，不写 `deleted_at`。
- 不删除图书文件、封面、书架引用、阅读进度或书签。
- 已加入他人书架的引用保留，仍可在书架展示，但 `readable=false` 且 `unreadable_reason='library_hidden'`。
- 已下架图书不能被新增加入书架，也不能继续阅读。

重新上架自己上传的公共图书：

- 只能由上传者调用。
- 图书必须是 `visibility='public'`、`owner_user_id=current_user_id`、`deleted_at IS NULL`。
- 只允许 `library_status='hidden'` 的图书重新上架。
- 只把 `library_status` 更新为 `approved`，并更新 `updated_at`。
- 不删除或修改图书文件、封面、书架引用、阅读进度或书签。
- `approved`、`deleted` 或其他状态不允许通过该接口恢复。
- 管理员物理删除后的公共图书不存在数据库记录，不能被重新上架。

### 14.2 管理员接口

```http
GET    /api/v1/admin/library/books
PATCH  /api/v1/admin/library/books/:id/status
DELETE /api/v1/admin/library/books/:id
```

状态流转：

```text
approved -> hidden
hidden -> approved
```

规则：

- `PATCH /status` 只负责上架/下架，不接受 `deleted`。
- `hidden` 表示下架，普通用户公共列表不可见，已有书架记录不可读。
- `deleted` 只由 `DELETE` 删除流程内部写入，作为物理删除中的标记。
- `DELETE` 是真正删除：先标记 `library_status='deleted'` 和 `deleted_at`，再删除图书文件和封面，最后清理数据库中的图书记录、章节、分类/标签关联、书架引用、阅读进度和书签。
- 删除完成后，数据库不再保留该公共图书索引记录，物理文件也应被删除。

## 15. 阅读模块

### 15.1 阅读权限

用户可以读取图书的条件：

- 管理员可以读取所有未物理丢失的图书。
- 私有图书：当前用户存在该书 `bookshelves` 记录。
- 公共图书：当前用户已登录，且 `library_status='approved'`。

公共图书 `hidden` 和 `deleted` 均不可读。

### 15.2 阅读接口

```http
GET /api/v1/reader/books/:bookId/meta
GET /api/v1/reader/books/:bookId/file
GET /api/v1/reader/books/:bookId/cover
GET /api/v1/reader/books/:bookId/text
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/resources?href=...
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
```

`GET /api/v1/reader/books/:bookId/cover` 读取图书本体 `books.cover_path`，不读取书架个人封面。书架展示封面必须使用 `BookshelfItem.cover_url` 返回的 `/bookshelf/:id/cover`。

文件读取：

- 先做阅读权限校验。
- 使用 `http.ServeContent` 或等价实现支持 Range。
- PDF 必须支持 Range，并正确返回 206。
- 设置 `Accept-Ranges: bytes`。
- 不暴露真实文件路径。

章节目录响应：

```json
{
  "data": [
    {
      "id": 1,
      "chapter_index": 0,
      "title": "第一章",
      "is_volume": false,
      "word_count": 1234
    }
  ]
}
```

章节内容响应：

```json
{
  "data": {
    "id": 1,
    "chapter_index": 0,
    "title": "第一章",
    "content_type": "text",
    "content": "..."
  }
}
```

规则：

- TXT 返回 `content_type='text'`，正文保留换行；章节正文开头多余空行会被归一，原文首段已有缩进时保留原缩进，首段无缩进时后端可按阅读排版补全角缩进 `　　`，由前端按纯文本排版。
- EPUB、PDF 返回 `content_type='html'`。

TXT 分块接口：

```http
GET /api/v1/reader/books/:bookId/text?offset=0&limit=65536
```

规则：

- 只允许 TXT。
- `offset` 是字节偏移。
- `limit` 默认 `TXT_CHUNK_SIZE`，最大 1MB。
- 该接口作为调试、兜底和超大 TXT 局部读取能力；前端优先使用章节接口。

EPUB 内嵌资源接口：

```http
GET /api/v1/reader/books/:bookId/resources?href=images/cover.jpg
```

章节 HTML 内资源 URL 示例：

```http
GET /api/v1/reader/books/:bookId/resources?expires=1780333831&href=images/cover.jpg&sig=...&uid=2
```

规则：

- `href` 必填，表示 EPUB 内部资源相对路径或已规范化路径。
- 后端必须对 `href` 做规范化和路径穿越检查，禁止读取 EPUB 外部文件。
- 响应应返回资源原始字节和正确 `Content-Type`。
- 图片类型至少支持 JPEG、PNG、GIF、WEBP、SVG。
- 常规 API 调用携带 `Authorization: Bearer <access_token>`，服务端按当前用户做阅读权限校验。
- 章节 HTML 内的 `<img>` 请求无法携带 Authorization，因此后端在章节内容中生成带 `uid`、`expires`、`sig` 的短期签名 URL。
- 签名必须绑定用户 ID、图书 ID、规范化资源 href 和过期时间。
- 签名访问时仍必须查询用户状态，并调用阅读权限校验；签名只替代 Authorization 头，不替代权限判断。
- 签名过期、缺失或不匹配时返回 401。
- 可返回 `Cache-Control: private, max-age=3600`。
- 章节 HTML 中的 `<img src>` 应改写为该接口 URL，并对 `href` 做 URL 编码。
- EPUB 的 `<svg><image href="...">` 和 `<svg><image xlink:href="...">` 必须归一为普通 `<img src="...">` 后再返回给前端。

阅读进度保存：

```json
{
  "progress_type": "epub_cfi",
  "progress_value": "epubcfi(/6/2!... )",
  "percentage": 12.345
}
```

规则：

- EPUB 只允许 `epub_cfi`。
- PDF 只允许 `pdf_page`。
- TXT 只允许 `txt_offset`。
- `percentage` 可空，非空范围 0 到 100。
- 保存使用 `INSERT ... ON CONFLICT (user_id, book_id) DO UPDATE`。
- 保存进度时更新当前用户对应书架 `last_read_at`。

## 16. 书签模块

```http
GET    /api/v1/books/:bookId/bookmarks
POST   /api/v1/books/:bookId/bookmarks
PATCH  /api/v1/bookmarks/:id
DELETE /api/v1/bookmarks/:id
```

规则：

- 用户必须有阅读权限才能创建和查看该书书签。
- 用户只能修改、删除自己的书签。
- 书签位置类型与阅读进度一致。
- 书签第一版作为增强能力实现；若前端第一版未启用书签按钮，后端接口仍可按阶段八交付。

## 17. 分类和标签模块

分类接口：

```http
GET    /api/v1/categories
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

标签接口：

```http
GET    /api/v1/tags
POST   /api/v1/admin/tags
PATCH  /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

规则：

- 普通用户可查询系统分类和标签。
- 普通用户不能创建分类或标签。
- 管理员删除分类或标签前检查关联关系。
- MVP 若有关联则拒绝删除，避免悬挂筛选和用户打标失效。

## 18. 系统模块

```http
GET /api/v1/system/info
GET /api/v1/system/icon
GET /api/v1/admin/system/storage
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
POST /api/v1/admin/system/icon
DELETE /api/v1/admin/system/icon
```

`/system/info` 匿名可访问，返回：

- `site_name`
- `site_icon_url`
- `allow_registration`
- `library_review_required`
- `supported_formats`
- `max_upload_size_mb`
- `default_user_storage_quota_mb`

管理员可修改：

- `allow_registration`
- `library_review_required`
- `site_name`
- `max_upload_size_mb`
- `default_user_storage_quota_mb`

管理员可通过 `POST /admin/system/icon` 上传站点图标，通过 `DELETE /admin/system/icon` 删除自定义图标。

管理员可通过 `POST /admin/system/login-background` 上传登录页背景图片，通过 `DELETE /admin/system/login-background` 删除登录页背景。

规则：

- 设置修改立即生效。
- 站点图标保存到 `ASSETS_DIR`，默认位于 `DATA_DIR/assets`。`/system/icon` 匿名读取当前图标文件，未设置时返回 `not_found`。
- 站点图标支持 `png`、`jpg/jpeg`、`webp`、`svg`、`ico`，最大 2MB；上传成功后替换旧文件并写入 `system_settings.site_icon_path`。
- 登录页背景保存到 `ASSETS_DIR/site/login-background{ext}`。`/system/login-background` 匿名读取当前背景图片文件，未设置时返回 `not_found`。
- 登录页背景支持 `png`、`jpg/jpeg`、`webp`、`gif`，最大 10MB；上传成功后替换旧文件并写入 `system_settings.login_background_path`。
- 登录页背景为匿名访问接口，便于登录页在未认证状态下直接展示背景。
- `max_upload_size_mb` 不能超过启动时 `REQUEST_BODY_LIMIT_MB`。
- `library_review_required` 保留用于兼容旧配置，当前不影响公共图书上传状态。
- 存储统计第一期可基于数据库 `file_size` 汇总，目录真实占用后续增强。

## 19. 路由总表

```text
/api/v1/health
/api/v1/system/info
/api/v1/system/icon
/api/v1/system/login-background

/api/v1/auth/register
/api/v1/auth/login
/api/v1/auth/refresh
/api/v1/auth/logout
/api/v1/auth/me

/api/v1/users/me
/api/v1/users/me/password
/api/v1/users/me/avatar/upload
/api/v1/users/me/avatar/default
/api/v1/users/:userId/avatar

/api/v1/bookshelf
/api/v1/bookshelf/:id
/api/v1/bookshelf/upload
/api/v1/bookshelf/:id/cover
/api/v1/bookshelf/:id/download
/api/v1/bookshelf/from-library/:bookId

/api/v1/library/books
/api/v1/library/books/:id
/api/v1/library/books/:id/download
/api/v1/library/books/upload
/api/v1/library/books/:id/cover
/api/v1/library/books/:id/hide
/api/v1/library/books/:id/show
/api/v1/library/books/:id/add-to-bookshelf

/api/v1/categories
/api/v1/tags

/api/v1/reader/books/:bookId/meta
/api/v1/reader/books/:bookId/file
/api/v1/reader/books/:bookId/cover
/api/v1/reader/books/:bookId/text
/api/v1/reader/books/:bookId/chapters
/api/v1/reader/books/:bookId/chapters/:chapterId/content
/api/v1/reader/books/:bookId/resources?href=...
/api/v1/reader/books/:bookId/progress

/api/v1/books/:bookId/bookmarks
/api/v1/bookmarks/:id

GET /api/v1/admin/users
GET /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
DELETE /api/v1/admin/users/:id
/api/v1/admin/library/books
/api/v1/admin/library/books/:id/status
/api/v1/admin/library/books/:id
/api/v1/admin/categories
/api/v1/admin/categories/:id
/api/v1/admin/tags
/api/v1/admin/tags/:id
/api/v1/admin/system/storage
/api/v1/admin/system/settings
/api/v1/admin/system/icon
/api/v1/admin/system/login-background
```

前端静态资源：

- API 路由优先注册。
- 非 API 路径返回前端 `index.html`，支持 Vue Router history 模式。
- `/api/*` 不允许回退到前端页面，必须返回 JSON 404。

## 20. 权限矩阵

```text
能力                           匿名  普通用户  管理员
注册                           是    是        是
登录                           是    是        是
获取系统信息                   是    是        是
获取站点图标                   是    是        是
获取登录页背景                 是    是        是
查看自己的信息                 否    是        是
修改自己的信息                 否    是        是
上传自定义头像                 否    是        是
设置默认头像                   否    是        是
获取任意用户头像               否    是        是
上传私有图书                   否    是        是
查看自己的书架                 否    是        是
阅读自己的私有图书             否    是        是
下载自己的书架图书             否    是        是
浏览 approved 公共图书         否    是        是
下载 approved 公共图书         否    是        是
上传公共图书                   否    是        是
下架自己上传公共图书           否    是        是
重新上架自己上传公共图书       否    是        是
加入公共图书到书架             否    是        是
管理自己的阅读进度             否    是        是
管理自己的书签                 否    是        是
管理用户                       否    否        是
调整用户存储配额               否    否        是
上下架公共图书                 否    否        是
物理删除公共图书               否    否        是
管理系统分类                   否    否        是
管理系统标签                   否    否        是
管理系统设置                   否    否        是
上传/删除站点图标              否    否        是
上传/删除登录页背景            否    否        是
```

## 21. 事务边界

必须使用事务：

- 注册用户，特别是第一个管理员判断。
- Refresh token 轮换。
- 私有图书上传：创建 book、移动文件、解析、创建 bookshelf。
- 公共图书上传：创建 book、移动文件、解析。
- 公共图书加入书架。
- 删除私有图书。
- 管理员删除公共图书并处理相关书架引用。
- 保存阅读进度并更新 `last_read_at`。

禁止在事务中执行慢速上传流读取。上传到 temp 应发生在事务之前。

## 22. 安全要求

密码：

- 使用 bcrypt。
- 不记录明文密码。
- 不返回 password hash。
- 修改密码后撤销 refresh token。

Token：

- Access token 默认 2 小时。
- Refresh token 默认 30 天。
- Refresh token 只存 hash。
- 生产环境 `JWT_SECRET` 必须强随机。

上传：

- 限制扩展名和大小。
- 不信任 MIME。
- 禁止路径穿越。
- 上传先写 temp。
- 校验后移动到正式目录。

文件访问：

- 不暴露 `/data`。
- 所有文件访问走 API。
- Range 请求也必须先鉴权。
- 封面接口同样鉴权，避免枚举私有图书。

日志：

- 不记录 token。
- 不记录密码。
- 不记录上传文件内容。
- 错误日志记录 request_id。

## 23. 性能和资源控制

必须遵守：

- 上传使用流式写入，不读入内存。
- 文件下载使用流式响应。
- PDF 文件接口支持 Range。
- TXT 按章节或分块读取，不一次性加载全书。
- 列表查询必须分页。
- 数据库连接池限制最大连接数。
- 不做后端重型转码。
- temp 目录定期清理。

数据库连接池默认：

```text
max_open_conns=20
max_idle_conns=5
conn_max_lifetime=1h
```

## 24. 测试计划

单元测试必须覆盖：

- 密码 hash 和校验。
- JWT 创建和解析。
- Refresh token hash。
- 上传格式判断。
- 路径生成和路径穿越防护。
- 分页参数解析。
- TXT 编码识别。
- TXT 目录规则选择。
- TXT 字节偏移分章。
- EPUB `<img src>` 相对路径改写。
- EPUB SVG `<image href>` / `<image xlink:href>` 归一成普通 `<img src>`。
- EPUB 资源 href 规范化和路径穿越防护。

集成测试必须覆盖：

- 第一个注册用户成为管理员。
- 并发注册不会产生多个管理员。
- 后续注册用户为普通用户。
- 禁用用户不能登录和刷新。
- Refresh token 可以轮换，退出后失效。
- 私有图书上传创建 `books` 和 `bookshelves`。
- 私有图书只有拥有者能读取。
- 公共图书上传创建单份文件。
- 公共图书加入书架不复制文件。
- 重复加入公共图书返回 conflict。
- 从书架移除公共图书不删除文件。
- 管理员隐藏公共图书后普通用户不可读。
- PDF 文件接口支持 Range。
- TXT 章节内容接口只读取对应字节范围。
- EPUB 章节内容中的普通图片和 SVG 封面图片都能被改写为可访问资源 URL。
- EPUB 内嵌资源接口在携带 Authorization 时可访问。
- EPUB 内嵌资源接口在使用后端生成的 `uid/expires/sig` 签名 URL 时可被浏览器图片请求访问。
- 过期或篡改的 EPUB 资源签名返回 401。
- 阅读进度 upsert 正常。

手工验收：

1. Docker compose 启动成功。
2. 访问前端首页。
3. 注册第一个用户并确认角色为管理员。
4. 注册第二个用户并确认角色为普通用户。
5. 上传私有 EPUB，确认生成章节。
6. 打开 EPUB 封面章节，确认 SVG `<image xlink:href>` 封面能显示为正常图片。
7. 上传 TXT，确认自动分章并能按章节读取。
8. 上传公共 PDF。
9. 第二个用户将公共 PDF 加入书架。
10. 磁盘上公共 PDF 只有一份。
11. 保存阅读进度。
12. 管理员隐藏公共图书，普通用户无法继续读取。

## 25. 开发阶段

### 阶段一：后端骨架

交付：

- Go module。
- 配置加载。
- logger。
- PostgreSQL 连接。
- migration runner。
- Gin router。
- `/api/v1/health`。
- 前端静态资源托管占位。

验收：

- 本地启动成功。
- 数据库迁移成功。
- health 返回正常。

### 阶段二：用户和认证

交付：

- users、refresh_tokens migration。
- 注册、登录、refresh、logout、me。
- auth middleware。
- admin middleware。
- 管理员用户列表和禁用。

验收：

- 首个用户为管理员。
- 并发注册不产生多个管理员。
- 禁用用户无法登录和刷新。

### 阶段三：文件存储和私有书架

交付：

- books、bookshelves migration。
- storage 模块。
- 上传私有图书。
- 我的书架列表、详情、编辑、删除。
- 文件读取权限。

验收：

- 上传文件落盘。
- 数据库记录正确。
- 非拥有者无法读取私有图书。

### 阶段四：解析和阅读接口

交付：

- book_chapters migration。
- reader meta。
- reader file，支持 Range。
- reader cover。
- TXT 分块。
- EPUB/PDF/TXT 元数据、封面和章节解析。
- 章节目录和章节内容接口。
- reading_progress migration。
- 进度保存和读取。

验收：

- PDF Range 请求返回 206。
- EPUB 上传后生成章节目录。
- PDF 上传后按页段生成章节目录。
- TXT 上传后生成章节目录，章节内容接口只读取对应字节范围。
- 保存进度后书架 `last_read_at` 更新。

### 阶段五：公共图书馆

交付：

- 公共图书上传。
- 公共图书列表和详情。
- 加入书架。
- 管理员公共图书列表。
- 上架、下架、物理删除。

验收：

- 公共图书加入书架不复制文件。
- 普通用户可以下架自己上传的公共图书。
- 普通用户可以重新上架自己上传且已下架的公共图书。
- 管理员删除公共图书后，数据库索引记录和物理文件都被清理。
- hidden/deleted 图书普通用户不可读。

### 阶段六：分类标签

交付：

- categories、tags、关联表 migration。
- 分类查询和管理员管理。
- 标签查询和管理员管理。
- 书架和图书馆筛选。

验收：

- 系统分类和标签唯一约束生效。
- 普通用户不能创建分类或标签。
- 用户只能使用已有系统分类和系统标签。

### 阶段七：管理和系统设置

交付：

- system_settings。
- storage 统计。
- 系统设置读取和修改。
- 管理员用户状态和角色调整。

验收：

- 关闭注册后非首个用户无法注册。
- 系统设置修改后立即生效。

### 阶段八：增强

交付：

- bookmarks。
- temp 清理任务。
- 更完整日志。
- 备份恢复文档。
- 管理员维护 TXT 目录规则。
- 用户对单本书重新解析目录。

## 26. 构建、运行和 Docker 部署

部署形态：

```text
booknest-test-backend    Go 后端测试容器，宿主机端口 8080
booknest-test-postgres   PostgreSQL 17 测试容器，宿主机端口 15432
```

应用容器挂载：

```text
dev/backend/test/backend-data -> /data
dev/backend/test/postgres/data -> PostgreSQL PGDATA
```

当前 `dev/backend/Dockerfile`：

1. `golang:1.25-alpine` 构建阶段执行 `go build -o /out/booknest-server ./cmd/server`。
2. `alpine:3.22` 运行阶段包含后端二进制、`docker-entrypoint.sh` 和 `su-exec`。
3. 容器默认以 root 启动 entrypoint，仅用于创建 `/data`、`/data/books`、`/data/covers`、`/data/temp` 并执行 `chown -R ${PUID}:${PGID} /data`。
4. 后端主进程通过 `su-exec ${PUID}:${PGID}` 降权运行，不长期使用 root。
5. 宿主机 bind mount 的 `backend-data` 中新生成的图书、封面和临时文件属主应显示为 `PUID:PGID`，不应显示为 root。
6. `docker exec` 未指定 `--user` 时仍以 root 进入容器，便于维护。

权限规则：

- compose 部署必须支持 `PUID`、`PGID` 环境变量，默认 `1000:1000`。
- `PUID`、`PGID` 必须是非 root 数字 ID；配置为 `0` 会直接拒绝启动。
- 不使用 `chmod 777` 作为长期权限方案。
- 不允许后端主进程长期 root 运行。
- Linux bind mount 部署时，如果宿主机用户 UID/GID 不是 `1000:1000`，应在 `.env` 或 compose 环境变量中设置为宿主机实际用户的 `id -u` 和 `id -g`。
- 同一份 compose 可在 Linux/amd64 和 Linux/arm64 上本机构建部署；不要在 compose 中写死 `platform: linux/amd64` 或 `GOARCH=amd64`。`golang:1.25-alpine`、`alpine:3.22`、`postgres:17-alpine` 都使用 Docker 官方多架构镜像，在哪个架构上构建就生成对应架构的后端二进制。

当前 Docker 测试栈：

```bash
cd /Users/cyilin/dev/book-server/dev/backend/test
docker compose up -d --build
```

只重建后端：

```bash
cd /Users/cyilin/dev/book-server/dev/backend/test
docker compose up -d --build backend
```

本地只运行后端，连接测试 PostgreSQL：

```bash
cd /Users/cyilin/dev/book-server/dev/backend/test
docker compose up -d postgres
bash run-backend.sh
```

后端测试和构建：

```bash
cd /Users/cyilin/dev/book-server/dev/backend
mkdir -p test/go-build-cache test/backend-build
GOCACHE=/Users/cyilin/dev/book-server/dev/backend/test/go-build-cache go test ./...
GOCACHE=/Users/cyilin/dev/book-server/dev/backend/test/go-build-cache go build -o /Users/cyilin/dev/book-server/dev/backend/test/backend-build/booknest-server ./cmd/server
```

Smoke test：

```bash
cd /Users/cyilin/dev/book-server/dev/backend/test
bash smoke.sh
```

前端静态资源托管：

- 当前后端通过 `FRONTEND_DIST_DIR` 从磁盘目录托管前端 `dist`。
- 若 `FRONTEND_DIST_DIR/index.html` 不存在，后端返回内置占位 HTML，表示后端正在运行。
- API 路由优先；非 `/api/*` 路径回退到前端 `index.html`，支持 Vue Router history 模式。
- 生产环境可在独立流程中构建前端，并将 `FRONTEND_DIST_DIR` 指向该构建产物目录；也可后续扩展 Dockerfile 增加 Node 构建阶段。

## 27. 已确认实现决策

后续开发必须遵守：

1. 私有图书从书架删除时，先标记删除，再物理删除真实文件，最后清理数据库图书条目和关联数据。
2. 公共图书状态只有 `approved`、`hidden`、`deleted` 三种。
3. `approved` 表示已上架；`hidden` 表示已下架；`deleted` 只用于管理员物理删除流程中的内部标记。
4. 用户上传公共图书后默认 `approved`，不自动加入自己的书架。
5. 普通用户只能下架自己上传的公共图书，下架不删除文件、不清理已有书架引用。
6. 管理员删除公共图书必须物理删除文件，并清理图书记录、章节、分类/标签关联、书架引用、阅读进度和书签。
7. 公共图书加入个人书架只创建引用，不复制文件。
8. Refresh Token 浏览器端默认使用 HttpOnly Cookie，开发调试可兼容 JSON body。
9. MVP 需要解析 EPUB/PDF/TXT 元数据、封面和章节目录。
10. TXT 必须解析章节目录，前端先请求章节目录，再按章节请求正文。
11. 分类和标签由管理员统一维护，普通用户只能从已有分类、标签中选择。
12. 需要用户存储配额：默认 10 GB，只限制私人书籍，公共图书不计入；管理员不受配额限制，且可调整每个用户的配额。
13. 不允许匿名浏览公共图书馆；公共图书馆仅登录用户可访问。
14. 所有真实文件路径必须由 `storage` 模块生成和校验。
15. 阅读文件和封面访问必须携带 Authorization；EPUB 内嵌资源访问必须携带 Authorization 或后端签名 URL，且两种方式都必须通过阅读权限校验。
