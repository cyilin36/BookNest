# 阅读器后端开发文档

本文档基于 `reader-project-plan.md`、`reader-backend-build.md` 和 `legado-local-book-parsing.md` 汇总整理，是后续后端开发的主依据。若其他文档与本文档冲突，开发时优先按本文档执行；确需调整时，应先更新本文档再编码。

## 1. 项目定位

本项目是一个支持多用户的 Web 阅读器系统。后端负责认证、用户管理、图书上传、个人书架、公共图书馆、阅读文件访问、章节解析、阅读进度、书签、分类标签、系统设置和 Docker 部署支撑。

核心目标：

- 多用户注册、登录、退出、刷新登录状态。
- 第一个注册用户自动成为管理员。
- 管理员可以禁用用户、调整角色、审核和管理公共图书。
- 用户可以上传私有图书到自己的书架。
- 用户可以上传公共图书到公共图书馆。
- 其他用户将公共图书加入书架时只建立引用，不复制文件。
- 支持 EPUB、PDF、TXT 的在线阅读。
- MVP 必须生成章节目录，前端优先按章节读取内容。
- PostgreSQL 持久化，本地磁盘 volume 存储文件。
- 后端最终同时托管前端静态资源。

第一期坚持轻量实现，不引入 Redis、Elasticsearch、MQ、对象存储、全文搜索服务或重型转码服务。

## 2. 技术栈

固定后端技术栈：

- 语言：Go 1.23 或更新稳定版本。
- HTTP 框架：Gin。
- ORM：GORM。
- 数据库：PostgreSQL 16 或 17。
- 认证：JWT Access Token + Refresh Token。
- 密码哈希：bcrypt。
- 日志：标准库 `log/slog`。
- 配置：环境变量。
- 文件存储：本地磁盘 `/data`。
- 数据库迁移：SQL migration 文件。
- 部署：Docker 多阶段构建 + docker-compose。

推荐依赖：

- `github.com/gin-gonic/gin`
- `gorm.io/gorm`
- `gorm.io/driver/postgres`
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- `github.com/google/uuid`
- `github.com/golang-migrate/migrate/v4`
- `github.com/joho/godotenv`，仅开发环境可选。

正式 schema 管理使用 SQL migration，不使用 GORM AutoMigrate 作为生产迁移方案。

## 3. 后端目录结构

后端目录统一放在 `backend/`。

```text
backend/
  cmd/
    server/
      main.go
    migrate/
      main.go
  internal/
    app/
      app.go
      router.go
      static.go
    config/
      config.go
    database/
      db.go
      migrate.go
      tx.go
    middleware/
      auth.go
      admin.go
      cors.go
      request_id.go
      recover.go
      logger.go
      body_limit.go
    auth/
      handler.go
      service.go
      repository.go
      token.go
      password.go
      model.go
      dto.go
    user/
      handler.go
      service.go
      repository.go
      model.go
      dto.go
    book/
      model.go
      repository.go
      service.go
      dto.go
    bookshelf/
      handler.go
      service.go
      repository.go
      dto.go
    library/
      handler.go
      service.go
      repository.go
      dto.go
    category/
      handler.go
      service.go
      repository.go
      model.go
      dto.go
    tag/
      handler.go
      service.go
      repository.go
      model.go
      dto.go
    reader/
      handler.go
      service.go
      progress.go
      resource.go
      dto.go
    parser/
      parser.go
      model.go
      htmlfmt/
      txt/
      epub/
      pdf/
    upload/
      service.go
      validator.go
      hash.go
    storage/
      storage.go
      local.go
      path.go
    admin/
      handler.go
      service.go
    system/
      handler.go
      service.go
      settings.go
    common/
      errors.go
      response.go
      pagination.go
      validate.go
      time.go
  migrations/
  tests/
    integration/
  go.mod
  go.sum
```

分层规则：

- `handler` 只处理 HTTP 参数、认证上下文和响应。
- `service` 处理业务规则、权限组合、事务和跨模块调用。
- `repository` 只处理数据库查询和持久化。
- `model` 定义数据库模型和枚举常量。
- `dto` 定义请求和响应结构。
- `storage` 是唯一允许生成真实文件路径的模块。
- `parser` 是唯一处理图书格式解析、章节定位和资源读取的模块。

