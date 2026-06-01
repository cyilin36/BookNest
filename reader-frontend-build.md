# 阅读器前端详细构建文档

本文档基于 `reader-project-plan.md`、`reader-backend-build.md` 和 `legado-local-book-parsing.md` 编写，用于后续前端开发落地。

本文档不规定视觉风格、配色、排版审美和具体设计语言，只规定前端功能边界、技术结构、接口协作、状态管理、阅读器数据流和部署要求。

## 1. 前端目标

前端负责提供以下能力：

- 用户注册、登录、退出、刷新登录状态。
- 当前用户信息和权限状态展示。
- 我的书架列表、筛选、搜索、详情、移除。
- 私有图书上传。
- 公共图书馆浏览、搜索、筛选、上传、加入书架。
- EPUB、PDF、TXT 在线阅读。
- 章节目录展示和按章节读取正文。
- 阅读进度读取、保存和继续阅读。
- 基础阅读设置，例如字号、主题、行距。
- 分类和标签选择。
- 管理员用户管理、公共图书审核、分类标签管理、系统设置。
- Docker 统一打包，由 Go 后端托管前端静态资源。

前端第一目标是稳定、清晰、能支撑阅读场景。不把图书解析逻辑放在前端；前端只消费后端提供的元数据、章节目录、章节内容和文件流。

## 2. 技术栈

固定技术栈：

- 框架：Vue 3。
- 构建工具：Vite。
- 语言：TypeScript。
- 路由：Vue Router。
- 状态管理：Pinia。
- UI 组件库：Naive UI。
- HTTP 客户端：Axios。
- EPUB 阅读：epub.js。
- PDF 阅读：pdf.js。
- TXT 阅读：自研章节内容渲染组件。

建议依赖：

- `vue`
- `vue-router`
- `pinia`
- `axios`
- `naive-ui`
- `@vueuse/core`
- `epubjs`
- `pdfjs-dist`

开发约束：

- 所有 API 类型放在 `src/api/types.ts` 或按模块拆分。
- 所有接口请求从 `src/api/` 发出，页面组件不直接写 Axios URL。
- 认证状态放在 Pinia，不散落在组件局部状态中。
- 阅读器状态可以按格式拆分 store，但持久化策略必须统一。
- 不在前端实现 EPUB/TXT/PDF 的源文件解析规则。

## 3. 文本解析边界

文本解析规则不由前端实现。

后端上传图书后负责解析：

```text
前端上传文件
  |
后端接收上传并落盘
  |
后端按格式解析元数据、封面、章节目录
  |
后端写入 books 和 book_chapters
  |
前端请求章节目录
  |
前端按章节请求内容并渲染
```

解析规则来源：

- TXT、EPUB、PDF 的解析规则参考 `legado-local-book-parsing.md`。
- 该文档是基于 Legado 项目分析出的解析规则说明。
- 后端实现时应优先复用或移植 Legado 的本地书籍解析思路。
- 前端不复刻 Legado 的 TXT 目录正则、EPUB fragment 截取、PDF 分段规则。

前端只依赖这些后端接口：

```http
GET /api/v1/reader/books/:bookId/meta
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/file
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
```

说明：

- EPUB 可用 `epub.js` 直接消费后端文件接口，也可以在需要统一体验时消费后端章节内容接口。
- PDF 使用 `pdf.js` 读取后端 Range 文件接口。
- TXT 优先使用后端章节目录和章节内容接口，不直接拉取整本 TXT。
- 前端阅读器不要假设章节规则；章节标题、顺序、正文范围以后端返回为准。

## 4. 项目目录

```text
frontend/
  src/
    api/
      client.ts
      types.ts
      auth.ts
      user.ts
      bookshelf.ts
      library.ts
      reader.ts
      category.ts
      tag.ts
      admin.ts
      system.ts
    assets/
    components/
      common/
      book/
      reader/
      upload/
      admin/
    composables/
      useAuth.ts
      usePagination.ts
      useReaderProgress.ts
      useUpload.ts
    layouts/
      AuthLayout.vue
      AppLayout.vue
      ReaderLayout.vue
      AdminLayout.vue
    router/
      index.ts
      guards.ts
    stores/
      auth.ts
      system.ts
      bookshelf.ts
      library.ts
      reader.ts
      settings.ts
    views/
      auth/
        LoginView.vue
        RegisterView.vue
      bookshelf/
        BookshelfView.vue
        BookshelfDetailView.vue
      library/
        LibraryView.vue
        LibraryDetailView.vue
      reader/
        ReaderView.vue
      upload/
        UploadView.vue
      admin/
        AdminHomeView.vue
        AdminUsersView.vue
        AdminLibraryView.vue
        AdminCategoriesView.vue
        AdminTagsView.vue
        AdminSettingsView.vue
      settings/
        SettingsView.vue
    App.vue
    main.ts
  index.html
  package.json
  vite.config.ts
  tsconfig.json
```

