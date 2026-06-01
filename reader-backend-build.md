# 阅读器后端详细构建文档

本文档基于 `reader-project-plan.md` 编写，用于后续后端开发落地。若本文档与总规划冲突，以总规划的核心业务目标为准；若总规划未明确，本文档给出已确认的后端实现决策。

## 1. 后端目标

后端负责提供以下能力：

- 多用户注册、登录、退出、刷新登录状态。
- 第一个注册用户自动成为管理员。
- 用户、角色、禁用状态管理。
- 私有图书上传、管理、删除、阅读访问。
- 公共图书馆上传、审核、隐藏、删除、加入个人书架。
- 公共图书引用式加入书架，不复制文件。
- 分类、标签、个人书架筛选。
- EPUB、PDF、TXT 文件访问和阅读进度保存。
- 书签接口，作为后续增强功能实现。
- PostgreSQL 持久化。
- 本地磁盘 volume 文件存储。
- Docker 部署，后端同时托管前端静态资源。

后端第一目标是稳定、简单、资源占用低。第一期不引入 Redis、Elasticsearch、MQ、对象存储、全文搜索服务、重型转码服务。

## 2. 技术栈

固定技术栈：

- 语言：Go 1.23 或更新的稳定版本。
- HTTP 框架：Gin。
- ORM：GORM。
- 数据库：PostgreSQL 16 或 17。
- 认证：JWT Access Token + Refresh Token。
- 密码哈希：bcrypt，后续如有需要再切换 argon2id。
- 日志：标准库 `log/slog`。
- 配置：环境变量。
- 文件存储：本地磁盘 `/data`。
- 数据库迁移：SQL migration 文件，应用启动时自动执行或通过独立命令执行。
- 部署：Docker 多阶段构建 + docker-compose。

默认第三方库建议：

- `github.com/gin-gonic/gin`
- `gorm.io/gorm`
- `gorm.io/driver/postgres`
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- `github.com/google/uuid`
- `github.com/joho/godotenv`，仅开发环境可选。

迁移工具二选一：

- 推荐：`github.com/golang-migrate/migrate/v4`，迁移边界清晰。
- 备选：自研简单 migration runner，维护 `schema_migrations` 表。

第一期建议使用 `golang-migrate`，不要依赖 GORM AutoMigrate 作为正式 schema 管理方式。GORM AutoMigrate 容易在后续变更中产生不可控差异，SQL migration 更适合作为多人协作和部署依据。

## 3. 项目目录

后端目录按业务模块拆分，每个模块内部保持简单，不强制过度分层。

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
      text.go
      dto.go
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
    000001_init.up.sql
    000001_init.down.sql
  tests/
    integration/
  go.mod
  go.sum
```

约定：

- `handler` 只处理 HTTP 参数、认证上下文、响应，不写业务规则。
- `service` 处理业务规则、事务、权限组合。
- `repository` 只处理数据库查询和持久化。
- `model` 定义数据库模型和常量。
- `dto` 定义请求和响应结构。
- 跨模块通用错误、分页、响应格式放在 `common`。
- 文件路径、落盘、移动、删除统一走 `storage`，业务模块不直接拼接真实路径。

## 4. 配置项

后端通过环境变量读取配置。

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
ALLOW_REGISTRATION=true
LIBRARY_REVIEW_REQUIRED=false

LOG_LEVEL=info
REQUEST_BODY_LIMIT_MB=110
TXT_CHUNK_SIZE=65536
```

默认规则：

- `BOOKS_DIR`、`COVERS_DIR`、`TEMP_DIR` 可以不配置，由 `DATA_DIR` 推导。
- 生产环境 `JWT_SECRET` 不允许使用 `change-me`。
- `MAX_UPLOAD_SIZE_MB` 控制单个图书文件最大大小。
- `REQUEST_BODY_LIMIT_MB` 应略大于 `MAX_UPLOAD_SIZE_MB`。
- `ALLOW_REGISTRATION=false` 时，除非系统没有任何用户，否则拒绝注册。
- 系统没有任何用户时，允许创建第一个管理员，即使 `ALLOW_REGISTRATION=false`。这是为了避免初始化死锁。
- `LIBRARY_REVIEW_REQUIRED=true` 时，用户上传公共图书默认 `pending`，管理员审核后才对普通用户可见。

## 5. HTTP 约定

### 5.1 基础路径

API 基础路径固定为：

```text
/api/v1
```

健康检查：

```http
GET /api/v1/health
```

### 5.2 响应格式

成功响应：

```json
{
  "data": {},
  "request_id": "req_xxx"
}
```

列表响应：

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

