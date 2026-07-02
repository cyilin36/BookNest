# 阅读器 API 契约文档

本文档是前后端独立开发时共同遵守的接口契约。若本文档与 `reader-backend-development-doc.md` 或 `reader-frontend-development-doc.md` 冲突，先更新本文档并同步更新对应开发文档，再编码。

## 1. 基础约定

API 基础路径固定为：

```text
/api/v1
```

所有 JSON 字段使用 `snake_case`。时间字段使用 ISO 8601 字符串。后端 ID 使用 number。可空字段必须显式返回 `null`，不要省略关键字段。

除文件流接口外，所有接口返回 JSON。成功响应统一为：

```json
{
  "data": {},
  "request_id": "req_xxx"
}
```

分页响应统一为：

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

错误响应统一为：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "请求参数不合法"
  },
  "request_id": "req_xxx"
}
```

## 2. 认证约定

Access Token 由前端保存并放入请求头：

```http
Authorization: Bearer <access_token>
```

Refresh Token 默认由后端写入 HttpOnly Cookie：

```text
refresh_token=<token>; HttpOnly; SameSite=Lax; Path=/api/v1/auth
```

生产 HTTPS 环境必须增加 `Secure`。`POST /api/v1/auth/refresh` 和 `POST /api/v1/auth/logout` 默认从 Cookie 读取 refresh token。开发和 API 调试可兼容 JSON body 传递 `refresh_token`。

401 处理流程：前端收到 401 后只发起一个 refresh 请求；refresh 成功后重放等待队列；refresh 失败则清空认证状态并跳转登录页。

## 3. 枚举

```ts
export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'disabled'
export type BookFormat = 'epub' | 'pdf' | 'txt'
export type BookVisibility = 'private' | 'public'
export type LibraryStatus = 'approved' | 'hidden' | 'deleted'
export type BookshelfSourceType = 'uploaded' | 'library'
export type BookshelfStatus = 'active' | 'removed' | 'unavailable'
export type ProgressType = 'epub_cfi' | 'pdf_page' | 'txt_offset'
export type ParseStatus = 'parsed' | 'partial' | 'failed'
```

## 4. 通用类型

### User

```ts
export interface User {
  id: number
  username: string
  email: string | null
  nickname: string | null
  avatar_url: string | null
  role: UserRole
  status: UserStatus
  storage_quota_bytes: number | null
  storage_used_bytes: number
  created_at?: string
  last_login_at?: string | null
}
```

### AuthSession

```ts
export interface AuthSession {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
  user: User
}
```

### BookMeta

```ts
export interface BookMeta {
  id: number
  title: string
  author: string | null
  description: string | null
  format: BookFormat
  cover_url: string | null
  file_size: number
  visibility: BookVisibility
  library_status: LibraryStatus | null
  parse_status?: ParseStatus
  created_at?: string
  updated_at?: string
}
```

### BookshelfItem

```ts
export interface BookshelfItem {
  id: number
  book_id: number
  title: string
  author: string | null
  format: BookFormat
  cover_url: string | null
  source_type: BookshelfSourceType
  visibility: BookVisibility
  library_status: LibraryStatus | null
  favorite: boolean
  pinned: boolean
  last_read_at: string | null
  added_at: string
  readable: boolean
  unreadable_reason: 'library_hidden' | 'library_deleted' | 'file_missing' | 'permission_denied' | null
  progress_percentage: number | null
}
```

### LibraryBook

```ts
export interface LibraryBook {
  id: number
  title: string
  author: string | null
  description: string | null
  format: BookFormat
  cover_url: string | null
  file_size: number
  library_status: LibraryStatus
  owner_user_id: number
  owner_username?: string
  in_bookshelf: boolean
  bookshelf_id: number | null
  created_at: string
  updated_at?: string
}
```

### Category 和 Tag

```ts
export interface Category {
  id: number
  name: string
  description: string | null
  scope: 'system'
  created_at?: string
}