## 4. 配置项

后端只从环境变量读取启动配置。系统设置可在运行后存入数据库，并优先于环境变量影响业务行为。

```text
APP_ENV=development
HTTP_ADDR=:8080
PUBLIC_BASE_URL=

DATABASE_DSN=postgres://book_reader:password@localhost:5432/book_reader?sslmode=disable

DATA_DIR=/data
BOOKS_DIR=/data/books
COVERS_DIR=/data/covers
TEMP_DIR=/data/temp

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
- 生产环境禁止使用默认 `JWT_SECRET=change-me`。
- `REQUEST_BODY_LIMIT_MB` 必须大于或等于 `MAX_UPLOAD_SIZE_MB`。
- `ALLOW_REGISTRATION=false` 时，系统已有用户后拒绝新注册。
- 系统没有任何用户时，仍允许创建第一个管理员，避免初始化死锁。
- `LIBRARY_REVIEW_REQUIRED=true` 时，普通用户上传公共图书默认 `pending`。

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
library_book_not_approved
library_review_required
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
- `storage_quota_bytes` 为空时使用系统默认配额。

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
  publisher VARCHAR(255),
  language VARCHAR(64),
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
    library_status IS NULL OR library_status IN ('pending', 'approved', 'rejected', 'hidden', 'deleted')
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
- 公共图书无需审核时：`library_status='approved'`。
- 公共图书需要审核时：`library_status='pending'`。
- `pending` 表示待管理员审核，不表示上传中。
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
  personal_category_id BIGINT,
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

- `personal_category_id` 应在 `categories` 表创建后补充外键约束，或在 migration 中先创建 `categories` 再创建 `bookshelves`。
- 私有上传成功后自动创建书架记录，`source_type='uploaded'`。
- 上传公共图书不自动加入上传者书架。
- 公共图书加入书架只创建 `bookshelves`，不复制文件。
- 公共图书被隐藏或删除后，已有书架记录可保留，但阅读时应返回不可读。
- 书架个人分类使用 `bookshelves.personal_category_id`。
- 书架个人标签使用 `bookshelf_tags`。
- 书架个人分类和标签只影响当前用户的书架组织，不修改公共图书元数据。

### 6.5 categories、tags 和关联表

```sql
CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT,
  parent_id BIGINT REFERENCES categories(id),
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
CREATE UNIQUE INDEX idx_categories_user_slug ON categories(owner_user_id, slug) WHERE scope = 'user';
CREATE UNIQUE INDEX idx_tags_system_slug ON tags(slug) WHERE scope = 'system';
CREATE UNIQUE INDEX idx_tags_user_slug ON tags(owner_user_id, slug) WHERE scope = 'user';
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
  start_fragment_id TEXT,
  end_fragment_id TEXT,
  next_locator TEXT,
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
- EPUB 使用 `href`、`start_fragment_id`、`end_fragment_id`、`next_locator`。
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
site_name=Book Reader
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

LibraryStatusPending=pending
LibraryStatusApproved=approved
LibraryStatusRejected=rejected
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
- `password` 至少 8 位。
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

GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
```

规则：

- 用户可修改 `nickname`、`email`。
- 修改密码需要旧密码。
- 修改密码后撤销该用户全部 refresh token。
- 管理员不能禁用自己。
- 第一期开发表现直接禁止管理员修改自己的角色。
- 禁用用户时撤销该用户全部 refresh token。
- `/users/me` 返回当前用户基础信息时必须包含 `storage_quota_bytes` 和 `storage_used_bytes`。
- `storage_used_bytes` 第一版按当前用户私有上传且未软删除的 `books.file_size` 汇总，不包含引自公共图书馆的引用。

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
    {book_id}.jpg
  temp/
    upload-{uuid}.tmp
```

数据库保存相对路径：

```text
books/private/user-1/100.epub
books/public/200.pdf
covers/100.jpg
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
6. 校验扩展名、格式、大小、用户配额。
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

### 13.2 我的书架

```http
GET    /api/v1/bookshelf
GET    /api/v1/bookshelf/:id
PATCH  /api/v1/bookshelf/:id
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
- 公共图书被删除或软删除时为 `library_deleted`。
- 公共图书被拒绝时为 `library_rejected`。
- 图书文件丢失时为 `file_missing`。
- 其他权限变化导致不可读时为 `permission_denied`。