### 5.3 错误码

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
```

### 5.4 分页和排序

统一 query 参数：

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
- `sort` 必须在接口允许列表中，不能直接拼接用户输入。
- `order` 只允许 `asc`、`desc`。

### 5.5 认证头

Access Token 使用：

```http
Authorization: Bearer <access_token>
```

Refresh Token 第一期开发表现：

- 使用 HttpOnly Cookie 保存 refresh token，降低 XSS 泄露长期凭证的风险。
- Cookie 属性：`HttpOnly`、`SameSite=Lax`；生产 HTTPS 下启用 `Secure`。
- `/api/v1/auth/refresh` 和 `/api/v1/auth/logout` 默认从 Cookie 读取 refresh token。
- 开发和 API 调试阶段可兼容 JSON body 传递 refresh token，但浏览器端以前者为准。

## 6. 数据库设计

数据库以 `reader-project-plan.md` 中的表结构为基础，后端落地时需要补充必要约束。

### 6.1 users

字段按总规划实现。

补充字段建议：

```sql
ALTER TABLE users ADD COLUMN storage_quota_bytes BIGINT;
```

说明：

- `storage_quota_bytes` 为空时使用系统默认配额。
- 管理员可以为单个用户调整配额。
- 用户已使用空间按其拥有的未软删除图书 `books.file_size` 汇总计算，第一期不单独维护 `storage_used_bytes`，避免计数不一致。

补充约束：

```sql
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user'));
ALTER TABLE users ADD CONSTRAINT chk_users_status CHECK (status IN ('active', 'disabled'));
CREATE UNIQUE INDEX idx_users_email_unique_not_null ON users(email) WHERE email IS NOT NULL;
```

说明：

- `username` 必填且唯一。
- `email` 可为空，但非空时唯一。
- 禁用用户不能登录、刷新 token、访问受保护接口。
- 禁用不删除用户历史数据。

### 6.2 refresh_tokens

补充字段建议：

```sql
ALTER TABLE refresh_tokens ADD COLUMN user_agent TEXT;
ALTER TABLE refresh_tokens ADD COLUMN ip_address VARCHAR(64);
```

说明：

- 数据库存储 refresh token 的 hash，不存原文。
- 登出时设置 `revoked_at`。
- 刷新 token 时建议轮换 refresh token：旧 token 立即 revoke，新 token 入库。

### 6.3 books

补充约束：

```sql
ALTER TABLE books ADD CONSTRAINT chk_books_format CHECK (format IN ('epub', 'pdf', 'txt'));
ALTER TABLE books ADD CONSTRAINT chk_books_visibility CHECK (visibility IN ('private', 'public'));
ALTER TABLE books ADD CONSTRAINT chk_books_library_status CHECK (
  library_status IS NULL OR library_status IN ('pending', 'approved', 'rejected', 'hidden', 'deleted')
);
```

默认规则：

- 私有图书 `visibility='private'`，`library_status=NULL`。
- 公共图书 `visibility='public'`。
- 公共图书无需审核时 `library_status='approved'`。
- 公共图书需要审核时 `library_status='pending'`。
- `pending` 表示“待管理员审核”，不是“上传中”。上传本身是同步请求，请求成功后文件已经落盘并写入数据库。
- `deleted_at` 用于软删除。
- `file_path` 保存相对路径，例如 `books/private/user-1/100.epub`，不保存 `/data` 绝对路径。
- `cover_path` 保存相对路径，例如 `covers/100.jpg`。

### 6.4 bookshelves

补充约束：

```sql
ALTER TABLE bookshelves ADD CONSTRAINT chk_bookshelves_source_type CHECK (source_type IN ('uploaded', 'library'));
ALTER TABLE bookshelves ADD CONSTRAINT chk_bookshelves_status CHECK (status IN ('active', 'removed', 'unavailable'));
```

默认规则：

- 用户上传私有图书时，同时创建 `bookshelves` 记录，`source_type='uploaded'`。
- 用户上传公共图书时，不自动加入自己的书架。
- 上传者如需把自己上传的公共图书放入个人书架，也走“加入书架”接口。
- 从公共图书馆加入书架时，只创建 `bookshelves` 记录，`source_type='library'`。
- 用户移除公共图书时，将 `bookshelves.status='removed'` 并设置 `removed_at`，或直接硬删除该书架记录。第一期建议硬删除，逻辑简单。
- 公共图书被管理员隐藏或软删除后，用户书架查询时仍可返回记录但标记不可读。第一期建议保留书架记录并通过 book 状态判断不可读。

### 6.5 categories 和 tags

补充唯一约束：

```sql
CREATE UNIQUE INDEX idx_categories_system_slug ON categories(slug) WHERE scope = 'system';
CREATE UNIQUE INDEX idx_categories_user_slug ON categories(owner_user_id, slug) WHERE scope = 'user';
CREATE UNIQUE INDEX idx_tags_system_slug ON tags(slug) WHERE scope = 'system';
CREATE UNIQUE INDEX idx_tags_user_slug ON tags(owner_user_id, slug) WHERE scope = 'user';
```

默认规则：

- 分类和标签由管理员统一维护。
- 普通用户不能创建新的分类或标签。
- 普通用户只能从管理员已创建的分类、标签中选择，并打到自己的书架图书或自己上传的图书上。
- 第一期开发表现只启用 `scope='system'`，`scope='user'` 字段保留但不开放接口。

### 6.6 book_chapters

为减少阅读流量并支持按章节请求内容，MVP 必须生成章节目录。

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

字段含义：

- `chapter_index` 是从 0 开始的章节顺序。
- `locator` 是统一定位符。EPUB 使用 href 或 CFI 辅助定位，PDF 使用 `pdf_{segment}`，TXT 使用字节偏移范围标识。
- `start_offset`、`end_offset` 主要用于 TXT。
- `href`、`start_fragment_id`、`end_fragment_id`、`next_locator` 主要用于 EPUB。
- `is_volume` 表示卷标题或无正文目录节点。

解析失败策略：

- 上传成功不应因为章节解析失败而整体失败，除非文件本身不可读或格式不合法。
- 解析失败时至少创建一个兜底章节，避免前端无法打开。
- 解析错误写入日志，并可在后续增加 `books.parse_status` 字段；MVP 可先不加。

### 6.7 reading_progress

默认规则：

- 一个用户对一本书只有一条阅读进度。
- `bookshelf_id` 可空，因为公共图书软删除或书架记录变化时仍可能保留历史进度。
- `progress_type` 允许：`epub_cfi`、`pdf_page`、`txt_offset`。
- `percentage` 范围 `0.000` 到 `100.000`。
- 保存进度时更新对应书架的 `last_read_at`。

### 6.8 bookmarks

默认规则：

- 书签第一期可实现后端接口和数据模型，前端可后续接入。
- 位置类型与阅读进度一致：`epub_cfi`、`pdf_page`、`txt_offset`。
- 用户只能管理自己的书签。

### 6.9 system_settings

建议初始化键：

```text
allow_registration=true
library_review_required=false
site_name=Book Reader
max_upload_size_mb=100
default_user_storage_quota_mb=10240
```

说明：

- 环境变量是启动默认值。
- 管理后台修改后写入 `system_settings`。
- 运行时读取设置时，数据库设置优先于环境变量。
- `default_user_storage_quota_mb` 是普通用户默认存储配额，管理员可为单个用户覆盖。

## 7. 核心枚举

后端代码中统一定义枚举常量，不散落字符串。

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

CategoryScopeSystem=system
CategoryScopeUser=user
TagScopeSystem=system
TagScopeUser=user

ProgressTypeEpubCFI=epub_cfi
ProgressTypePDFPage=pdf_page
ProgressTypeTXTOffset=txt_offset
```