export interface Tag {
  id: number
  name: string
  description: string | null
  scope: 'system'
  created_at?: string
}
```

### ReaderChapter

```ts
export interface ReaderChapter {
  id: number
  chapter_index: number
  title: string
  is_volume: boolean
  word_count: number | null
}
```

### ReaderChapterContent

```ts
export interface ReaderChapterContent {
  id: number
  chapter_index: number
  title: string
  content_type: 'html' | 'text'
  content: string
}
```

### ReadingProgress

```ts
export interface ReadingProgress {
  progress_type: ProgressType
  progress_value: string
  percentage: number | null
  updated_at: string
}
```

前端章节阅读器使用现有 `progress_value` 字符串保存恢复位置：

- 兼容旧值：纯数字字符串按章节 ID 解释，只恢复到该章节顶部。
- 新值：JSON 字符串，格式为 `{"chapterId":123,"scrollRatio":0.42}`，其中 `chapterId` 是章节 ID，`scrollRatio` 是当前页面滚动比例，范围 0 到 1。
- 后端只需按字符串持久化 `progress_value`，不解析该 JSON。

## 5. 通用查询参数

分页接口统一支持：

```text
page=1
page_size=20
keyword=xxx
sort=created_at
order=desc
```

规则：

- `page` 最小为 1。
- `page_size` 默认 20，最大 100。
- `order` 只允许 `asc`、`desc`。
- `sort` 必须按接口白名单校验。

## 6. 公开接口

### 健康检查

```http
GET /api/v1/health
```

权限：匿名。

响应：

```json
{
  "data": {
    "status": "ok"
  }
}
```

### 系统信息

```http
GET /api/v1/system/info
```

权限：匿名。

响应：

```ts
interface SystemInfo {
  site_name: string
  site_icon_url: string | null
  login_background_url: string | null
  allow_registration: boolean
  library_review_required: boolean
  supported_formats: BookFormat[]
  max_upload_size_mb: number
  default_user_storage_quota_mb: number
}
```

`site_icon_url` 为站点图标地址，未配置时为 `null`。`login_background_url` 为登录页背景图地址，未配置时为 `null`。两者均为不透明的后端文件流地址，前端直接当作图片 URL 使用。

### 站点图标文件流

```http
GET /api/v1/system/icon
HEAD /api/v1/system/icon
```

权限：匿名。

响应：站点图标原始字节和正确 `Content-Type`。未配置图标时返回 `404 not_found`。

### 登录页背景文件流

```http
GET /api/v1/system/login-background
HEAD /api/v1/system/login-background
```

权限：匿名。

响应：登录页背景图片原始字节和正确 `Content-Type`。未配置背景时返回 `404 not_found`。

规则：

- 该接口为匿名访问，便于登录页在未认证状态下直接展示背景。
- 仅返回当前配置的单张背景图，由管理员通过后台上传或删除。

## 7. 认证接口

### 注册

```http
POST /api/v1/auth/register
```

权限：匿名。

请求：

```ts
interface RegisterRequest {
  username: string
  email?: string | null
  password: string
  nickname?: string | null
}
```

响应：`AuthSession`。首个注册用户必须返回 `user.role='admin'`。

常见错误码：`registration_disabled`、`username_exists`、`email_exists`、`validation_failed`。

### 登录

```http
POST /api/v1/auth/login
```

权限：匿名。

请求：

```ts
interface LoginRequest {
  login: string
  password: string
}
```

响应：`AuthSession`，并设置 `refresh_token` HttpOnly Cookie。

常见错误码：`invalid_credentials`、`user_disabled`。

### 刷新 Token

```http
POST /api/v1/auth/refresh
```

权限：Refresh Token Cookie；开发调试可用 body。

调试请求：

```ts
interface RefreshRequest {
  refresh_token?: string
}
```

响应：`AuthSession`，并轮换 `refresh_token` Cookie。

常见错误码：`refresh_token_invalid`、`user_disabled`。

### 退出

```http
POST /api/v1/auth/logout
```

权限：登录或 Refresh Token Cookie。

响应：`{}`，并清除 `refresh_token` Cookie。

### 当前认证用户

```http
GET /api/v1/auth/me
```

权限：登录。

响应：`User`。

## 8. 当前用户接口

### 当前用户信息

```http
GET /api/v1/users/me
```

权限：登录。

响应：`User`，必须包含 `storage_quota_bytes` 和 `storage_used_bytes`。

### 修改当前用户信息

```http
PATCH /api/v1/users/me
```

权限：登录。

请求：

```ts
interface UpdateMeRequest {
  email?: string | null
  nickname?: string | null
}
```

响应：`User`。

### 修改密码

```http
PATCH /api/v1/users/me/password
```

权限：登录。

请求：

```ts
interface ChangePasswordRequest {
  old_password: string
  new_password: string
}
```

响应：`{}`。成功后撤销该用户全部 refresh token。

### 上传自定义头像

```http
POST /api/v1/users/me/avatar/upload
```

权限：登录。

请求类型：`multipart/form-data`。

字段：

```text
file=<image file>
```

响应：`User`。

规则：

- 支持格式：PNG、JPEG、WebP、GIF。
- 文件大小限制：5 MB。
- 上传成功后会替换原有头像。
- 旧的自定义头像文件会被自动删除（默认头像不会删除）。

常见错误码：`avatar_format_not_supported`、`avatar_content_invalid`、`avatar_too_large`、`payload_too_large`。

### 设置默认头像

```http
POST /api/v1/users/me/avatar/default
```

权限：登录。

请求：

```ts
interface SetDefaultAvatarRequest {
  avatar_name: string
}
```

响应：`User`。

规则：

- `avatar_name` 必须是 `default1`、`default2`、`default3`、`default4`、`default5`、`default6` 之一。
- 设置默认头像后，旧的自定义头像文件会被自动删除。

常见错误码：`invalid_avatar_name`。

### 获取用户头像

```http
GET /api/v1/users/:userId/avatar
HEAD /api/v1/users/:userId/avatar
```

权限：登录。

响应：头像图片原始字节和正确 `Content-Type`。未设置头像时返回 `404`。

规则：

- 可以获取任何用户的头像（包括自己和其他用户）。
- 默认头像返回 SVG 格式。
- 自定义头像返回上传时的原始格式。
- 响应头包含 `Cache-Control: public, max-age=3600`。

## 9. 个人书架接口

### 上传私有图书

```http
POST /api/v1/bookshelf/upload
```

权限：登录。

请求类型：`multipart/form-data`。

字段：

```text
file=<book file>
title=<optional>
author=<optional>
description=<optional>
category_ids=<optional comma separated>
tag_ids=<optional comma separated>
```

响应：`BookshelfItem`。

常见错误码：`book_format_not_supported`、`payload_too_large`、`storage_quota_exceeded`、`validation_failed`。

### 我的书架列表

```http
GET /api/v1/bookshelf
```

权限：登录。

查询参数：

```text
keyword
format=epub|pdf|txt
category_id
tag_id
source_type=uploaded|library
favorite=true|false
page
page_size
sort=last_read_at|added_at|title
order=asc|desc
```

响应：分页 `BookshelfItem[]`。

默认排序：`pinned desc, last_read_at desc nulls last, added_at desc`。

### 书架详情

```http
GET /api/v1/bookshelf/:id
```

权限：登录，只能访问自己的书架项。

响应：`BookshelfItem`。

### 下载书架图书

```http
GET /api/v1/bookshelf/:id/download
```

权限：登录，只能下载自己的 active 书架项。

响应：图书源文件字节。

规则：

- 私有上传图书：当前用户必须拥有该 active 书架项。
- 公共图书馆引用：当前用户必须拥有该 active 书架项，且公共图书仍为 `library_status='approved'`、`deleted_at IS NULL`。
- 下载内容是原始图书文件，不重新转码或修改文件内容。
- 响应头使用 `Content-Disposition: attachment` 触发下载。
- 下载文件名使用书架上展示的标题；若书架项设置了 `personal_title`，优先使用 `personal_title`，否则使用图书标题。
- 文件扩展名使用图书 `format`，例如 `.epub`、`.pdf`、`.txt`。
- 文件名中的路径分隔符、控制字符和常见非法文件名字符会被替换或移除。

### 编辑书架项

```http
PATCH /api/v1/bookshelf/:id
```

权限：登录，只能编辑自己的书架项。

请求：

```ts
interface UpdateBookshelfItemRequest {
  personal_title?: string | null
  personal_category_id?: number | null
  favorite?: boolean
  pinned?: boolean
  tag_ids?: number[]
}
```

响应：`BookshelfItem`。

### 移除书架项

```http
DELETE /api/v1/bookshelf/:id
```

权限：登录，只能移除自己的书架项。

响应：`{}`。

规则：

- 移除公共图书馆引用时，只删除当前用户书架引用，不删除公共图书文件和公共图书记录。
- 移除自己上传的私有图书时，后端先写入 `books.deleted_at` 作为删除标记，再物理删除图书文件和封面文件，最后删除该书的书架项、书架标签、阅读进度、书签、章节和 `books` 图书记录。
- 私有图书删除成功后，数据库不再保留该图书条目和相关个人数据，效果等同从未添加过该图书。
- 私有图书物理文件删除失败时，后端会撤销 `books.deleted_at` 删除标记，并返回 `book_file_delete_failed`；前端应把错误消息提示给用户。

### 从公共图书加入书架

```http
POST /api/v1/bookshelf/from-library/:bookId
POST /api/v1/library/books/:id/add-to-bookshelf
```

权限：登录。

响应：`BookshelfItem`。

常见错误码：`book_already_in_bookshelf`、`library_book_not_approved`、`book_not_found`。

## 10. 公共图书馆接口

### 公共图书列表

```http
GET /api/v1/library/books
```

权限：登录。

查询参数：

```text
keyword
format=epub|pdf|txt
category_id
tag_id
mine=true|false
status=approved|hidden
page
page_size
sort=created_at|title
order=asc|desc
```

响应：分页 `LibraryBook[]`。

普通查询默认只返回 `approved` 公共图书。`mine=true` 可查看当前用户自己上传的公共图书，并可用 `status=approved|hidden` 过滤。

### 公共图书详情

```http
GET /api/v1/library/books/:id
```

权限：登录。

响应：`LibraryBook`。

规则：普通用户可查看 `approved` 图书；上传者可查看自己上传的 `approved` 或 `hidden` 图书。

### 下载公共图书馆图书

```http
GET /api/v1/library/books/:id/download
```

权限：登录。

响应：图书源文件字节。

规则：

- 只允许下载 `visibility='public'`、`library_status='approved'`、`deleted_at IS NULL` 的公共图书。
- `hidden`、`deleted` 或已被管理员物理删除的公共图书不可下载。
- 下载内容是原始图书文件，不重新转码或修改文件内容。
- 响应头使用 `Content-Disposition: attachment` 触发下载。
- 下载文件名使用公共图书馆展示标题，扩展名使用图书 `format`。
- 文件名中的路径分隔符、控制字符和常见非法文件名字符会被替换或移除。

### 上传公共图书

```http
POST /api/v1/library/books/upload
```

权限：登录。

请求类型：`multipart/form-data`。

字段同私有上传。

响应：`LibraryBook`。

规则：上传后返回 `library_status='approved'`；不自动加入上传者书架。

### 下架自己上传的公共图书

```http
POST /api/v1/library/books/:id/hide
```

权限：登录，且必须是该公共图书上传者。

响应：`LibraryBook`。

规则：

- 只允许下架自己上传的公共图书。
- `approved` 可以下架为 `hidden`。
- 已经是 `hidden` 时幂等返回当前图书。
- `deleted` 不允许下架。
- 只更新 `books.library_status='hidden'` 和 `books.updated_at`，不更新 `deleted_at`。
- 不删除图书文件、封面、书架引用、阅读进度或书签。
- 下架后公共图书馆普通列表不可见，不能被新增加入书架。
- 已加入他人书架的引用保留，书架仍可展示，但 `readable=false` 且 `unreadable_reason='library_hidden'`，不能继续阅读。

### 重新上架自己上传的公共图书

```http
POST /api/v1/library/books/:id/show
```

权限：登录，且必须是该公共图书上传者。

响应：`LibraryBook`。

规则：

- 只允许重新上架自己上传的公共图书。
- 图书必须是 `visibility='public'`、`owner_user_id=current_user_id`、`deleted_at IS NULL`。
- 只允许 `library_status='hidden'` 的图书重新上架。
- 更新 `books.library_status='approved'` 和 `books.updated_at`。
- 不删除或修改图书文件、封面、书架引用、阅读进度或书签。
- `approved`、`deleted` 或其他状态不允许通过该接口恢复。
- 管理员已物理删除的公共图书不会出现在 `mine=true` 列表，也无法重新上架，效果等同用户从未上传过该书。

## 11. 阅读接口

阅读接口统一权限：登录。私有图书要求当前用户存在书架记录；公共图书要求 `approved`；管理员可读取所有未物理丢失图书。

### 阅读元数据

```http
GET /api/v1/reader/books/:bookId/meta
```

响应：`BookMeta`。

### 图书文件流

```http
GET /api/v1/reader/books/:bookId/file
```

响应：文件原始字节。

规则：

- 必须校验阅读权限。
- 必须支持 Range。
- PDF Range 请求必须正确返回 `206 Partial Content`。
- 响应头必须包含 `Accept-Ranges: bytes`。
- 不暴露真实文件路径。

### 封面文件流

```http
GET /api/v1/reader/books/:bookId/cover
```

响应：封面原始字节和正确 `Content-Type`。无封面时返回 `404 not_found` 或 `book_file_missing`。

### TXT 分块

```http
GET /api/v1/reader/books/:bookId/text?offset=0&limit=65536
```

响应：

```ts
interface TxtChunkResponse {
  offset: number
  limit: number
  next_offset: number | null
  content: string
}
```

规则：只允许 TXT；`offset` 是字节偏移；`limit` 默认 `TXT_CHUNK_SIZE`，最大 1MB。前端优先使用章节接口，该接口作为兜底。

### 章节目录

```http
GET /api/v1/reader/books/:bookId/chapters
```

响应：`ReaderChapter[]`。

### 章节内容

```http
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
```

响应：`ReaderChapterContent`。

规则：

- TXT 返回 `content_type='text'`。
- TXT 正文保留换行；章节正文开头多余空行会被归一，原文首段已有缩进时保留原缩进，首段无缩进时后端可按阅读排版补全角缩进 `　　`。
- EPUB 返回后端净化后的 `html`。
- EPUB 章节 HTML 内图片地址应改写为 `/api/v1/reader/books/:bookId/resources?href=...`。
- EPUB 中的 SVG `<image href="...">`、`<image xlink:href="...">` 应由后端归一成普通 `<img src="...">`，便于前端直接渲染。
- 章节 HTML 中的资源 URL 可由后端追加 `uid`、`expires`、`sig` 签名参数；前端应把该 URL 当作不透明地址直接渲染，不需要自行构造签名。

### EPUB 内嵌资源

```http
GET /api/v1/reader/books/:bookId/resources?href=images/cover.jpg
```

章节内容接口返回的图片 URL 可能形如：

```http
GET /api/v1/reader/books/:bookId/resources?expires=1780333831&href=images/cover.jpg&sig=...&uid=2
```

响应：资源原始字节和正确 `Content-Type`。

规则：

- `href` 必填，必须 URL 编码。
- 后端必须规范化路径并防止路径穿越。
- 图片类型至少支持 JPEG、PNG、GIF、WEBP、SVG。
- 资源接口支持两种访问方式：
  - 常规 API 调用：携带 `Authorization: Bearer <access_token>`。
  - 章节 HTML 内图片请求：使用后端生成的 `uid`、`expires`、`sig` 签名参数。
- 签名参数由后端生成，绑定用户、图书、资源路径和过期时间；签名过期或无效时返回 401。
- 即使使用签名 URL，后端仍必须校验用户状态和该用户对图书的阅读权限。
- 可返回 `Cache-Control: private, max-age=3600`。

### 获取阅读进度

```http
GET /api/v1/reader/books/:bookId/progress
```

响应：`ReadingProgress | null`。

无进度时：

```json
{
  "data": null
}
```

### 保存阅读进度

```http
PUT /api/v1/reader/books/:bookId/progress
```

请求：

```ts
interface SaveReadingProgressRequest {
  progress_type: ProgressType
  progress_value: string
  percentage?: number | null
}
```

响应：`ReadingProgress`。

规则：

- EPUB 只允许 `epub_cfi`。
- PDF 只允许 `pdf_page`。
- TXT 只允许 `txt_offset`。
- `progress_value` 按字符串保存；章节阅读器当前写入 `{"chapterId":number,"scrollRatio":number}`，并兼容历史纯章节 ID 字符串。
- `percentage` 可空，非空范围 0 到 100。
- 保存成功后更新当前用户对应书架项的 `last_read_at`。

## 12. 书签接口

书签第一版可以不在前端启用；若启用，必须按本节契约实现。

```ts
export interface Bookmark {
  id: number
  book_id: number
  bookshelf_id: number | null
  title: string | null
  note: string | null
  progress_type: ProgressType
  progress_value: string
  percentage: number | null
  created_at: string
  updated_at: string
}
```

### 书签列表

```http
GET /api/v1/books/:bookId/bookmarks
```

权限：登录且有阅读权限。

响应：`Bookmark[]`。

### 创建书签

```http
POST /api/v1/books/:bookId/bookmarks
```

权限：登录且有阅读权限。

请求：

```ts
interface CreateBookmarkRequest {
  title?: string | null
  note?: string | null
  progress_type: ProgressType
  progress_value: string
  percentage?: number | null
}
```

响应：`Bookmark`。

### 修改书签

```http
PATCH /api/v1/bookmarks/:id
```

权限：登录，只能修改自己的书签。

请求：

```ts
interface UpdateBookmarkRequest {
  title?: string | null
  note?: string | null
}
```

响应：`Bookmark`。

### 删除书签

```http
DELETE /api/v1/bookmarks/:id
```

权限：登录，只能删除自己的书签。

响应：`{}`。

## 13. 分类和标签接口

### 分类列表

```http
GET /api/v1/categories?scope=system
```

权限：登录。

响应：`Category[]`。

### 标签列表

```http
GET /api/v1/tags?scope=system
```

权限：登录。

响应：`Tag[]`。

### 管理员分类管理

```http
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