编辑书架只能改个性化字段：

- `personal_title`
- `personal_category_id`
- `favorite`
- `pinned`
- 个人书架标签集合

删除规则：

- 私有上传图书：删除书架记录，软删除 `books`，物理删除真实文件和封面。
- 公共图书：只删除当前用户书架记录，不删除 `books` 和真实文件。
- 私有文件物理删除失败时，不提交数据库删除状态。

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
POST /api/v1/library/books/upload
```

普通用户只能看到：

- `visibility='public'`
- `library_status='approved'`
- `deleted_at IS NULL`

例外：

- `GET /api/v1/library/books/:id` 允许上传者查看自己上传的 `pending` 公共图书。
- `GET /api/v1/library/books` 默认只返回 `approved` 公共图书。
- 若需要让上传者在列表中查看自己的待审核公共图书，使用 `mine=true&status=pending` 查询；该查询只返回当前用户自己上传的 pending 图书。
- 其他普通用户不可查看非本人上传的 pending、rejected、hidden、deleted 公共图书。

列表支持：

```text
keyword
format
category_id
tag_id
page
page_size
sort=created_at|title
order=asc|desc
```

返回中必须包含：

- 当前用户是否已加入书架：`in_bookshelf`。
- 对应书架 id：`bookshelf_id`，未加入为空。

上传公共图书：

- `books.visibility='public'`
- `books.owner_user_id=current_user_id`
- 审核开启时 `library_status='pending'`。
- 审核关闭时 `library_status='approved'`。
- 文件路径：`books/public/{book_id}.{ext}`。
- 不自动给上传者创建书架记录。

`pending` 特殊规则：

- 上传者可以查看和读取自己上传的 pending 公共图书，用于确认内容。
- 上传者可通过详情接口查看自己的 pending 图书，也可通过 `mine=true&status=pending` 在列表中查看。
- 其他普通用户不可见不可读。
- 管理员可见并可审核。

### 14.2 管理员接口

```http
GET    /api/v1/admin/library/books
PATCH  /api/v1/admin/library/books/:id/status
DELETE /api/v1/admin/library/books/:id
```

状态流转：

```text
pending -> approved
pending -> rejected
approved -> hidden
hidden -> approved
approved -> deleted
hidden -> deleted
rejected -> deleted
```

规则：

- `rejected` 表示审核拒绝，普通用户不可见。
- `hidden` 表示下架，普通用户不可见，已有书架记录不可读。
- `deleted` 表示软删除，普通用户不可见，默认不返回。
- `DELETE` 默认软删除。
- `DELETE ?delete_file=true` 支持物理删除文件，并同步移除或标记相关书架引用不可用。

## 15. 阅读模块

### 15.1 阅读权限

用户可以读取图书的条件：

- 管理员可以读取所有未物理丢失的图书。
- 私有图书：当前用户存在该书 `bookshelves` 记录。
- 公共图书：当前用户已登录，且 `library_status='approved'`。
- 公共图书上传者可读取自己上传的 `pending` 图书。

公共图书 `hidden`、`deleted`、`rejected`、非上传者的 `pending` 均不可读。

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
    "content_type": "html",
    "content": "..."
  }
}
```

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

规则：

- 先做阅读权限校验。
- `href` 必填，表示 EPUB 内部资源相对路径或已规范化路径。
- 后端必须对 `href` 做规范化和路径穿越检查，禁止读取 EPUB 外部文件。
- 响应应返回资源原始字节和正确 `Content-Type`。
- 图片类型至少支持 JPEG、PNG、GIF、WEBP、SVG。
- 可返回 `Cache-Control: private, max-age=3600`，但不能绕过鉴权。
- 章节 HTML 中的 `<img src>` 应改写为该接口 URL，并对 `href` 做 URL 编码。

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
GET /api/v1/admin/system/storage
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
```

`/system/info` 匿名可访问，返回：

- `site_name`
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

规则：

- 设置修改立即生效。
- `max_upload_size_mb` 不能超过启动时 `REQUEST_BODY_LIMIT_MB`。
- 存储统计第一期可基于数据库 `file_size` 汇总，目录真实占用后续增强。

## 19. 路由总表

```text
/api/v1/health
/api/v1/system/info

