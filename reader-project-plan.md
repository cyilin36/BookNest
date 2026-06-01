# 阅读器项目总规划文档

## 1. 项目目标

本项目是一个支持多用户的 Web 阅读器系统，采用前后端分离开发、统一打包部署的架构。

系统需要支持用户注册、登录、图书上传、个人书架、在线阅读、公共图书馆、分类标签、管理员管理、Docker 部署等功能。

项目核心特色是公共图书馆能力：用户可以上传图书作为公共图书，其他用户可以将公共图书加入自己的书架，但加入时只建立引用关系，不复制图书文件，避免重复存储和磁盘压力。

## 2. 核心需求

### 2.1 用户系统

- 支持多用户注册。
- 第一个注册用户自动成为管理员。
- 后续注册用户默认为普通用户。
- 支持登录、退出、刷新登录状态。
- 支持管理员禁用用户。
- 支持管理员调整用户角色。

### 2.2 图书管理

- 用户可以上传图书到个人书架。
- 用户可以上传图书到公共图书馆。
- 图书支持分类。
- 图书支持标签。
- 支持编辑图书基础元数据。
- 支持删除或移除图书。

### 2.3 个人书架

- 用户可以查看自己的图书。
- 用户可以阅读自己上传的私有图书。
- 用户可以将公共图书添加到自己的书架。
- 用户从书架移除公共图书时，只删除个人引用，不删除公共图书文件。
- 支持阅读进度保存。
- 支持最近阅读。
- 支持按分类、标签、名称搜索。

### 2.4 公共图书馆

- 用户可以上传公共图书。
- 所有登录用户可以浏览公共图书。
- 所有登录用户可以将公共图书加入自己的书架。
- 加入书架时只保存关联关系，不复制文件。
- 管理员可以审核、隐藏、删除公共图书。
- 可配置公共图书是否需要审核。

### 2.5 在线阅读

- 支持 EPUB 阅读。
- 支持 PDF 阅读。
- 支持 TXT 阅读。
- 支持阅读进度保存。
- 支持继续阅读。
- 支持基础阅读设置，例如字号、主题、行距。
- 支持书签，作为可选增强功能。

### 2.6 部署要求

- 支持 Docker 部署。
- 前后端最终打包在同一个部署单元中。
- 使用 PostgreSQL 作为数据库。
- 不引入高内存占用组件。
- 默认不依赖 Redis、Elasticsearch、MQ、对象存储。

## 3. 推荐技术架构

### 3.1 后端技术栈

- 语言：Go
- Web 框架：Gin
- ORM：GORM
- 数据库：PostgreSQL
- 认证：JWT + Refresh Token
- 文件存储：本地磁盘 volume
- 日志：slog 或 zerolog
- 配置：环境变量
- 部署：Docker

选择 Go 的原因：

- 运行时资源占用低。
- 单二进制部署简单。
- 文件上传、下载、Range 请求支持直接。
- 适合做轻量 Web 服务。
- Docker 镜像可以控制得比较小。
- 比 Java / Spring Boot 更适合低内存部署目标。

### 3.2 前端技术栈

- 框架：Vue 3
- 构建工具：Vite
- 语言：TypeScript
- 路由：Vue Router
- 状态管理：Pinia
- UI 组件库：Naive UI
- 请求库：Axios
- EPUB 阅读：epub.js
- PDF 阅读：pdf.js
- TXT 阅读：自研分块加载与阅读组件

选择 Vue 3 的原因：

- 适合中后台和内容管理类 WebUI。
- 开发效率高。
- 生态成熟。
- Vite 构建快，方便嵌入后端静态资源。

### 3.3 数据库

数据库固定使用 PostgreSQL。

推荐 PostgreSQL 版本：

```text
PostgreSQL 16 或 PostgreSQL 17
```

使用 PostgreSQL 的原因：

- 多用户并发写入能力强于 SQLite。
- 更适合长期运行的多人系统。
- 标签、分类、搜索、统计能力更好扩展。
- 后续可以平滑支持全文搜索、复杂筛选、审计日志。

### 3.4 部署形态

推荐部署为两个容器：

```text
book-reader-app        Go 后端 + 前端静态资源
book-reader-postgres   PostgreSQL 数据库
```

说明：