权限：管理员。

创建/修改请求：

```ts
interface UpsertCategoryRequest {
  name: string
  description?: string | null
}
```

创建/修改响应：`Category`。删除响应：`{}`。

### 管理员标签管理

```http
POST   /api/v1/admin/tags
PATCH  /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

权限：管理员。

创建/修改请求：

```ts
interface UpsertTagRequest {
  name: string
  description?: string | null
}
```

创建/修改响应：`Tag`。删除响应：`{}`。

## 14. 管理员接口

### 用户列表

```http
GET /api/v1/admin/users
```

权限：管理员。

查询参数：`keyword`、`status`、`role`、`page`、`page_size`、`sort=created_at|last_login_at|username`、`order`。

响应：分页 `User[]`。

### 用户详情

```http
GET /api/v1/admin/users/:id
```

权限：管理员。

响应：`User`。

### 修改用户基础信息

```http
PATCH /api/v1/admin/users/:id
```

权限：管理员。

请求：

```ts
interface AdminUpdateUserRequest {
  email?: string | null
  nickname?: string | null
  storage_quota_bytes?: number | null
}
```

响应：`User`。

### 修改用户状态

```http
PATCH /api/v1/admin/users/:id/status
```

权限：管理员。

请求：

```ts
interface AdminUpdateUserStatusRequest {
  status: UserStatus
}
```

响应：`User`。管理员不能禁用自己；禁用用户时撤销该用户全部 refresh token。

### 修改用户角色

```http
PATCH /api/v1/admin/users/:id/role
```

权限：管理员。

请求：

```ts
interface AdminUpdateUserRoleRequest {
  role: UserRole
}
```

响应：`User`。第一期禁止管理员修改自己的角色。

### 删除用户

```http
DELETE /api/v1/admin/users/:id
```

权限：管理员。

响应：`{}`。

规则：

- 管理员不能删除自己。
- 删除用户会物理清除该用户拥有的所有图书文件和封面文件。
- 删除用户会清除该用户账号、refresh token、书架、阅读进度、书签等所有个人数据。
- 删除用户拥有的公共图书时，同时清除其他用户引用这些图书产生的书架项、阅读进度和书签。

### 管理员公共图书列表

```http
GET /api/v1/admin/library/books
```

权限：管理员。

查询参数：`keyword`、`format`、`status`、`page`、`page_size`、`sort=created_at|title`、`order`。

响应：分页 `LibraryBook[]`。

### 修改公共图书状态

```http
PATCH /api/v1/admin/library/books/:id/status
```

权限：管理员。

请求：

```ts
interface UpdateLibraryBookStatusRequest {
  status: 'approved' | 'hidden'
  reason?: string | null
}
```

响应：`LibraryBook`。

允许状态流转：

```text
approved -> hidden
hidden -> approved
```

### 删除公共图书

```http
DELETE /api/v1/admin/library/books/:id
```

权限：管理员。

响应：`{}`。

规则：

- 管理员删除公共图书是真正删除。
- 执行时先将图书标记为 `library_status='deleted'` 并写入 `deleted_at`。
- 标记成功后物理删除图书文件和封面文件。
- 最后删除数据库中的图书记录、章节、图书分类/标签关联、书架引用、阅读进度和书签。
- 删除完成后该图书不会再出现在公共图书馆或任何用户书架中。

### 存储统计

```http
GET /api/v1/admin/system/storage
```

权限：管理员。

响应：

```ts
interface StorageStats {
  private_books_bytes: number
  public_books_bytes: number
  covers_bytes: number
  total_books: number
  private_books: number
  public_books: number
}
```

### 系统设置读取

```http
GET /api/v1/admin/system/settings
```

权限：管理员。

响应：

```ts
interface SystemSettings {
  site_name: string
  site_icon_url: string | null
  login_background_url: string | null
  allow_registration: boolean
  library_review_required: boolean
  max_upload_size_mb: number
  default_user_storage_quota_mb: number
}
```

### 系统设置修改

```http
PUT /api/v1/admin/system/settings
```

权限：管理员。

请求：`SystemSettings`。

响应：`SystemSettings`。

规则：

- `max_upload_size_mb` 不能超过启动时 `REQUEST_BODY_LIMIT_MB`。
- `library_review_required` 保留用于兼容旧配置，当前不影响公共图书上传状态。
- `site_icon_url` 和 `login_background_url` 为只读字段，通过专用接口上传和删除。

### 上传登录页背景

```http
POST /api/v1/admin/system/login-background
```

权限：管理员。

请求：`multipart/form-data`。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| file | file | 是 | 背景图片文件 |

响应：

```ts
interface LoginBackgroundUploadResponse {
  login_background_url: string
  content_type: string
}
```

规则：

- 支持格式：PNG、JPEG、WebP、GIF。
- 文件大小限制：10 MB。
- 上传成功后会替换原有的登录页背景。
- 旧背景文件会被自动删除。

### 删除登录页背景

```http
DELETE /api/v1/admin/system/login-background
```

权限：管理员。

响应：

```json
{
  "data": {
    "login_background_url": null
  },
  "request_id": "req_xxx"
}
```

规则：

- 删除后登录页将不显示背景图片（或使用前端默认样式）。
- 删除操作同时删除服务器上的图片文件。

## 15. 错误码

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
library_review_required
category_not_found
tag_not_found
storage_quota_exceeded
```

## 16. 部署契约

本地开发：

- 前端 `VITE_API_BASE=/api/v1`。
- Vite proxy 将 `/api` 转发到 `http://localhost:8080`。
- Axios 必须设置 `withCredentials=true`。

统一部署：

- 前端执行 build 后由 Go 后端托管静态资源。
- 浏览器访问后端地址，例如 `http://server:8080`。
- API 和前端同源。
- API 路由必须优先注册。
- 非 `/api/*` 路径返回前端 `index.html`，支持 Vue Router history 模式。
- `/api/*` 不允许回退到前端页面，必须返回 JSON 404。

## 17. 联调验收顺序

前后端独立开发后，按以下顺序合并联调：

1. `GET /api/v1/health` 和 `GET /api/v1/system/info`。
2. 注册、登录、refresh、logout、`/auth/me`。
3. 私有上传、书架列表、书架详情、删除。
4. 阅读 meta、chapters、chapter content、progress。
5. PDF `Range` 请求返回 206。
6. 公共图书上传、下架、加入书架。
7. 分类和标签筛选。
8. 管理员用户、系统设置、存储统计。