约定：

- `api/` 只负责请求和响应类型。
- `stores/` 负责跨页面状态。
- `views/` 负责页面级组合。
- `components/reader/` 放不同格式阅读器组件。
- `composables/` 放可复用交互逻辑，例如分页、进度节流、上传状态。
- `layouts/` 只组织页面骨架，不写业务请求。

## 5. 环境变量

```text
VITE_API_BASE=/api/v1
VITE_APP_NAME=Book Reader
```

说明：

- Docker 统一部署时，前端和后端同源，`VITE_API_BASE=/api/v1`。
- 本地开发可通过 Vite proxy 转发到后端。
- 不在前端环境变量中保存任何密钥。

Vite proxy 示例：

```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080'
  }
}
```

## 6. API 客户端

### 6.1 Axios 实例

`src/api/client.ts` 创建统一 Axios 实例。

默认配置：

- `baseURL` 使用 `import.meta.env.VITE_API_BASE || '/api/v1'`。
- `withCredentials=true`，因为 refresh token 使用 HttpOnly Cookie。
- 请求头携带 access token：`Authorization: Bearer <token>`。
- 响应拦截统一处理错误格式。

### 6.2 Token 策略

认证模型：

- Access Token 保存在内存和 Pinia 状态中。
- 可选持久化到 `sessionStorage`，用于刷新页面后短期恢复。
- Refresh Token 由后端写入 HttpOnly Cookie，前端 JavaScript 不能读取。
- 刷新登录状态时调用 `/auth/refresh`，依赖浏览器自动携带 Cookie。

刷新流程：

```text
API 返回 401
  |
若当前没有刷新请求，调用 /auth/refresh
  |
刷新成功，更新 access token
  |
重放原请求
  |
刷新失败，清空认证状态并跳转登录页
```

并发控制：

- 同一时间只允许一个 refresh 请求。
- 多个 401 请求等待同一个 refresh 结果。
- refresh 失败后不要无限重试。

### 6.3 错误处理

后端错误格式：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "请求参数不合法"
  },
  "request_id": "req_xxx"
}
```

前端处理规则：

- 业务页面展示 `message`。
- 开发调试可展示 `request_id`。
- 401 交给认证拦截器处理。
- 403 展示无权限提示。
- 404 页面内处理为空状态或跳转 Not Found。
- 上传体积过大展示后端返回的限制信息。

## 7. 类型模型

核心类型建议：

```ts
export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'disabled'
export type BookFormat = 'epub' | 'pdf' | 'txt'
export type BookVisibility = 'private' | 'public'
export type LibraryStatus = 'pending' | 'approved' | 'rejected' | 'hidden' | 'deleted'
export type BookshelfSourceType = 'uploaded' | 'library'
export type ProgressType = 'epub_cfi' | 'pdf_page' | 'txt_offset'

export interface User {
  id: number
  username: string
  email?: string | null
  nickname?: string | null
  role: UserRole
  status: UserStatus
  storage_quota_bytes?: number | null
}

export interface BookshelfItem {
  id: number
  book_id: number
  title: string
  author?: string | null
  format: BookFormat
  cover_url?: string | null
  source_type: BookshelfSourceType
  visibility: BookVisibility
  library_status?: LibraryStatus | null
  favorite: boolean
  pinned: boolean
  last_read_at?: string | null
  added_at: string
  readable: boolean
}

export interface ReaderChapter {
  id: number
  chapter_index: number
  title: string
  is_volume: boolean
  word_count?: number | null
}

export interface ReaderChapterContent {
  id: number
  chapter_index: number
  title: string
  content_type: 'html' | 'text'
  content: string
}