- 应用前后端打包在一个 app 容器中。
- PostgreSQL 单独作为数据库容器。
- 图书文件使用 Docker volume 挂载到 app 容器。
- PostgreSQL 数据使用 Docker volume 挂载到 db 容器。
- 不额外引入 Redis、Nginx、MinIO。

## 4. 总体架构

```text
Browser
  |
  | HTTP / HTTPS
  v
book-reader-app
  |
  |-- Vue WebUI static files
  |-- REST API
  |-- JWT Auth
  |-- Upload Service
  |-- Reader File Service
  |-- Book Parser Service
  |-- Permission Service
  |
  | SQL
  v
PostgreSQL

book-reader-app
  |
  | local filesystem
  v
/data/books
/data/covers
/data/temp
```

最终访问入口：

```text
http://server:8080
```

同一个应用服务提供：

- 前端页面。
- 后端 API。
- 图书上传。
- 图书文件读取。
- 封面文件读取。
- 阅读器资源访问。

## 5. 项目目录规划

```text
book-reader/
  backend/
    cmd/
      server/
        main.go
    internal/
      app/
      config/
      database/
      middleware/
      auth/
      user/
      book/
      bookshelf/
      library/
      category/
      tag/
      reader/
      parser/
      upload/
      admin/
      storage/
      system/
      common/
    migrations/
    go.mod
    go.sum

  frontend/
    src/
      api/
      assets/
      components/
      layouts/
      router/
      stores/
      views/
        auth/
        bookshelf/
        library/
        reader/
        upload/
        admin/
        settings/
      main.ts
    index.html
    package.json
    vite.config.ts

  deploy/
    docker/
      Dockerfile
    docker-compose.yml

  docs/
    project-plan.md
    api.md
    database.md
    deployment.md
    security.md
    backup-restore.md

  README.md
```

## 6. 核心业务模型

### 6.1 用户模型

用户角色：

```text
admin
user
```

用户状态：

```text
active
disabled
```

规则：

- 系统首次启动时没有默认管理员。
- 第一个注册成功的用户自动成为管理员。
- 后续注册用户默认为普通用户。
- 第一个用户的管理员设置必须在数据库事务中完成，避免并发注册导致多个首位管理员。

### 6.2 图书模型

图书统一保存在 `books` 表中。

图书可见性：

```text
private   私有图书
public    公共图书馆图书
```

私有图书：

- 用户上传到自己的书架。
- 默认只有上传者可以访问。
- 文件路径记录在 `books.file_path` 中。

公共图书：

- 用户上传到公共图书馆。
- 所有登录用户可以浏览。
- 用户加入书架时不复制文件。
- 多个用户通过书架记录引用同一条 `books` 记录。

### 6.3 书架模型

个人书架使用 `bookshelves` 表表达。

每条书架记录代表：

```text
某个用户拥有某本书的阅读入口
```

书架来源：

```text
uploaded   用户自己上传
library    从公共图书馆加入
```

关键约束：

```text
unique(user_id, book_id)
```

作用：

- 防止用户重复添加同一本公共图书。
- 保证阅读进度、书签等功能更清晰。

## 7. 公共图书馆存储策略

公共图书馆必须使用引用式添加，不复制文件。

正确模型：

```text
books
  id = 100
  visibility = public
  file_path = /data/books/public/100.epub

bookshelves
  user_id = 1
  book_id = 100
  source_type = library

bookshelves
  user_id = 2
  book_id = 100
  source_type = library

bookshelves
  user_id = 3
  book_id = 100
  source_type = library
```

磁盘上只有一份文件：

```text
/data/books/public/100.epub
```

用户从书架移除公共图书时：

- 删除自己的 `bookshelves` 记录。
- 不删除 `books` 记录。
- 不删除真实文件。

管理员删除公共图书时：

- 优先软删除 `books` 记录。
- 标记所有相关书架引用不可用，或统一移除引用。
- 真实文件是否删除由管理员确认。

## 8. 数据库设计

### 8.1 users

```sql
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  email VARCHAR(255) UNIQUE,
  password_hash TEXT NOT NULL,
  nickname VARCHAR(128),
  avatar_path TEXT,
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ
);
```

### 8.2 refresh_tokens