## 8. 启动流程

服务启动顺序：

1. 加载配置。
2. 初始化 logger。
3. 校验生产环境必要配置，例如 `JWT_SECRET`。
4. 创建数据目录：`/data/books/private`、`/data/books/public`、`/data/covers`、`/data/temp`。
5. 连接 PostgreSQL。
6. 执行数据库迁移。
7. 初始化系统设置默认值。
8. 初始化 Gin router 和中间件。
9. 注册 API 路由。
10. 注册前端静态资源路由。
11. 启动 HTTP server。
12. 接收到退出信号时优雅关闭。

优雅关闭：

- 停止接收新请求。
- 等待正在处理的请求完成，最多 10 到 30 秒。
- 关闭数据库连接池。

## 9. 中间件

### 9.1 Request ID

每个请求生成或透传 `X-Request-ID`。

响应头返回：

```http
X-Request-ID: req_xxx
```

日志和错误响应都包含 `request_id`。

### 9.2 Logger

记录：

- request_id
- method
- path
- status
- latency_ms
- user_id，若已认证
- client_ip
- user_agent

不记录：

- 密码。
- token 原文。
- 上传文件内容。

### 9.3 Recover

捕获 panic，记录堆栈，返回 `internal_error`。

### 9.4 AuthRequired

处理逻辑：

1. 读取 `Authorization` header。
2. 校验 Bearer token。
3. 解析 user id、role、token type、expires_at。
4. 查询用户是否存在且状态为 `active`。
5. 将用户信息写入 Gin context。

Access Token 中必须包含：