export interface ReadingProgress {
  progress_type: ProgressType
  progress_value: string
  percentage?: number | null
  updated_at?: string
}
```

约定：

- 后端 ID 使用 number。
- 时间字段使用 ISO string。
- 可空字段显式标注 `null`。
- 不在组件里手写重复类型。

## 8. 路由

路由规划：

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

路由守卫：

- 匿名可访问：`/login`、`/register`。
- 登录用户可访问：书架、图书馆、阅读器、上传、设置。
- 管理员可访问：`/admin/*`。
- 已登录用户访问 `/login`、`/register` 时跳转 `/bookshelf`。
- 未登录用户访问受保护页面时跳转 `/login`，并携带 redirect。

启动时认证恢复：

```text
App 启动
  |
若 sessionStorage 有 access token，先尝试 /auth/me
  |
失败则尝试 /auth/refresh
  |
刷新成功进入目标页面
  |
刷新失败进入匿名状态
```

## 9. 状态管理

### 9.1 auth store

保存：

- `accessToken`
- `user`
- `isAuthenticated`
- `isAdmin`
- `authReady`

动作：

- `register(payload)`
- `login(payload)`
- `refresh()`
- `logout()`
- `fetchMe()`
- `clearAuth()`

### 9.2 system store

保存：

- `siteName`
- `allowRegistration`
- `libraryReviewRequired`
- `supportedFormats`
- `maxUploadSizeMb`
- `defaultUserStorageQuotaMb`

动作：

- `fetchSystemInfo()`

### 9.3 bookshelf store

保存：

- 当前筛选条件。
- 当前分页。
- 书架列表。
- 加载状态。

动作：

- `fetchBookshelf()`
- `removeBookshelfItem(id)`
- `updateBookshelfItem(id, payload)`

### 9.4 reader store

保存：

- `bookMeta`
- `chapters`
- `activeChapterId`
- `chapterContentCache`
- `progress`
- `readerSettings`
- `loading`
- `error`

动作：

- `loadMeta(bookId)`
- `loadChapters(bookId)`
- `loadChapterContent(bookId, chapterId)`
- `loadProgress(bookId)`
- `saveProgress(bookId, progress)`
- `updateReaderSettings(settings)`

缓存规则：

- 当前章节内容缓存。
- 可预加载上一章和下一章。
- 缓存数量设上限，避免长时间阅读占用过多内存。

## 10. 页面规划

本节规定每个页面的路由、权限、数据接口、主要状态、用户操作和跳转结果。页面可以自由设计视觉风格，但不能改变业务行为和接口边界。

### 10.1 登录页

路由：

```text
/login
```

权限：

- 匿名用户可访问。
- 已登录用户访问时跳转 `/bookshelf` 或 redirect 目标。

使用接口：

```http
GET  /api/v1/system/info
POST /api/v1/auth/login
GET  /api/v1/auth/me
```

功能：

- 用户名或邮箱登录。
- 密码输入。
- 登录成功后跳转 redirect 或 `/bookshelf`。
- 根据 `system.info.allow_registration` 决定是否展示注册入口。

页面状态：

- `loadingSystemInfo`
- `submitting`
- `errorMessage`
- `redirect`

表单字段：

- `username`，用户名或邮箱。
- `password`。

提交成功：

- 保存 access token。
- 拉取当前用户信息。
- 跳转 redirect；没有 redirect 时跳转 `/bookshelf`。

提交失败：

- `invalid_credentials` 展示账号或密码错误。
- `user_disabled` 展示账号已禁用。
- 其他错误展示后端 message。

### 10.2 注册页

路由：

```text
/register
```

权限：

- 匿名用户可访问。
- 已登录用户访问时跳转 `/bookshelf`。

使用接口：

```http
GET  /api/v1/system/info
POST /api/v1/auth/register
GET  /api/v1/auth/me
```

功能：

- 用户名、邮箱、昵称、密码。
- 注册关闭时展示不可注册状态。
- 第一个用户注册成功后，后端返回角色 `admin`，前端展示管理员身份提示。

页面状态：

- `loadingSystemInfo`
- `registrationAllowed`
- `submitting`
- `errorMessage`

表单字段：

- `username`
- `email`
- `nickname`
- `password`
- `confirmPassword`，仅前端校验使用。

提交成功：

- 保存 access token。
- 拉取当前用户信息。
- 如果用户角色为 `admin`，展示首个管理员提示。
- 跳转 `/bookshelf`。

提交失败：

- `registration_disabled` 展示注册关闭。
- `username_exists` 展示用户名已存在。
- `email_exists` 展示邮箱已存在。
- 其他错误展示后端 message。

### 10.3 我的书架

路由：

```text
/bookshelf
```

权限：

- 登录用户可访问。

使用接口：

```http
GET    /api/v1/bookshelf
GET    /api/v1/categories?scope=system
GET    /api/v1/tags?scope=system
DELETE /api/v1/bookshelf/:id
```

功能：

- 列表展示当前用户书架。
- 搜索、格式筛选、分类筛选、标签筛选。
- 最近阅读排序。
- 上传入口。
- 继续阅读入口。
- 移除图书。
- 不可读图书展示状态，例如公共图书已隐藏或已删除。

页面状态：

- `items`
- `pagination`
- `keyword`
- `format`
- `categoryId`
- `tagId`
- `sourceType`
- `favorite`
- `sort`
- `order`
- `loading`
- `errorMessage`

URL query：

```text
keyword
format
category_id
tag_id
source_type
favorite
page
page_size
sort
order
```

用户操作：

- 点击图书进入 `/bookshelf/:id`。
- 点击继续阅读进入 `/reader/:bookId`。
- 点击上传进入 `/upload?target=bookshelf`。
- 修改筛选条件后重新请求列表。
- 删除或移除书架项前弹确认。

空状态：

- 没有任何图书时展示上传入口。
- 有筛选条件但无结果时展示清除筛选入口。

### 10.4 图书详情

路由：

```text
/bookshelf/:id
```

权限：

- 登录用户可访问。
- 只能查看自己的书架记录。

使用接口：

```http
GET   /api/v1/bookshelf/:id
PATCH /api/v1/bookshelf/:id
GET   /api/v1/categories?scope=system
GET   /api/v1/tags?scope=system
GET   /api/v1/reader/books/:bookId/meta
```

功能：

- 展示图书元数据。
- 展示分类、标签、来源、状态。
- 打开阅读器。
- 编辑个人标题、收藏、置顶、分类、标签。

页面状态：

- `bookshelfItem`
- `bookMeta`
- `categoryOptions`
- `tagOptions`
- `editing`
- `saving`
- `errorMessage`

用户操作：

- 打开阅读器：跳转 `/reader/:bookId`。
- 保存个性化设置：调用 `PATCH /bookshelf/:id`。
- 返回书架：跳转 `/bookshelf` 并保留 query 状态。

错误处理：

- 404 展示图书不存在或已移除。
- 403 展示无权访问。
- `readable=false` 时禁用打开阅读器入口，并展示不可读原因。

### 10.5 上传页

路由：

```text
/upload
```

权限：

- 登录用户可访问。

使用接口：

```http
GET  /api/v1/system/info
GET  /api/v1/categories?scope=system
GET  /api/v1/tags?scope=system
POST /api/v1/bookshelf/upload
POST /api/v1/library/books/upload
```

功能：

- 选择上传到个人书架或公共图书馆。
- 文件选择。
- 客户端基础校验扩展名和大小。
- 元数据表单：标题、作者、简介、分类、标签。
- 上传进度展示。
- 上传成功后展示解析结果入口或跳转目标页面。

页面状态：

- `target`，`bookshelf` 或 `library`。
- `file`
- `title`
- `author`
- `description`
- `categoryIds`
- `tagIds`
- `uploading`
- `uploadProgress`
- `result`
- `errorMessage`

客户端校验：

- 文件扩展名只能是 `epub`、`pdf`、`txt`。
- 文件大小不超过 `system.info.max_upload_size_mb`。
- 标题、作者、简介可以为空，后端会解析或兜底。

提交成功：

- 私有上传成功后可跳转 `/bookshelf` 或 `/reader/:bookId`。
- 公共上传成功后展示 `library_status`。
- 公共上传不自动加入个人书架。
- 若 `library_status='pending'`，展示“待审核”状态。

上传说明：

- 客户端校验只是提前提示，最终以后端校验为准。
- 公共图书上传成功后不会自动加入个人书架。
- 如果公共图书需要审核，上传成功后状态为 `pending`，表示待审核，不是上传中。

### 10.6 公共图书馆

路由：

```text
/library
```

权限：

- 登录用户可访问。
- 匿名用户不可访问。

使用接口：

```http
GET  /api/v1/library/books
GET  /api/v1/categories?scope=system
GET  /api/v1/tags?scope=system
POST /api/v1/library/books/:id/add-to-bookshelf
```

功能：

- 仅登录用户可访问。
- 列表展示 approved 公共图书。
- 搜索、格式、分类、标签筛选。
- 显示是否已加入书架。
- 加入书架。
- 上传公共图书入口。

页面状态：

- `items`
- `pagination`
- `keyword`
- `format`
- `categoryId`
- `tagId`
- `sort`
- `order`
- `loading`
- `joiningBookId`
- `errorMessage`

URL query：

```text
keyword
format
category_id
tag_id
page
page_size
sort
order
```

用户操作：

- 点击图书进入 `/library/:id`。
- 点击加入书架调用 add-to-bookshelf。
- 已加入图书不重复提交加入请求。
- 点击上传公共图书进入 `/upload?target=library`。

### 10.7 公共图书详情

路由：

```text
/library/:id
```

权限：

- 登录用户可访问。
- 普通用户只能查看 approved 公共图书。
- 上传者本人可查看自己 pending 的公共图书。

使用接口：

```http
GET  /api/v1/library/books/:id
POST /api/v1/library/books/:id/add-to-bookshelf
```

功能：

- 展示公共图书元数据。
- 展示加入状态。
- 加入书架。
- 上传者本人可查看自己 pending 的公共图书。

页面状态：

- `book`
- `inBookshelf`
- `bookshelfId`
- `joining`
- `errorMessage`

用户操作：

- 加入书架成功后更新 `inBookshelf=true`。
- 如果已加入，可跳转 `/reader/:bookId` 或 `/bookshelf/:bookshelfId`。
- 返回公共图书馆时保留 query。

错误处理：

- `library_book_not_approved` 展示未审核或不可见。
- `book_already_in_bookshelf` 更新本地加入状态。
- 404 展示图书不存在。

### 10.8 阅读器页

路由：

```text
/reader/:bookId
```

权限：

- 登录用户可访问。
- 必须拥有阅读权限。

使用接口：

```http
GET /api/v1/reader/books/:bookId/meta
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/file
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
GET /api/v1/books/:bookId/bookmarks
POST /api/v1/books/:bookId/bookmarks
```

功能：

- 加载图书元数据。
- 加载章节目录。
- 加载阅读进度。
- 根据格式选择阅读器组件。
- 保存阅读进度。
- 章节跳转。
- 阅读设置。
- 返回书架或图书馆。

页面状态：

- `bookMeta`
- `chapters`
- `activeChapterId`
- `chapterContent`
- `progress`
- `readerSettings`
- `loadingMeta`
- `loadingChapters`
- `loadingContent`
- `savingProgress`
- `errorMessage`

加载顺序：

```text
进入 /reader/:bookId
  |
GET meta
  |
GET chapters
  |
GET progress
  |
定位初始章节或位置
  |
加载章节内容或文件流
```

用户操作：

- 打开或关闭目录。
- 点击章节跳转。
- 上一章、下一章。
- 修改阅读设置。
- 手动保存进度。
- 返回上一来源页面。

格式行为：

- TXT：优先请求章节内容接口。
- EPUB：可使用文件接口交给 epub.js，也可使用章节接口做统一目录展示。
- PDF：使用文件接口和 Range 请求，章节目录作为页段跳转。

错误处理：

- `book_not_accessible` 展示无权阅读。
- `book_file_missing` 展示文件缺失。
- 章节内容加载失败时提供重试。

### 10.9 设置页

路由：

```text
/settings
```

权限：

- 登录用户可访问。

使用接口：

```http
GET   /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password
```

功能：

- 查看当前用户资料。
- 修改昵称、邮箱。
- 修改密码。
- 查看自己的角色和状态。
- 查看自己的存储配额和已用空间，如果后端返回。

页面状态：

- `profile`
- `profileSaving`
- `passwordSaving`
- `errorMessage`
- `successMessage`

表单字段：

- 资料表单：`nickname`、`email`。
- 密码表单：`old_password`、`new_password`、`confirm_password`。

提交成功：

- 更新 auth store 中的用户信息。
- 修改密码成功后后端可能撤销 refresh token，前端应重新登录或刷新当前认证状态。

### 10.10 管理首页

路由：

```text
/admin
```

权限：

- 管理员可访问。

使用接口：

```http
GET /api/v1/admin/system/storage
GET /api/v1/admin/system/settings
GET /api/v1/admin/library/books?page=1&page_size=5
GET /api/v1/admin/users?page=1&page_size=5
```

功能：

- 管理后台入口聚合。
- 展示存储概览。
- 展示待审核公共图书入口。
- 展示用户管理入口。
- 展示分类标签管理入口。

页面状态：

- `storageSummary`
- `pendingLibraryBooks`
- `recentUsers`
- `settingsSummary`
- `loading`
- `errorMessage`

### 10.11 管理员用户页

路由：

```text
/admin/users
```

权限：

- 管理员可访问。

使用接口：

```http
GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
```

功能：

- 用户列表。
- 搜索用户。
- 查看用户详情。
- 禁用或启用用户。
- 调整用户角色。
- 调整用户存储配额。

页面状态：

- `users`
- `pagination`
- `keyword`
- `status`
- `role`
- `selectedUser`
- `editingQuota`
- `loading`
- `savingUserId`
- `errorMessage`

用户操作：

- 修改状态前确认。
- 修改角色前确认。
- 修改配额后刷新当前用户行。
- 后端拒绝管理员修改自身状态或角色时展示错误。

### 10.12 管理员图书页

路由：

```text
/admin/books
```

权限：

- 管理员可访问。

使用接口：

```http
GET /api/v1/admin/library/books
```

功能：

- 作为全站图书管理入口。
- 第一阶段可以复用公共图书管理列表。
- 后续如果后端增加私有图书管理接口，再扩展查看私有图书。

页面状态：

- `items`
- `pagination`
- `keyword`
- `format`
- `visibility`
- `libraryStatus`
- `loading`
- `errorMessage`

说明：

- 当前后端文档只明确了管理员公共图书接口。
- 因此 `/admin/books` 第一版可以跳转或内嵌 `/admin/library`。

### 10.13 管理员公共图书页

路由：

```text
/admin/library
```

权限：

- 管理员可访问。

使用接口：

```http
GET    /api/v1/admin/library/books
PATCH  /api/v1/admin/library/books/:id/status
DELETE /api/v1/admin/library/books/:id
DELETE /api/v1/admin/library/books/:id?delete_file=true
```

功能：

- 查看所有公共图书，包括 pending、approved、rejected、hidden、deleted。
- 审核 pending 图书。
- 拒绝 pending 图书。
- 隐藏 approved 图书。
- 恢复 hidden 图书。
- 删除公共图书。
- 删除时选择是否物理删除文件。

页面状态：

- `items`
- `pagination`
- `keyword`
- `format`
- `libraryStatus`
- `loading`
- `operatingBookId`
- `deleteFileConfirm`
- `errorMessage`

用户操作：

- 审核通过：状态改为 `approved`。
- 审核拒绝：状态改为 `rejected`。
- 隐藏：状态改为 `hidden`。
- 恢复：状态改为 `approved`。
- 删除：默认软删除；勾选物理删除时调用 `delete_file=true`。

### 10.14 管理员分类页

路由：

```text
/admin/categories
```

权限：

- 管理员可访问。

使用接口：

```http
GET    /api/v1/categories?scope=system
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

功能：

- 分类列表。
- 新增分类。
- 修改分类名称、slug、描述、父级。
- 删除未使用分类。

页面状态：

- `categories`
- `editingCategory`
- `creating`
- `saving`
- `deletingId`
- `errorMessage`

用户操作：

- 创建前校验名称和 slug 非空。
- 删除前确认。
- 后端返回有关联图书时展示不可删除原因。

### 10.15 管理员标签页

路由：

```text
/admin/tags
```

权限：

- 管理员可访问。

使用接口：

```http
GET    /api/v1/tags?scope=system
POST   /api/v1/admin/tags
PATCH  /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

功能：

- 标签列表。
- 新增标签。
- 修改标签名称和 slug。
- 删除未使用标签。

页面状态：

- `tags`
- `keyword`
- `editingTag`
- `creating`
- `saving`
- `deletingId`
- `errorMessage`

用户操作：

- 创建前校验名称和 slug 非空。
- 删除前确认。
- 后端返回有关联图书或书架时展示不可删除原因。

### 10.16 管理员系统设置页

路由：

```text
/admin/settings
```

权限：

- 管理员可访问。

使用接口：

```http
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
GET /api/v1/admin/system/storage
```

功能：

- 修改站点名称。
- 开关用户注册。
- 开关公共图书审核。
- 修改最大上传大小。
- 修改默认用户存储配额。
- 查看存储统计。

页面状态：

- `settings`
- `storageSummary`
- `saving`
- `loading`
- `errorMessage`
- `successMessage`

用户操作：

- 保存设置。
- 保存成功后刷新 system store。
- 最大上传大小超过后端限制时展示错误。

### 10.17 Not Found 页面

路由：

```text
/:pathMatch(.*)*
```

权限：

- 任意用户可访问。

功能：

- 展示页面不存在。
- 已登录用户提供返回书架入口。
- 匿名用户提供返回登录入口。

## 11. 阅读器实现

### 11.1 通用阅读器壳

`ReaderView.vue` 负责：

- 读取路由 `bookId`。
- 调用 reader store 加载 meta、chapters、progress。
- 控制目录展示和章节切换。
- 根据 `bookMeta.format` 选择具体阅读器。
- 统一处理阅读设置和进度保存。

组件建议：

```text
ReaderView.vue
  ReaderToolbar.vue
  ReaderToc.vue
  ReaderSettingsPanel.vue
  EpubReader.vue
  PdfReader.vue
  TxtReader.vue
```

### 11.2 EPUB 阅读器

实现方式：

- 使用 `epub.js` 加载 `/api/v1/reader/books/:bookId/file`。
- 请求必须带 `Authorization` header。
- 如果 epub.js 资源请求无法稳定携带认证头，需要改为后端签发短期阅读 URL 或由前端 fetch blob 后交给 epub.js。

章节目录：

- 页面目录优先展示后端 `GET /chapters` 返回结果。
- epub.js 内部目录仅作为渲染定位辅助，不作为业务目录真源。

进度：

- 使用 EPUB CFI 保存。
- 切章或位置变化时节流保存。
- 保存格式：`progress_type='epub_cfi'`。

### 11.3 PDF 阅读器

实现方式：

- 使用 `pdf.js`。
- 文件来源为 `/api/v1/reader/books/:bookId/file`。
- 后端必须支持 Range 请求。
- 若 pdf.js 无法携带认证头，需要通过 pdf.js `httpHeaders` 配置传入 Authorization。

章节目录：

- 后端按每 10 页生成一个章节分段。
- 前端目录点击后跳转到对应页。

进度：

- 保存当前页码。
- 保存格式：`progress_type='pdf_page'`。
- 可附带百分比。

### 11.4 TXT 阅读器

实现方式：

- 不拉取完整 TXT 文件。
- 先请求后端章节目录。
- 用户进入章节时请求章节内容。
- 内容由前端渲染为文本阅读视图。

章节内容：

- 请求：`GET /api/v1/reader/books/:bookId/chapters/:chapterId/content`。
- 后端按 Legado 规则解析出的字节范围读取正文。
- 前端不处理 TXT 目录正则，不识别章节标题。

进度：

- 保存当前章节内位置或全书字节偏移。
- 后端进度类型为 `txt_offset`。
- 若后端返回章节 start/end，可由前端换算当前 offset；否则保存章节 id 和滚动比例作为本地辅助状态。

### 11.5 阅读进度保存

触发时机：

- 用户切换章节。
- 用户停止滚动一段时间。
- 页面隐藏或离开阅读器。
- 阅读器组件卸载。

节流策略：

- 常规阅读中 5 到 10 秒最多保存一次。
- 章节切换、页面关闭前可以立即保存一次。
- 避免每次滚动都请求后端。

本地兜底：

- 后端保存失败时，可把最近进度临时保存到 `localStorage`。
- 下次进入同一本书时，若本地进度比后端进度新，可以提示用户恢复。

## 12. 上传实现

上传请求：

```http
POST /api/v1/bookshelf/upload
POST /api/v1/library/books/upload
```

表单字段：

```text
file
title
author
description
category_ids
tag_ids
```

前端流程：

```text
选择文件
  |
校验扩展名 epub/pdf/txt
  |
校验大小不超过 system.info.max_upload_size_mb
  |
填写元数据、分类、标签
  |
FormData 上传
  |
展示上传进度
  |
成功后展示结果
```

公共上传状态：

- `pending`：已上传完成，等待管理员审核。
- `approved`：已进入公共图书馆。
- `rejected`：审核拒绝。
- `hidden`：管理员隐藏。
- `deleted`：管理员删除。

## 13. 分类和标签

规则：

- 分类和标签由管理员维护。
- 普通用户不能创建分类或标签。
- 普通用户上传、编辑图书时，只能从已有分类和标签中选择。
- 书架和图书馆筛选使用同一套分类、标签数据。

接口：

```http
GET /api/v1/categories?scope=system
GET /api/v1/tags?scope=system
POST /api/v1/admin/categories
PATCH /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
POST /api/v1/admin/tags
PATCH /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

前端行为：

- 上传页加载分类和标签选项。
- 书架筛选加载分类和标签选项。
- 公共图书馆筛选加载分类和标签选项。
- 管理后台提供分类和标签维护入口。

## 14. 管理后台

### 14.1 用户管理

功能：

- 用户列表。
- 搜索用户。
- 查看用户状态、角色、注册时间、最后登录时间。
- 禁用或启用用户。
- 调整用户角色。
- 调整用户存储配额。

约束：

- 管理员不能禁用自己。
- 管理员不能随意把自己降级，具体以后端校验为准。

### 14.2 公共图书管理

功能：

- 查看所有公共图书，包括 pending、approved、rejected、hidden、deleted。
- 审核 pending 图书。
- 隐藏 approved 图书。
- 恢复 hidden 图书。
- 删除公共图书。
- 删除时选择是否物理删除文件。

### 14.3 分类标签管理

功能：

- 创建分类。
- 修改分类。
- 删除未使用分类。
- 创建标签。
- 修改标签。
- 删除未使用标签。

### 14.4 系统设置

功能：

- 是否允许注册。
- 公共图书是否需要审核。
- 站点名称。
- 最大上传大小。
- 默认用户存储配额。

## 15. 权限处理

前端根据当前用户角色控制入口展示，但不能把前端控制当作安全边界。

规则：

- 普通用户不显示管理后台入口。
- 管理员入口需要后端接口再次校验。
- 403 时展示无权限状态。
- 401 时尝试 refresh，失败后跳转登录。
- 公共图书馆仅登录用户可访问。
- 匿名用户只能访问登录、注册和系统信息所需页面。

## 16. 数据加载策略

列表页面：

- 进入页面加载第一页。
- 筛选条件变化后重置到第一页。
- 搜索输入需要 debounce。
- 分页、排序、筛选状态同步到 URL query，便于刷新和分享。

详情页面：

- 根据路由 id 加载详情。
- 404 展示不存在或已删除。
- 403 展示无权限。

阅读器页面：

- 先加载 meta，再加载 chapters 和 progress。
- 当前章节内容加载失败时允许重试。
- 可预取下一章内容。

## 17. 本地持久化

可持久化：

- Access Token，建议 `sessionStorage`。
- 阅读设置，例如字号、行距、主题。
- 最近一次未成功上传的表单草稿，不包含文件对象。
- 阅读进度保存失败时的临时兜底。

不可持久化：

- Refresh Token，必须由 HttpOnly Cookie 管理。
- 用户密码。
- 大段章节内容。
- 图书文件 Blob。

## 18. 可访问性和交互底线

本文档不规定视觉审美，但前端必须满足基本可用性：

- 所有表单字段有明确标签。
- 关键按钮有 loading 和 disabled 状态。
- 错误提示能被用户看到。
- 阅读器支持键盘基本操作，例如上一章、下一章或翻页。
- 移动端可以完成登录、上传以外的主要阅读和浏览操作。
- 删除、禁用、物理删除文件等危险操作需要确认。

## 19. 测试计划

### 19.1 单元测试

建议覆盖：

- API client token 注入。
- 401 refresh 队列。
- 分页参数构造。
- 阅读进度节流。
- TXT 章节内容加载状态。

### 19.2 组件测试

建议覆盖：

- 登录表单。
- 注册表单。
- 上传表单校验。
- 书架筛选。
- 公共图书加入书架按钮状态。
- 阅读器目录切换。

### 19.3 端到端测试

建议覆盖：

- 首个用户注册成为管理员。
- 登录和刷新登录状态。
- 上传私有图书。
- 打开阅读器并保存进度。
- 上传公共图书后显示 pending。
- 管理员审核公共图书。
- 普通用户加入公共图书到书架。
- 管理员维护分类和标签。
- 管理员调整用户存储配额。

## 20. 开发阶段拆解

### 阶段一：前端骨架

交付：

- Vite + Vue 3 + TypeScript 项目。
- Router。
- Pinia。
- Axios client。
- 基础布局。
- 系统信息接口。

### 阶段二：认证

交付：

- 登录页。
- 注册页。
- auth store。
- refresh token 流程。
- 路由守卫。

### 阶段三：书架和上传

交付：

- 我的书架列表。
- 书架筛选。
- 私有上传。
- 图书详情。
- 移除图书。

### 阶段四：阅读器

交付：

- ReaderView。
- 章节目录。
- TXT 章节阅读。
- EPUB 阅读。
- PDF 阅读。
- 阅读进度保存。
- 阅读设置持久化。

### 阶段五：公共图书馆

交付：

- 公共图书列表。
- 公共图书详情。
- 公共上传。
- 加入书架。
- pending 状态展示。

### 阶段六：分类标签

交付：

- 分类选项加载。
- 标签选项加载。
- 上传和编辑时选择分类标签。
- 管理员分类管理。
- 管理员标签管理。

### 阶段七：管理后台

交付：

- 用户管理。
- 用户配额调整。
- 公共图书审核。
- 公共图书删除和物理删除选项。
- 系统设置。
- 存储统计。

### 阶段八：稳定性

交付：

- 错误边界。
- 重试能力。
- 上传体验完善。
- 阅读器性能优化。
- E2E 测试。

## 21. 构建和部署

构建命令：

```bash
npm ci
npm run build
```

构建产物：

```text
frontend/dist
```

Docker 集成：

- Dockerfile 的 frontend-builder 阶段执行前端构建。
- 构建产物复制到后端 `backend/internal/app/static`。
- Go 后端最终托管前端静态资源。
- 浏览器访问 `http://server:8080`，API 和前端同源。

## 22. 已确认实现决策

1. 前端不实现书籍文本解析规则。
2. TXT、EPUB、PDF 解析规则以后端为准，后端参考 `legado-local-book-parsing.md` 和 Legado 项目实现。
3. 前端阅读 TXT 时优先按章节请求内容，不拉取整本书。
4. 前端不规定视觉风格审美。
5. Refresh Token 使用 HttpOnly Cookie，前端只管理 Access Token。
6. 公共图书上传成功后的 `pending` 是待审核状态，不是上传中。
7. 公共图书上传后不自动加入个人书架。
8. 分类和标签由管理员维护，普通用户只能选择已有项。
9. 管理员可调整用户存储配额。
10. 公共图书馆不允许匿名访问。