```sql
CREATE TABLE refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 8.3 books

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
  visibility VARCHAR(32) NOT NULL,
  owner_user_id BIGINT NOT NULL REFERENCES users(id),
  library_status VARCHAR(32),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_books_visibility ON books(visibility);
CREATE INDEX idx_books_owner_user_id ON books(owner_user_id);
CREATE INDEX idx_books_file_hash ON books(file_hash);
CREATE INDEX idx_books_library_status ON books(library_status);
```

`visibility` 取值：

```text
private
public
```

`library_status` 取值：

```text
pending
approved
rejected
hidden
deleted
```

### 8.4 bookshelves

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
  UNIQUE(user_id, book_id)
);

CREATE INDEX idx_bookshelves_user_id ON bookshelves(user_id);
CREATE INDEX idx_bookshelves_book_id ON bookshelves(book_id);
CREATE INDEX idx_bookshelves_last_read_at ON bookshelves(last_read_at);
```

`source_type` 取值：

```text
uploaded
library
```

### 8.5 categories

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
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_scope ON categories(scope);
CREATE INDEX idx_categories_owner_user_id ON categories(owner_user_id);
```

`scope` 取值：

```text
system
user
```

### 8.6 tags

```sql
CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  scope VARCHAR(32) NOT NULL,
  owner_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tags_scope ON tags(scope);
CREATE INDEX idx_tags_owner_user_id ON tags(owner_user_id);
```

### 8.7 book_categories

```sql
CREATE TABLE book_categories (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, category_id)
);
```

### 8.8 book_tags

```sql
CREATE TABLE book_tags (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, tag_id)
);
```

### 8.9 bookshelf_tags

```sql
CREATE TABLE bookshelf_tags (
  bookshelf_id BIGINT NOT NULL REFERENCES bookshelves(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(bookshelf_id, tag_id)
);
```

### 8.10 reading_progress

```sql
CREATE TABLE reading_progress (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE CASCADE,
  format VARCHAR(32) NOT NULL,
  progress_type VARCHAR(32) NOT NULL,
  progress_value TEXT NOT NULL,
  percentage NUMERIC(6, 3),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, book_id)
);

CREATE INDEX idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX idx_reading_progress_book_id ON reading_progress(book_id);
```

`progress_type` 示例：

```text
epub_cfi
pdf_page
txt_offset
```

### 8.11 bookmarks