/api/v1/auth/register
/api/v1/auth/login
/api/v1/auth/refresh
/api/v1/auth/logout
/api/v1/auth/me

/api/v1/users/me
/api/v1/users/me/password

/api/v1/bookshelf
/api/v1/bookshelf/:id
/api/v1/bookshelf/upload
/api/v1/bookshelf/from-library/:bookId

/api/v1/library/books
/api/v1/library/books/:id
/api/v1/library/books/upload
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

/api/v1/admin/users
/api/v1/admin/users/:id
/api/v1/admin/users/:id/status
/api/v1/admin/users/:id/role
/api/v1/admin/library/books
/api/v1/admin/library/books/:id/status
/api/v1/admin/library/books/:id
/api/v1/admin/categories
/api/v1/admin/categories/:id
/api/v1/admin/tags
/api/v1/admin/tags/:id
/api/v1/admin/system/storage
/api/v1/admin/system/settings
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
查看自己的信息                 否    是        是
修改自己的信息                 否    是        是
上传私有图书                   否    是        是
查看自己的书架                 否    是        是
阅读自己的私有图书             否    是        是
浏览 approved 公共图书         否    是        是
上传公共图书                   否    是        是
读取自己 pending 公共图书      否    是        是
加入公共图书到书架             否    是        是
管理自己的阅读进度             否    是        是
管理自己的书签                 否    是        是
管理用户                       否    否        是
审核公共图书                   否    否        是
隐藏/删除公共图书              否    否        是
管理系统分类                   否    否        是
管理系统标签                   否    否        是
管理系统设置                   否    否        是
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
- 阅读进度 upsert 正常。

手工验收：

1. Docker compose 启动成功。
2. 访问前端首页。
3. 注册第一个用户并确认角色为管理员。
4. 注册第二个用户并确认角色为普通用户。
5. 上传私有 EPUB，确认生成章节。
6. 上传 TXT，确认自动分章并能按章节读取。
7. 上传公共 PDF。
8. 第二个用户将公共 PDF 加入书架。
9. 磁盘上公共 PDF 只有一份。
10. 保存阅读进度。
11. 管理员隐藏公共图书，普通用户无法继续读取。

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
- 审核、隐藏、删除。

验收：

- 公共图书加入书架不复制文件。
- 审核开关生效。
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
- 修改公共图书审核开关后立即生效。

### 阶段八：增强

交付：

- bookmarks。
- temp 清理任务。
- 更完整日志。
- 备份恢复文档。
- 管理员维护 TXT 目录规则。
- 用户对单本书重新解析目录。

## 26. Docker 部署

部署形态：

```text
book-reader-app        Go 后端 + 前端静态资源
book-reader-postgres   PostgreSQL 数据库
```

应用容器挂载：

```text
/data
```

Dockerfile 使用多阶段构建：

1. Node 阶段构建前端。
2. Go 阶段构建后端并拷贝前端 `dist`。
3. Alpine 或 distroless 运行阶段只包含二进制和必要目录。

后端第一期可使用 Go `embed` 托管前端资源。若前端尚未创建，提供占位 `index.html`。

## 27. 已确认实现决策

后续开发必须遵守：

1. 私有图书从书架删除时，物理删除真实文件。
2. `pending` 是公共图书待审核状态，不是上传中状态。
3. 公共图书上传者在 pending 状态下允许自己查看和读取。
4. 用户上传公共图书后，不自动加入自己的书架。
5. 公共图书加入个人书架只创建引用，不复制文件。
6. Refresh Token 浏览器端默认使用 HttpOnly Cookie，开发调试可兼容 JSON body。
7. MVP 需要解析 EPUB/PDF/TXT 元数据、封面和章节目录。
8. TXT 必须解析章节目录，前端先请求章节目录，再按章节请求正文。
9. 分类和标签由管理员统一维护，普通用户只能从已有分类、标签中选择。
10. 管理员删除公共图书时支持物理删除文件选项。
11. 需要用户存储配额，管理员可调整单个用户配额。
12. 不允许匿名浏览公共图书馆；公共图书馆仅登录用户可访问。
13. 所有真实文件路径必须由 `storage` 模块生成和校验。
14. 所有阅读文件、封面和资源访问都必须先鉴权。