```json
{
  "sub": "user_id",
  "role": "user",
  "typ": "access",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### 9.5 AdminRequired

要求已认证且 `role='admin'`。

### 9.6 Body Limit

上传接口限制请求体大小，大小取 `REQUEST_BODY_LIMIT_MB`。

## 10. 认证模块

### 10.1 注册

接口：

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

响应：

```json
{
  "data": {
    "user": {
      "id": 1,
      "username": "alice",
      "email": "alice@example.com",
      "nickname": "Alice",
      "role": "admin",
      "status": "active"
    },
    "access_token": "...",
    "refresh_token": "...",
    "expires_in": 7200
  }
}
```

校验：

- `username` 长度 3 到 64，只允许字母、数字、下划线、短横线。
- `email` 可空，非空时必须是合法邮箱格式。
- `password` 长度至少 8。
- `nickname` 可空，最大 128。

第一个管理员规则：

- 注册逻辑必须在数据库事务中完成。
- 事务内执行全局 advisory lock，避免并发注册创建多个管理员。
- PostgreSQL 推荐使用：`SELECT pg_advisory_xact_lock(10001)`。
- 加锁后统计 `users` 表用户数。
- 用户数为 0 时，新用户角色为 `admin`。
- 用户数大于 0 时，若注册关闭则返回 `registration_disabled`，否则角色为 `user`。

### 10.2 登录

接口：

```http
POST /api/v1/auth/login
```

请求：

```json
{
  "username": "alice",
  "password": "password123"
}
```

说明：

- `username` 可以是用户名，也可以是邮箱。
- 密码错误统一返回 `invalid_credentials`，不暴露用户是否存在。
- 用户禁用返回 `user_disabled`。
- 登录成功更新 `last_login_at`。

### 10.3 刷新 token

接口：

```http
POST /api/v1/auth/refresh
```

请求：

```json
{
  "refresh_token": "..."
}
```

规则：

- Refresh token 原文只返回给客户端一次。
- 数据库存储 SHA-256 hash。
- 刷新时查 hash，检查未过期、未撤销。
- 用户被禁用时拒绝刷新。
- 刷新成功后撤销旧 refresh token，创建新 refresh token。

### 10.4 退出

接口：

```http
POST /api/v1/auth/logout
```

请求：

```json
{
  "refresh_token": "..."
}
```

规则：

- 撤销当前 refresh token。
- Access token 不做服务端黑名单，等待自然过期。

### 10.5 当前用户

接口：

```http
GET /api/v1/auth/me
```

返回当前认证用户信息。

## 11. 用户模块

### 11.1 当前用户资料

```http
GET   /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password
```

可修改字段：

- `nickname`
- `email`
- `avatar_path`，第一期可先不开放头像上传，仅保留字段。

修改密码规则：

- 需要旧密码。
- 新密码至少 8 位。
- 修改成功后可选择撤销该用户所有 refresh token。第一期建议撤销全部，安全性更高。

### 11.2 管理员用户管理

```http
GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
```

规则：

- 管理员不能禁用自己。
- 管理员不能把自己降级为普通用户，除非系统中还有其他管理员。第一期建议直接禁止管理员修改自己的角色。
- 被禁用用户不能访问受保护 API。
- 禁用用户时撤销该用户所有 refresh token。

## 12. 文件存储模块

### 12.1 目录结构

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
- 对所有最终路径做 `filepath.Clean` 和目录前缀检查。

### 12.2 上传格式校验

第一期支持：

```text
epub
pdf
txt
```

校验规则：

- 扩展名白名单。
- MIME 仅作辅助，不作为唯一依据。
- 文件头魔数校验：
  - PDF：以 `%PDF-` 开头。
  - EPUB：ZIP 文件，且包含 `mimetype` 或 `META-INF/container.xml`。第一期可只校验 ZIP 头和扩展名，后续增强。
  - TXT：不强依赖魔数，限制为非二进制高概率文本。
- 文件大小不超过 `MAX_UPLOAD_SIZE_MB`。
- 空文件拒绝。

### 12.3 上传流程

私有上传和公共上传共用底层流程：

1. 认证用户发起 multipart upload。
2. 校验请求体大小。
3. 创建 temp 文件。
4. 流式读取上传内容并写入 temp 文件。
5. 同时计算 SHA-256 和文件大小。
6. 校验扩展名、格式、大小。
7. 开启数据库事务。
8. 插入 `books` 记录，先写临时路径或空正式路径。
9. 根据 `book_id` 生成正式相对路径。
10. 移动 temp 文件到正式路径。
11. 更新 `books.file_path`。
12. 创建必要的 `bookshelves` 记录。
13. 提交事务。

失败处理：

- temp 文件写入失败：返回错误，删除残留 temp 文件。
- 数据库插入失败：删除 temp 文件。
- 文件移动失败：回滚事务，删除 temp 文件。
- 事务提交失败：删除正式文件。

说明：

- 文件移动无法被数据库事务自动回滚，因此 service 必须显式补偿。
- 第一阶段不做跨用户文件去重，即使 hash 相同也保存多份私有文件。
- 公共图书加入个人书架时不复制文件。

## 13. 图书和书架模块

### 13.1 私有图书上传

接口：

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

### 13.2 我的书架列表

接口：

```http
GET /api/v1/bookshelf
```

支持 query：

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

返回字段建议：

```json
{
  "id": 10,
  "book_id": 100,
  "title": "Book Title",
  "author": "Author",
  "format": "epub",
  "cover_url": "/api/v1/reader/books/100/cover",
  "source_type": "uploaded",
  "visibility": "private",
  "library_status": null,
  "favorite": false,
  "pinned": false,
  "last_read_at": null,
  "added_at": "2026-05-31T00:00:00Z",
  "readable": true
}
```

规则：

- 只返回当前用户书架记录。
- 若关联公共图书被隐藏、软删除，返回 `readable=false`。
- 默认排序：`pinned desc`、`last_read_at desc nulls last`、`added_at desc`。

### 13.3 书架详情

```http
GET /api/v1/bookshelf/:id
```

规则：

- 用户只能访问自己的书架记录。
- 管理员不通过该接口查看他人书架，管理员另走管理接口。

### 13.4 从公共图书加入书架

```http
POST /api/v1/bookshelf/from-library/:bookId
POST /api/v1/library/books/:id/add-to-bookshelf
```

两个接口语义重复，第一期建议都路由到同一个 service。

规则：

- 目标书籍必须存在。
- `visibility='public'`。
- `library_status='approved'`。
- 未软删除。
- 当前用户没有同一本书的书架记录。
- 只创建 `bookshelves`，不复制文件，不修改 `books.file_path`。

### 13.5 编辑书架个性化信息

```http
PATCH /api/v1/bookshelf/:id
```

可编辑：

- `personal_title`
- `personal_category_id`
- `favorite`
- `pinned`
- 个人标签集合

不可编辑：

- 真实 `books.file_path`
- `books.owner_user_id`
- `books.visibility`

### 13.6 移除书架图书

```http
DELETE /api/v1/bookshelf/:id
```

私有上传图书：

- 如果 `books.visibility='private'` 且 `books.owner_user_id=current_user_id`，删除书架记录，软删除 `books`，并物理删除真实文件，释放空间。
- 物理删除失败时返回错误，不提交数据库删除状态，避免数据库显示已删除但文件仍占用空间。

公共图书：

- 只删除当前用户 `bookshelves` 记录。
- 不删除 `books`。
- 不删除真实文件。

## 14. 公共图书馆模块

### 14.1 公共图书列表

```http
GET /api/v1/library/books
```

普通用户只看到：

- `visibility='public'`
- `library_status='approved'`
- `deleted_at IS NULL`

支持 query：

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

返回中包含：

- 当前用户是否已加入书架：`in_bookshelf`。
- 当前用户对应书架 id：`bookshelf_id`，未加入时为空。

### 14.2 公共图书详情

```http
GET /api/v1/library/books/:id
```

规则同列表，普通用户只能查看 approved 公共图书。

### 14.3 上传公共图书

```http
POST /api/v1/library/books/upload
```

创建结果：

- `books.visibility='public'`
- `books.owner_user_id=current_user_id`
- `books.library_status='pending'` 或 `approved`
- 文件路径：`books/public/{book_id}.{ext}`
- 不自动给上传者创建书架记录。

若 `LIBRARY_REVIEW_REQUIRED=true`：

- 普通用户上传后状态为 `pending`。
- `pending` 是待管理员审核状态，不是上传中状态；接口返回成功时文件已经上传完成。
- 上传者可以查看和读取自己上传的 pending 公共图书，方便确认内容，但该书不会自动出现在个人书架。
- 其他普通用户不可见。
- 管理员可见并审核。

若 `LIBRARY_REVIEW_REQUIRED=false`：

- 上传后状态为 `approved`。
- 所有登录用户可见。

### 14.4 管理员公共图书管理

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
- `hidden` 表示已下架，普通用户不可见，已有书架记录不可读。
- `deleted` 表示软删除，普通用户不可见，默认不返回。
- `DELETE` 默认软删除：设置 `library_status='deleted'` 和 `deleted_at`。
- 管理员删除公共图书时支持物理删除文件，建议使用 `DELETE /api/v1/admin/library/books/:id?delete_file=true`。
- 物理删除公共图书时，应同步移除或标记所有相关书架引用不可用，并删除真实文件和封面文件。

## 15. 分类模块

### 15.1 查询分类

```http
GET /api/v1/categories
```

query：

```text
scope=system
```

规则：

- 普通用户可查询管理员维护的系统分类。
- 普通用户不能创建个人分类。
- 管理员通过管理接口维护系统分类。

### 15.2 管理系统分类

```http
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