```sql
CREATE TABLE bookmarks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE CASCADE,
  title VARCHAR(255),
  position_type VARCHAR(32) NOT NULL,
  position_value TEXT NOT NULL,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 8.12 system_settings

```sql
CREATE TABLE system_settings (
  key VARCHAR(128) PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 9. 文件存储设计

应用容器挂载：

```text
/data
```

目录结构：

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
```

规则：

- 不使用用户上传的原始文件名作为实际文件名。
- 原始文件名只保存到数据库。
- 实际文件名使用 `book_id` 或 `sha256`。
- 所有文件访问都必须经过后端鉴权。
- 不直接暴露 `/data` 目录。

上传流程：

```text
前端选择文件
  |
前端校验格式和大小
  |
后端接收 multipart stream
  |
写入 /data/temp
  |
计算 sha256
  |
校验格式和大小
  |
写入 books 记录
  |
移动到正式目录
  |
创建书架记录或公共图书记录
```

## 10. 支持的图书格式

第一期支持：

```text
epub
pdf
txt
```

暂不建议第一期支持：

```text
mobi
azw3
cbz
docx
```

原因：

- MOBI / AZW3 浏览器端阅读支持复杂。
- 转码会增加 CPU、内存、依赖复杂度。
- CBZ、漫画阅读可以作为后续单独模块。

## 11. 阅读器设计

### 11.0 书籍解析职责划分

长期架构中，书籍解析能力主要放在后端，前端只负责阅读器 UI、交互和渲染。

后端负责：

- 文件鉴权和文件流读取。
- 元数据抽取，例如书名、作者、简介、封面。
- 章节列表生成。
- 章节正文读取。
- 图片和内嵌资源读取。
- TXT 编码识别、目录规则匹配、字节偏移分章。
- EPUB 目录解析、fragment 截取、跨 XHTML 拼接、图片资源解析。
- 后续 MOBI / AZW3 / UMD 等格式解析。

前端负责：

- 书架、图书馆、目录、阅读器页面展示。
- 字号、主题、行距、翻页或滚动等阅读交互。
- 请求章节列表、章节正文和资源图片。
- 节流上报阅读进度。

第一期可以采用渐进式实现：

- TXT 优先由后端解析和分块读取。
- EPUB 可以先用 `epub.js` 在前端阅读，后续迁移为后端解析章节和正文。
- PDF 可以继续用 `pdf.js` 在前端渲染，后端提供鉴权后的 Range 文件流。
- 若后续按 Legado 的图片式 PDF 方案实现，则 PDF 分页和页面图片渲染迁入后端。

后端解析器建议抽象为统一接口：

```text
ReadMetadata(file) -> BookMetadata
ReadChapters(file) -> []Chapter
ReadContent(file, chapter) -> Content
ReadResource(file, href) -> stream
```

### 11.1 EPUB

- 第一期可以使用 `epub.js`。
- 长期建议迁移到后端解析 EPUB 元数据、目录、章节正文和图片资源。
- 支持目录。
- 支持章节跳转。
- 支持主题。
- 支持字号和行距。
- 进度使用 EPUB CFI 保存。

### 11.2 PDF

- 使用 `pdf.js`。
- 后端文件接口必须支持 Range 请求。
- 进度保存页码。
- 支持缩放。
- 不做复杂重排。
- 如果后续采用 Legado 式 PDF 阅读，则后端每若干页生成一个章节，并把页面渲染为图片资源供前端展示。

### 11.3 TXT

- 后端负责 TXT 编码识别、目录规则选择、按字节偏移分章和分块读取。
- 避免一次性加载超大 TXT。
- 进度保存字符偏移量。
- 前端支持字号、行距、主题。

## 12. API 规划

基础路径：

```text
/api/v1
```

### 12.1 认证接口

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

### 12.2 用户接口

```http
GET   /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password
```

### 12.3 管理员用户接口

```http
GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
```

### 12.4 个人书架接口

```http
GET    /api/v1/bookshelf
GET    /api/v1/bookshelf/:id
POST   /api/v1/bookshelf/upload
POST   /api/v1/bookshelf/from-library/:bookId
PATCH  /api/v1/bookshelf/:id
DELETE /api/v1/bookshelf/:id
```

### 12.5 公共图书馆接口

```http
GET    /api/v1/library/books
GET    /api/v1/library/books/:id
POST   /api/v1/library/books/upload
POST   /api/v1/library/books/:id/add-to-bookshelf
```

### 12.6 管理员图书馆接口

```http
GET    /api/v1/admin/library/books
PATCH  /api/v1/admin/library/books/:id/status
DELETE /api/v1/admin/library/books/:id
```

### 12.7 分类接口

```http
GET    /api/v1/categories
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

### 12.8 标签接口

```http
GET    /api/v1/tags
POST   /api/v1/tags
PATCH  /api/v1/tags/:id
DELETE /api/v1/tags/:id
```

### 12.9 阅读接口

```http
GET   /api/v1/reader/books/:bookId/meta
GET   /api/v1/reader/books/:bookId/chapters
GET   /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET   /api/v1/reader/books/:bookId/resources
GET   /api/v1/reader/books/:bookId/file
GET   /api/v1/reader/books/:bookId/text
GET   /api/v1/reader/books/:bookId/progress
PUT   /api/v1/reader/books/:bookId/progress
```

说明：

- `chapters`、`content`、`resources` 用于后端解析模式。
- `file` 用于 EPUB/PDF 等前端直接解析或渲染模式，必须鉴权并支持 Range。
- `text` 用于 TXT 后端分块读取兼容路径，后续可并入章节正文接口。

### 12.10 书签接口

```http
GET    /api/v1/books/:bookId/bookmarks
POST   /api/v1/books/:bookId/bookmarks
PATCH  /api/v1/bookmarks/:id
DELETE /api/v1/bookmarks/:id
```

### 12.11 系统接口

```http
GET /api/v1/system/info
GET /api/v1/admin/system/storage
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
```

## 13. 权限设计

匿名用户：

- 访问登录页。
- 访问注册页。
- 注册账号，前提是系统允许注册。

普通用户：

- 管理自己的书架。
- 上传私有图书。
- 上传公共图书。
- 浏览公共图书馆。
- 添加公共图书到自己的书架。
- 阅读自己的书架图书。
- 管理自己的阅读进度。
- 管理自己的书签。
- 管理自己的个人标签。

管理员：

- 拥有普通用户全部权限。
- 管理用户。
- 管理系统分类。
- 管理系统标签。
- 管理公共图书。
- 审核公共图书。
- 删除违规图书。
- 查看存储统计。
- 修改系统设置。

文件访问规则：

- 私有图书只有拥有书架记录的用户可以读取。
- 公共图书只有登录用户可以读取。
- 管理员可以读取和管理所有图书。
- 图书文件必须通过 API 读取，不暴露真实路径。

## 14. 前端页面规划

路由：

```text
/login
/register
/bookshelf
/bookshelf/:id
/library
/library/:id
/reader/:bookId
/upload
/settings
/admin
/admin/users
/admin/books
/admin/library
/admin/categories
/admin/tags
/admin/settings
```

### 14.1 登录注册页

- 登录表单。
- 注册表单。
- 第一个用户注册后提示已成为管理员。
- 注册关闭时展示提示。

### 14.2 我的书架

- 卡片视图。
- 列表视图。
- 搜索。
- 分类筛选。
- 标签筛选。
- 最近阅读排序。
- 上传入口。
- 继续阅读入口。
- 移除图书。

### 14.3 公共图书馆

- 公共图书列表。
- 搜索。
- 分类筛选。
- 标签筛选。
- 上传公共图书。
- 加入书架。
- 已加入状态提示。

### 14.4 阅读器

- 返回书架。
- 目录。
- 阅读主体。
- 阅读设置。
- 进度保存。
- 继续阅读。
- 书签。

### 14.5 管理后台

- 用户管理。
- 图书管理。
- 公共图书审核。
- 分类管理。
- 标签管理。
- 系统设置。
- 存储统计。

## 15. Docker 部署规划

### 15.1 Dockerfile

使用多阶段构建：

```dockerfile
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/dist ./internal/app/static
RUN go build -o /app/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser
COPY --from=backend-builder /app/server /app/server
RUN mkdir -p /data && chown -R appuser:appuser /data /app
USER appuser
EXPOSE 8080
ENV DATA_DIR=/data
CMD ["/app/server"]
```

### 15.2 docker-compose.yml

```yaml
services:
  app:
    image: book-reader:latest
    build:
      context: ..
      dockerfile: deploy/docker/Dockerfile
    container_name: book-reader-app
    ports:
      - "8080:8080"
    volumes:
      - book_reader_files:/data
    environment:
      - APP_ENV=production
      - HTTP_ADDR=:8080
      - DATA_DIR=/data
      - DATABASE_DSN=postgres://book_reader:book_reader_password@db:5432/book_reader?sslmode=disable
      - JWT_SECRET=change-me
      - ACCESS_TOKEN_TTL=2h
      - REFRESH_TOKEN_TTL=720h
      - MAX_UPLOAD_SIZE_MB=100
      - ALLOW_REGISTRATION=true
      - LIBRARY_REVIEW_REQUIRED=false
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    container_name: book-reader-postgres
    environment:
      - POSTGRES_DB=book_reader
      - POSTGRES_USER=book_reader
      - POSTGRES_PASSWORD=book_reader_password
    volumes:
      - book_reader_postgres:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  book_reader_files:
  book_reader_postgres:
```

说明：

- `app` 容器包含 Go 后端和前端静态资源。
- `db` 容器只运行 PostgreSQL。
- 图书文件在 `book_reader_files` volume。
- 数据库文件在 `book_reader_postgres` volume。

## 16. 配置项规划

```text
APP_ENV=production
HTTP_ADDR=:8080
DATA_DIR=/data

DATABASE_DSN=postgres://book_reader:password@db:5432/book_reader?sslmode=disable

JWT_SECRET=change-me
ACCESS_TOKEN_TTL=2h
REFRESH_TOKEN_TTL=720h

MAX_UPLOAD_SIZE_MB=100
ALLOW_REGISTRATION=true
LIBRARY_REVIEW_REQUIRED=false

LOG_LEVEL=info
```

配置说明：

- `JWT_SECRET` 生产环境必须修改。
- `MAX_UPLOAD_SIZE_MB` 控制单文件最大上传大小。
- `ALLOW_REGISTRATION` 控制是否允许新用户注册。
- `LIBRARY_REVIEW_REQUIRED` 控制公共图书是否需要管理员审核。

## 17. 资源占用控制

不引入以下组件：

- Redis。
- Elasticsearch。
- RabbitMQ。
- Kafka。
- MinIO。
- Java Spring Boot。
- 额外全文搜索服务。

控制策略：

- 后端使用 Go。
- 数据库使用 PostgreSQL Alpine 镜像。
- 文件上传使用流式写入。
- 文件读取使用流式响应。
- PDF 文件接口支持 Range 请求。
- TXT 文件支持分块读取。
- 上传大小可配置。
- 阅读进度保存做前端节流。
- 不在后端执行重型图书转码。
- 封面提取第一期可以简化，允许默认封面或手动上传。

## 18. 安全设计

### 18.1 密码安全

- 使用 bcrypt 或 argon2id。
- 不保存明文密码。
- 不在日志中记录密码。

### 18.2 Token 安全

- Access Token 使用 JWT。
- Refresh Token 存储 hash。
- 用户禁用后不允许刷新 Token。
- 生产环境必须设置强随机 `JWT_SECRET`。

### 18.3 上传安全

- 限制扩展名。
- 限制文件大小。
- 文件名不可信。
- 禁止路径穿越。
- 上传先进入临时目录。
- 校验通过后移动到正式目录。

### 18.4 文件访问安全

- 不暴露真实文件路径。
- 所有图书文件访问必须鉴权。
- 后端根据用户和书架关系判断访问权限。
- 私有图书只能被拥有者访问。
- 公共图书只允许登录用户访问。

## 19. 开发阶段规划

### 阶段一：项目骨架

目标：

- 建立 Go 后端项目。
- 建立 Vue 3 前端项目。
- 接入 PostgreSQL。
- 建立 Docker compose。
- 后端托管前端静态资源。
- 提供健康检查接口。

交付：

- `/api/v1/health`。
- Docker compose 可启动。
- Web 首页可访问。

### 阶段二：用户与认证

目标：

- 注册。
- 登录。
- JWT 鉴权。
- Refresh Token。
- 第一个注册用户为管理员。
- 管理员查看用户列表。

交付：

- 登录页面。
- 注册页面。
- 用户状态管理。
- 后端鉴权中间件。

### 阶段三：个人书架与上传

目标：

- 私有图书上传。
- 文件落盘。
- 创建图书记录。
- 创建书架记录。
- 我的书架列表。
- 删除或移除图书。

交付：

- 上传页面。
- 我的书架页面。
- 文件权限访问接口。

### 阶段四：在线阅读器

目标：

- EPUB 阅读。
- PDF 阅读。
- TXT 阅读。
- 后端解析器基础抽象。
- TXT 后端编码识别、目录解析和按章节读取。
- 阅读进度保存。
- 继续阅读。

交付：

- 阅读器页面。
- 阅读进度接口。
- 章节列表接口。
- 章节正文接口。
- Range 文件访问支持。

### 阶段五：公共图书馆

目标：

- 上传公共图书。
- 浏览公共图书。
- 添加公共图书到个人书架。
- 引用式添加，不复制文件。
- 公共图书状态管理。

交付：

- 图书馆页面。
- 加入书架功能。
- 管理员审核接口。

### 阶段六：分类标签

目标：

- 系统分类。
- 系统标签。
- 用户个人标签。
- 图书分类筛选。
- 图书标签筛选。

交付：

- 分类管理页面。
- 标签管理页面。
- 书架筛选。
- 图书馆筛选。

### 阶段七：管理后台

目标：

- 用户管理。
- 图书管理。
- 公共图书审核。
- 系统设置。
- 存储统计。

交付：

- 管理后台。
- 管理 API。

### 阶段八：稳定性和体验优化

目标：

- 上传进度。
- 错误提示。
- 阅读设置持久化。
- 书签。
- 备份恢复文档。
- 日志完善。

交付：

- 稳定可用版本。
- 完整部署文档。

## 20. MVP 范围

首版 MVP 必须包含：

- 用户注册。
- 用户登录。
- 第一个注册用户为管理员。
- PostgreSQL 数据库。
- 私有图书上传。
- 我的书架。
- EPUB / PDF / TXT 阅读。
- 阅读进度保存。
- 公共图书馆上传。
- 公共图书加入书架，不复制文件。
- 基础分类。
- 基础标签。
- Docker compose 部署。

首版 MVP 暂不包含：

- 图书全文搜索。
- OCR。
- 图书转码。
- 第三方登录。
- 邮箱验证。
- WebDAV。
- OPDS。
- 自动 ISBN 元数据抓取。
- 复杂笔记系统。
- 对象存储。

## 21. 测试规划

后端测试：

- 第一个注册用户成为管理员。
- 后续注册用户为普通用户。
- 登录成功。
- 登录失败。
- 私有图书权限校验。
- 公共图书添加书架不复制文件。
- 删除公共图书书架引用不删除文件。
- 管理员审核公共图书。
- 阅读进度保存。

前端测试：

- 登录注册流程。
- 书架列表展示。
- 上传图书。
- 图书馆加入书架。
- 阅读器打开 EPUB。
- 阅读器打开 PDF。
- 阅读器打开 TXT。
- 管理后台权限。

集成测试：

- Docker compose 启动。
- 首次注册管理员。
- 上传一本私有 EPUB。
- 上传一本公共 EPUB。
- 第二个用户添加公共 EPUB 到书架。
- 检查公共图书文件只有一份。

## 22. 风险与处理方案

### 22.1 公共图书删除影响用户书架

风险：

- 管理员删除公共图书后，用户书架中的引用失效。

方案：

- 默认软删除公共图书。
- 用户书架显示图书已下架。
- 管理员可以选择强制删除并移除所有引用。

### 22.2 大文件导致内存压力

风险：

- 上传或阅读大文件时占用过多内存。

方案：

- 上传流式写入。
- 文件流式响应。
- PDF 支持 Range。
- TXT 分块读取。
- 限制单文件大小。

### 22.3 第一个用户管理员并发问题

风险：

- 两个用户同时注册时可能都成为管理员。

方案：

- 注册逻辑放入数据库事务。
- 对首个用户判断使用事务锁或数据库约束辅助。

### 22.4 数据库和文件不一致

风险：

- 文件保存成功但数据库失败。
- 数据库成功但文件移动失败。

方案：

- 先写临时文件。
- 数据库事务创建记录。
- 文件移动失败则回滚。
- 定期清理 temp 目录。
- 后续提供孤儿文件扫描工具。

### 22.5 PostgreSQL 运维复杂度

风险：

- 相比 SQLite 多一个数据库容器。

方案：

- 使用 docker-compose 默认配置。
- 提供备份恢复文档。
- 默认使用轻量 `postgres:16-alpine`。

## 23. 备份恢复规划

需要备份：

- PostgreSQL 数据库。
- `/data/books` 图书文件。
- `/data/covers` 封面文件。

推荐备份命令：

```bash
docker exec book-reader-postgres pg_dump -U book_reader book_reader > backup.sql
docker run --rm -v book_reader_files:/data -v $(pwd):/backup alpine tar czf /backup/files.tar.gz /data
```

恢复时：

```bash
docker exec -i book-reader-postgres psql -U book_reader book_reader < backup.sql
docker run --rm -v book_reader_files:/data -v $(pwd):/backup alpine tar xzf /backup/files.tar.gz -C /
```

## 24. 最终架构结论

推荐最终架构：

```text
Go + Gin + GORM + PostgreSQL + Vue 3 + Vite + Docker Compose
```

部署形态：

```text
app 容器：后端 API + 前端静态资源 + 文件服务
db 容器：PostgreSQL
files volume：图书与封面文件
postgres volume：数据库数据
```

最关键的数据设计：

```text
books 表保存真实图书文件信息
bookshelves 表保存用户书架关系
公共图书加入个人书架时只创建 bookshelves 记录
不复制公共图书文件
```

该方案满足：

- 前后端统一打包部署。
- 使用 PostgreSQL。
- 多用户注册。
- 第一个注册为管理员。
- 私有图书上传。
- 公共图书馆。
- 公共图书引用式加入书架。
- 分类标签。
- 在线阅读。
- Docker 部署。
- 控制内存和存储压力。