规则：

- 系统分类 `scope='system'`，`owner_user_id=NULL`。
- 删除分类前检查是否有关联图书。
- 第一期若有关联则拒绝删除，避免产生悬挂筛选。

## 16. 标签模块

### 16.1 查询标签

```http
GET /api/v1/tags
```

query：

```text
scope=system
keyword=xxx
```

规则：

- 普通用户可查询管理员维护的系统标签。
- 普通用户不能创建个人标签。
- 用户给图书打标签时，只能从已有系统标签中选择。

### 16.2 管理标签

```http
POST   /api/v1/admin/tags
PATCH  /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

规则：

- 只有管理员可以创建、修改、删除标签。
- 管理员创建的标签 `scope='system'`，`owner_user_id=NULL`。
- 删除标签前检查是否有关联图书或书架记录。
- 第一期若有关联则拒绝删除，避免用户已有打标失效。

## 17. 阅读模块

### 17.1 阅读权限

用户可以读取图书文件的条件：

- 管理员可以读取所有未物理丢失的图书。
- 私有图书：当前用户存在 `bookshelves.user_id=current_user_id AND book_id=:bookId`。
- 公共图书：当前用户已登录，且 `library_status='approved'`。
- 公共图书上传者在 pending 状态下可以读取自己上传的书。

说明：

- 总规划中写“私有图书只有拥有书架记录的用户可以读取”，因此以书架关系作为私有访问依据。
- 公共图书对所有登录用户可读，但 hidden/deleted/rejected/pending 默认不可读，上传者 pending 例外。

### 17.2 图书元数据

```http
GET /api/v1/reader/books/:bookId/meta
```

返回：

```json
{
  "data": {
    "id": 100,
    "title": "Book Title",
    "author": "Author",
    "format": "epub",
    "file_size": 123456,
    "cover_url": "/api/v1/reader/books/100/cover",
    "progress": {
      "progress_type": "epub_cfi",
      "progress_value": "...",
      "percentage": 12.3
    }
  }
}
```

### 17.3 文件读取

```http
GET /api/v1/reader/books/:bookId/file
```

规则：

- 先做阅读权限校验。
- 使用 `http.ServeContent` 或等价实现支持 Range 请求。
- PDF 必须支持 Range。
- EPUB 可直接返回完整文件，由前端 `epub.js` 处理。
- 响应 `Content-Type`：
  - EPUB：`application/epub+zip`
  - PDF：`application/pdf`
  - TXT：`text/plain; charset=utf-8` 或 `application/octet-stream`
- 设置 `Accept-Ranges: bytes`。
- 不暴露真实文件路径。

### 17.4 封面读取

```http
GET /api/v1/reader/books/:bookId/cover
```

规则：

- 有封面文件则返回封面。
- 没有封面则返回 404 或默认封面。第一期建议返回 404，由前端展示默认封面。
- 封面访问同样要鉴权，避免通过封面枚举私有图书。

### 17.5 TXT 分块读取

```http
GET /api/v1/reader/books/:bookId/text?offset=0&limit=65536
```

规则：

- 只允许 TXT 格式。
- `offset` 是字节偏移。
- `limit` 默认 `TXT_CHUNK_SIZE`，最大 1MB。
- 返回实际读取字节范围、是否 EOF。
- 该接口作为章节内容接口的底层能力保留，用于调试、兜底和超大 TXT 局部读取。

响应：

```json
{
  "data": {
    "offset": 0,
    "limit": 65536,
    "next_offset": 65536,
    "eof": false,
    "content": "..."
  }
}
```

说明：

- MVP 必须生成 TXT 章节目录，前端优先按章节请求内容，减少不必要的数据传输。
- TXT 章节读取仍以字节偏移为准，避免多字节编码和大文件读取错位。

### 17.6 章节目录和章节内容

```http
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
```

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

规则：

- 所有格式都必须提供章节目录接口。
- EPUB 章节内容返回清洗后的 HTML 子集，保留必要图片引用。
- TXT 章节内容按 `book_chapters.start_offset/end_offset` 读取，不一次性加载全书。
- PDF 章节可以按每 10 页一个分段返回页面范围，前端仍通过 PDF 文件 Range 渲染。
- `is_volume=true` 的章节允许返回空内容。
- 前端阅读优先调用章节目录和章节内容接口，而不是直接拉取完整 TXT。

### 17.7 阅读进度

```http
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
```

保存请求：

```json
{
  "progress_type": "epub_cfi",
  "progress_value": "epubcfi(/6/2!...)",
  "percentage": 12.345
}
```

校验：

- EPUB 只允许 `epub_cfi`。
- PDF 只允许 `pdf_page`。
- TXT 只允许 `txt_offset`。
- `percentage` 可空，非空范围 `0` 到 `100`。
- 当前用户必须有权限阅读该书。

保存逻辑：

- `INSERT ... ON CONFLICT (user_id, book_id) DO UPDATE`。
- 同步更新当前用户对应 `bookshelves.last_read_at`。

## 18. 书签模块

接口：

```http
GET    /api/v1/books/:bookId/bookmarks
POST   /api/v1/books/:bookId/bookmarks
PATCH  /api/v1/bookmarks/:id
DELETE /api/v1/bookmarks/:id
```

规则：

- 用户必须有阅读权限才能创建和查看该书书签。
- 用户只能修改、删除自己的书签。
- 第一期开发表现可以先实现后端 CRUD，前端后续接入。

## 19. 系统模块

### 19.1 系统信息

```http
GET /api/v1/system/info
```

返回：

```json
{
  "data": {
    "site_name": "Book Reader",
    "allow_registration": true,
    "library_review_required": false,
    "supported_formats": ["epub", "pdf", "txt"],
    "max_upload_size_mb": 100,
    "default_user_storage_quota_mb": 10240
  }
}
```

匿名用户可访问，用于登录注册页展示。

### 19.2 存储统计

```http
GET /api/v1/admin/system/storage
```

返回：

- 图书总数。
- 公共图书数。
- 私有图书数。
- 数据库记录总文件大小。
- `/data/books` 实际占用大小。
- `/data/temp` 实际占用大小。

第一期可以只基于数据库统计 `file_size`，目录实际占用后续增强。

### 19.3 系统设置

```http
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
```

可修改：

- `allow_registration`
- `library_review_required`
- `site_name`
- `max_upload_size_mb`
- `default_user_storage_quota_mb`

规则：

- 修改立即生效。
- `max_upload_size_mb` 不能超过服务启动时 `REQUEST_BODY_LIMIT_MB`，否则实际上传仍会被请求体限制拦截。
- `default_user_storage_quota_mb` 控制普通用户默认存储配额；管理员仍可在用户管理中单独覆盖用户配额。

## 20. 路由注册

建议路由结构：

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

```text
GET /*path
```

规则：

- API 路由优先注册。
- 非 API 路径返回前端 `index.html`，支持 Vue Router history 模式。
- 静态资源不存在时，非 API 路径回退 `index.html`。
- `/api/*` 不允许回退到前端页面，应返回 JSON 404。

## 21. 权限矩阵

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
加入公共图书到书架             否    是        是
管理自己的阅读进度             否    是        是
管理自己的书签                 否    是        是
管理用户                       否    否        是
审核公共图书                   否    否        是
隐藏/删除公共图书              否    否        是
管理系统分类                   否    否        是
管理系统设置                   否    否        是
```

## 22. 事务边界

必须使用事务的场景：

- 注册用户，特别是第一个管理员判断。
- 登录刷新 refresh token 轮换。
- 私有图书上传：创建 book、移动文件、创建 bookshelf。
- 公共图书上传：创建 book、移动文件、解析元数据和章节；不创建上传者书架记录。
- 公共图书加入书架。
- 删除私有图书时，更新 book 和删除 bookshelf。
- 管理员删除公共图书时，更新 book 状态和相关书架可用性。
- 保存阅读进度并更新 `last_read_at`。

注意：

- 文件系统操作不能真正纳入数据库事务。
- 涉及文件的事务必须设计补偿逻辑。
- 不要在事务中执行慢速大文件上传；上传写 temp 应发生在事务前。

## 23. 图书解析处理

MVP 必须进行基础元数据、封面和章节目录解析。

上传时处理策略：

- 上传时允许用户传 `title`、`author`、`description`。
- 后端解析 EPUB/PDF/TXT 元数据，解析结果作为默认值。
- 用户上传时显式传入的 `title`、`author`、`description` 优先级高于解析结果。
- 如果解析和用户输入都没有 `title`，则使用原始文件名去掉扩展名。
- 后端尽量提取封面；提取失败时 `cover_path` 为空，由前端展示默认封面。
- 后端必须生成 `book_chapters`，前端基于章节目录请求内容，减少流量消耗。

格式解析要求：

- EPUB：解析 OPF metadata 获取标题、作者、简介、封面；解析 nav/ncx/spine 生成章节，支持 fragment 范围。
- PDF：读取 metadata，封面可渲染第一页；章节按每 10 页一个分段生成。
- TXT：识别编码，使用目录正则生成章节；无目录命中时按固定大小分段兜底。

解析失败处理：

- 元数据或封面解析失败不阻断上传。
- 章节解析失败时必须创建一个全书兜底章节。
- 文件格式不可读或不符合声明格式时，上传失败。

## 24. TXT 解析边界

总规划要求 TXT 阅读支持分块加载和进度保存。结合“按章节请求内容、减少流量消耗”的目标，MVP 必须实现 TXT 章节解析。

第一期实现：

- TXT 编码识别，至少支持 UTF-8、UTF-8 BOM、GBK/GB18030 常见中文文本。
- 移植或内置 Legado 默认目录正则中的启用规则。
- 解析章节时保存字节级 `start_offset` 和 `end_offset`。
- 前端先请求章节目录，再按章节请求正文。
- 进度保存 `txt_offset`，也可以附带章节 id 作为前端状态。
- 后端读取章节内容时按字节范围读取，不能一次性加载超大 TXT。

兜底策略：

- 如果没有匹配到目录规则，按固定字节大小分段生成章节。
- 章节标题格式为 `第 N 段`。
- 最后一段过小时合并到上一段。

后续增强：

- 允许管理员维护 TXT 目录规则。
- 允许用户对单本书重新解析目录。
- 支持用户手动合并、拆分、重命名章节。

## 25. 安全要求

### 25.1 密码

- 使用 bcrypt。
- 不记录明文密码。
- 不返回 password hash。
- 修改密码后撤销 refresh token。

### 25.2 Token

- Access token 短期有效，默认 2 小时。
- Refresh token 长期有效，默认 30 天。
- Refresh token 只存 hash。
- 生产环境 `JWT_SECRET` 必须强随机。

### 25.3 上传

- 限制扩展名。
- 限制大小。
- 禁止路径穿越。
- 不信任 MIME。
- 上传先写 temp。
- 校验后移动到正式目录。

### 25.4 文件访问

- 不暴露 `/data`。
- 所有文件访问走 API。
- 读取前执行权限校验。
- Range 请求也必须先鉴权。

### 25.5 日志

- 不记录 token。
- 不记录密码。
- 不记录上传文件内容。
- 错误日志记录 request_id，方便追踪。

## 26. 性能和资源控制

策略：

- 上传使用流式写入，不读入内存。
- 文件下载使用流式响应。
- PDF 文件接口支持 Range。
- TXT 分块读取。
- 查询列表必须分页。
- 数据库连接池限制最大连接数。
- 不做后端重型转码。
- temp 目录定期清理。

数据库连接池默认建议：

```text
max_open_conns=20
max_idle_conns=5
conn_max_lifetime=1h
```

第一期可以写死默认值，后续再暴露为环境变量。

## 27. 测试计划

### 27.1 单元测试

必须覆盖：

- 密码 hash 和校验。
- JWT 创建和解析。
- Refresh token hash。
- 上传格式判断。
- 路径生成和路径穿越防护。
- 分页参数解析。

### 27.2 集成测试

建议使用测试 PostgreSQL 容器或 docker-compose test profile。

必须覆盖：

- 第一个注册用户成为管理员。
- 并发注册不会产生多个管理员。
- 后续注册用户为普通用户。
- 禁用用户不能登录。
- 登录成功返回 access token 和 refresh token。
- Refresh token 可以轮换。
- 退出后 refresh token 失效。
- 私有图书上传创建 `books` 和 `bookshelves`。
- 私有图书只有拥有者能读取。
- 公共图书上传创建单份文件。
- 公共图书加入书架不复制文件。
- 重复加入公共图书返回 conflict。
- 从书架移除公共图书不删除文件。
- 管理员隐藏公共图书后普通用户不可读。
- PDF 文件接口支持 Range。
- TXT 分块读取不会一次性加载整本书。
- 阅读进度 upsert 正常。

### 27.3 手工验收

Docker compose 启动后验证：

1. 访问前端首页。
2. 注册第一个用户并确认角色为管理员。
3. 注册第二个用户并确认角色为普通用户。
4. 上传私有 EPUB。
5. 打开 EPUB 文件接口。
6. 上传公共 PDF。
7. 第二个用户将公共 PDF 加入书架。
8. 检查公共 PDF 在磁盘上只有一份。
9. 保存阅读进度。
10. 管理员隐藏公共图书，普通用户无法继续读取。

## 28. 开发阶段拆解

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
- 注册。
- 登录。
- refresh。
- logout。
- me。
- auth middleware。
- admin middleware。
- 管理员用户列表。

验收：

- 首个用户为管理员。
- 并发注册不产生多个管理员。
- 禁用用户无法登录和刷新。

### 阶段三：文件存储和私有书架

交付：

- books、bookshelves migration。
- storage 模块。
- 上传私有图书。
- 我的书架列表。
- 书架详情。
- 移除图书。
- 文件读取权限。

验收：

- 上传文件落盘。
- 数据库记录正确。
- 非拥有者无法读取私有图书。

### 阶段四：阅读接口

交付：

- reader meta。
- reader file，支持 Range。
- reader text 分块。
- book_chapters migration。
- EPUB/PDF/TXT 元数据、封面和章节解析。
- 章节目录接口。
- 章节内容接口。
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
- 管理员标签 CRUD。
- 用户从管理员维护的分类、标签中选择并打标。
- 书架和图书馆筛选。

验收：

- 系统分类唯一约束生效。
- 普通用户不能创建分类或标签。
- 用户只能使用已有系统分类、系统标签。

### 阶段七：管理和系统设置

交付：

- system_settings。
- storage 统计。
- 系统设置读取和修改。
- 管理员用户状态和角色调整。

验收：

- 关闭注册后非首个用户无法注册。
- 修改公共图书审核开关后立即生效。

### 阶段八：补充增强

交付：

- bookmarks。
- temp 清理。
- 更完整日志。
- 备份恢复文档。
- 管理员维护 TXT 目录规则。
- 用户对单本书重新解析目录。

## 29. Docker 构建

后端构建产物为单二进制 `server`。

Dockerfile 按总规划使用多阶段构建：

1. Node 阶段构建前端。
2. Go 阶段构建后端并拷贝前端 `dist`。
3. Alpine 运行阶段只包含二进制和必要目录。

后端需要支持从内嵌目录或磁盘目录托管前端资源。第一期建议使用 Go `embed`：

```text
backend/internal/app/static/
```

构建时将前端 `dist` 拷贝到该目录，再由 Go 编译进二进制。这样运行镜像只需要一个 server 文件。

如果前端尚未创建，后端开发阶段可以提供简单占位 `index.html`。

## 30. 迁移文件初始规划

初始 migration 建议拆分：

```text
000001_create_users.up.sql
000002_create_refresh_tokens.up.sql
000003_create_books.up.sql
000004_create_bookshelves.up.sql
000005_create_book_chapters.up.sql
000006_create_categories_tags.up.sql
000007_create_reading_progress.up.sql
000008_create_bookmarks.up.sql
000009_create_system_settings.up.sql
```

也可以第一期合并为 `000001_init.up.sql`。若项目刚开始，合并简单；若多人并行开发，拆分更清晰。

建议第一期使用拆分 migration，便于按阶段开发和回滚。

## 31. 已确认实现决策

以下边界已确认，后续开发按此实现：

1. 私有图书从书架删除时，物理删除真实文件。
2. `pending` 是公共图书待管理员审核状态，不是上传中状态；上传接口成功返回时文件已经上传完成。
3. 公共图书上传者在 pending 状态下允许自己查看和读取，用于确认内容。
4. 用户上传公共图书后，不自动加入自己的书架；需要时手动调用加入书架接口。
5. Refresh Token 使用 HttpOnly Cookie 作为浏览器端默认方案，开发调试可兼容 JSON body。
6. MVP 需要解析 EPUB/PDF/TXT 元数据、封面和章节目录。
7. TXT 必须解析章节目录；前端先请求章节目录，再按章节请求正文，减少流量消耗。
8. 分类和标签由管理员统一维护，普通用户只能从已有分类、标签中选择并打标。
9. 管理员删除公共图书时支持物理删除文件选项。
10. 需要用户存储配额，管理员可调整单个用户配额。
11. 不允许匿名浏览公共图书馆；公共图书馆仅登录用户可访问。
