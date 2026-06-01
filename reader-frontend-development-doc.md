# 阅读器前端开发文档

本文档基于 `reader-project-plan.md`、`reader-frontend-build.md`、`VISUAL_STYLE_GUIDE.md` 和 `frontend-style-demo.html` 汇总整理，是后续前端开发的主依据。若其他前端相关文档与本文档冲突，开发时优先按本文档执行；确需调整时，应先更新本文档再编码。

`frontend-style-demo.html` 是视觉和多端交互参照，不是正式技术栈约束。正式前端必须使用 Vue 3、Vite、TypeScript、Vue Router、Pinia、Naive UI、Axios、epub.js 和 pdf.js 实现。

## 1. 项目定位

本项目是一个支持多用户的 Web 阅读器系统。前端负责用户认证、书架、公共图书馆、上传、在线阅读、分类标签选择、管理后台和系统设置的交互体验；后端负责认证安全、权限校验、图书解析、文件鉴权、章节内容、阅读进度和数据持久化。

前端核心目标：

- 支持注册、登录、退出、刷新登录状态。
- 支持当前用户信息、角色和状态展示。
- 支持个人书架列表、筛选、搜索、详情、移除、继续阅读。
- 支持上传私有图书到个人书架。
- 支持公共图书馆浏览、搜索、筛选、上传和加入书架。
- 支持 EPUB、PDF、TXT 在线阅读。
- 支持章节目录、章节跳转、阅读进度读取和保存。
- 支持基础阅读设置：主题、字号、行高、字体。
- 支持分类和标签选择。
- 支持管理员用户管理、公共图书审核、分类标签管理、系统设置。
- 支持 Docker 统一打包，由 Go 后端托管前端静态资源。

第一期坚持稳定、清晰、轻量，不在前端实现图书源文件解析规则，不引入重型前端框架或不必要的状态同步系统。

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
- TXT 阅读：自研章节正文渲染组件。

推荐依赖：

```text
vue
vue-router
pinia
axios
naive-ui
@vueuse/core
epubjs
pdfjs-dist
lucide-vue-next
```

约束：

- 所有 API 请求必须从 `src/api/` 发出，页面组件不能直接写 Axios URL。
- 所有接口类型放在 `src/api/types.ts` 或按业务模块拆分到 `src/api/*Types.ts`。
- 认证状态统一放在 Pinia `auth` store。
- 阅读器状态可以拆分 store，但持久化策略必须统一。
- 组件内不得重复定义后端响应类型。
- 不在前端实现 TXT 章节正则、EPUB fragment 截取、PDF 分段等源文件解析规则。
- 不在前端保存 refresh token、密码、图书文件 Blob 或大段章节缓存。

## 3. 职责边界

后端负责：

- 用户认证、角色权限、禁用状态。
- 图书上传、格式校验、文件落盘、文件鉴权。
- 元数据抽取、封面抽取、章节目录生成。
- TXT 编码识别、目录规则匹配、字节偏移分章。
- EPUB 目录解析、章节正文读取、资源图片处理。
- PDF 文件 Range 读取和后续可能的页段章节生成。
- 阅读进度、书签、分类、标签、系统设置持久化。

前端负责：

- 页面布局、表单、列表、筛选、分页和交互反馈。
- 请求后端提供的元数据、章节目录、章节正文、文件流。
- TXT/EPUB/PDF 的浏览器端渲染和用户操作。
- 阅读设置、本地 UI 偏好、进度节流上报。
- 401 refresh 队列、403/404/业务错误展示。

阅读器数据流：

```text
用户上传文件
  |
后端接收 multipart stream 并落盘
  |
后端解析元数据、封面、章节目录
  |
前端请求 meta / chapters / progress
  |
前端按格式渲染阅读器
  |
前端节流上报阅读进度
```

前端只依赖这些阅读接口：

```http
GET /api/v1/reader/books/:bookId/meta
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/file
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
```

## 4. 项目目录

前端统一放在 `frontend/`。

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
      styles/
        variables.css
        themes.css
        reader.css
    components/
      common/
      book/
      reader/
      upload/
      admin/
    composables/
      useAuth.ts
      usePagination.ts
      useQuerySync.ts
      useReaderProgress.ts
      useReaderSettings.ts
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
        AdminBooksView.vue
        AdminLibraryView.vue
        AdminCategoriesView.vue
        AdminTagsView.vue
        AdminSettingsView.vue
      settings/
        SettingsView.vue
      NotFoundView.vue
    App.vue
    main.ts
  index.html
  package.json
  vite.config.ts
  tsconfig.json
```

目录规则：

- `api/` 只负责请求函数、请求类型、响应类型和错误归一化。
- `stores/` 负责跨页面状态，不直接承载复杂 UI 组件细节。
- `views/` 负责页面级组合和路由参数读取。
- `layouts/` 负责页面骨架、导航、响应式区域，不直接写业务请求。
- `components/reader/` 放阅读器相关组件。
- `components/book/` 放书卡、封面、格式标签、进度条等可复用图书组件。
- `components/common/` 放空状态、错误状态、确认弹窗、分页包装、过滤栏等基础组件。
- `composables/` 放可复用交互逻辑。

## 5. 环境变量

```text
VITE_API_BASE=/api/v1
VITE_APP_NAME=Book Reader
```

说明：

- Docker 统一部署时前后端同源，`VITE_API_BASE=/api/v1`。
- 本地开发通过 Vite proxy 转发到后端。
- 前端环境变量中不得保存密钥。

Vite proxy：

```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080'
  }
}
```

## 6. API 客户端

`src/api/client.ts` 创建唯一 Axios 实例。

默认配置：

- `baseURL` 使用 `import.meta.env.VITE_API_BASE || '/api/v1'`。
- `withCredentials=true`，用于携带 HttpOnly refresh token cookie。
- 请求头携带 access token：`Authorization: Bearer <token>`。
- 响应拦截统一归一化后端错误。

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

错误处理规则：

- 业务页面展示 `message`。
- 开发调试可展示 `request_id`。
- 401 交给认证拦截器处理。
- 403 展示无权限状态。
- 404 在页面内展示空状态或跳转 Not Found。
- 上传体积过大展示后端返回的限制信息。
- 文件缺失、章节加载失败等阅读器错误必须提供重试入口。

401 refresh 流程：

```text
API 返回 401
  |
若当前无 refresh 请求，调用 /auth/refresh
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
- refresh 请求本身失败后不要无限重试。
- 退出登录时必须清空等待队列。

## 7. 类型模型

核心类型：

```ts
export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'disabled'
export type BookFormat = 'epub' | 'pdf' | 'txt'
export type BookVisibility = 'private' | 'public'
export type LibraryStatus = 'pending' | 'approved' | 'rejected' | 'hidden' | 'deleted'
export type BookshelfSourceType = 'uploaded' | 'library'
export type ProgressType = 'epub_cfi' | 'pdf_page' | 'txt_offset'
export type ThemeName = 'modern' | 'sepia' | 'dark'

export interface User {
  id: number
  username: string
  email?: string | null
  nickname?: string | null
  role: UserRole
  status: UserStatus
  storage_quota_bytes?: number | null
  storage_used_bytes?: number
}

export interface BookMeta {
  id: number
  title: string
  author?: string | null
  description?: string | null
  format: BookFormat
  cover_url?: string | null
  file_size: number
  visibility: BookVisibility
  library_status?: LibraryStatus | null
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
  unreadable_reason?: 'library_hidden' | 'library_deleted' | 'library_rejected' | 'file_missing' | 'permission_denied' | null
  progress_percentage?: number | null
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

export interface ReaderSettings {
  theme: ThemeName
  fontSize: number
  lineHeight: 1.5 | 1.8 | 2.2
  fontFamily: 'sans' | 'serif'
}
```

类型约定：

- 后端 ID 使用 `number`。
- 时间字段使用 ISO string。
- 可空字段显式标注 `null`。
- 枚举用联合字符串类型。
- 请求 payload 和响应 DTO 必须分开定义。

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
/:pathMatch(.*)*
```

路由守卫：

- 匿名可访问：`/login`、`/register`。
- 登录用户可访问：书架、图书馆、阅读器、上传、设置。
- 管理员可访问：`/admin/*`。
- 已登录用户访问 `/login`、`/register` 时跳转 `/bookshelf` 或 redirect 目标。
- 未登录用户访问受保护页面时跳转 `/login`，并携带 `redirect`。
- 普通用户访问 `/admin/*` 时展示 403 页面或跳转 `/bookshelf` 并提示无权限。

启动认证恢复：

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

状态：

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

状态：

- `siteName`
- `allowRegistration`
- `libraryReviewRequired`
- `supportedFormats`
- `maxUploadSizeMb`
- `defaultUserStorageQuotaMb`

动作：

- `fetchSystemInfo()`
- `refreshSystemInfo()`

### 9.3 bookshelf store

状态：

- 当前筛选条件。
- 当前分页。
- 书架列表。
- 加载状态。
- 最近一次错误。

动作：

- `fetchBookshelf()`
- `removeBookshelfItem(id)`
- `updateBookshelfItem(id, payload)`
- `resetFilters()`

### 9.4 library store

状态：

- 公共图书列表。
- 当前筛选条件。
- 当前分页。
- `joiningBookId`
- 加载状态。

动作：

- `fetchLibraryBooks()`
- `addToBookshelf(bookId)`
- `resetFilters()`

### 9.5 reader store

状态：

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
- `clearReaderState()`

缓存规则：

- 缓存当前章节内容。
- 可预加载上一章和下一章。
- 缓存数量设上限，建议 5 章以内。
- 不缓存整本书，不持久化大段章节内容。

### 9.6 settings store

状态：

- `theme`
- `readerFontSize`
- `readerLineHeight`
- `readerFontFamily`

动作：

- `loadLocalSettings()`
- `setTheme(theme)`
- `updateReaderSettings(partial)`
- `resetReaderSettings()`

持久化：

- 使用 `localStorage` 保存阅读设置和主题。
- Access Token 如需短期恢复使用 `sessionStorage`。

## 10. 视觉设计系统

视觉风格必须参考 `VISUAL_STYLE_GUIDE.md` 和 `frontend-style-demo.html`。正式实现使用 Naive UI 的主题覆盖、CSS variables 和局部 CSS，不直接依赖 demo 中的 Tailwind CDN。

### 10.1 设计原则

- 稳定清晰优先，界面去装饰化。
- 通过排版、留白、边框和状态标记建立层级。
- 不使用夸张营销式 hero 页面。
- 书架、图书馆、后台属于工具型界面，应密集但有秩序。
- 阅读器必须降低干扰，正文优先。
- 公共图书馆必须显化“引用添加，不复制文件”的存储模型。
- 移动端优先保证浏览、阅读、目录、设置和加入书架操作顺畅。

### 10.2 主题变量

全局主题提供三套：

```css
:root,
.theme-modern {
  --color-primary: #18a058;
  --color-primary-hover: #36ad6a;
  --color-primary-suppl: rgba(24, 160, 88, 0.1);
  --color-bg-page: #f3f4f6;
  --color-bg-card: #ffffff;
  --color-bg-sidebar: #ffffff;
  --color-border: #e5e7eb;
  --color-text-main: #1f2937;
  --color-text-sec: #6b7280;
  --font-reader: Inter, system-ui, sans-serif;
}

.theme-sepia {
  --color-primary: #8b5a2b;
  --color-primary-hover: #a06d3b;
  --color-primary-suppl: rgba(139, 90, 43, 0.1);
  --color-bg-page: #f4edd8;
  --color-bg-card: #fcf8ec;
  --color-bg-sidebar: #efe5cb;
  --color-border: #e3d5b5;
  --color-text-main: #2c2214;
  --color-text-sec: #6d583e;
  --font-reader: "Noto Serif SC", Georgia, serif;
}

.theme-dark {
  --color-primary: #10b981;
  --color-primary-hover: #34d399;
  --color-primary-suppl: rgba(16, 185, 129, 0.15);
  --color-bg-page: #09090b;
  --color-bg-card: #18181b;
  --color-bg-sidebar: #111113;
  --color-border: #27272a;
  --color-text-main: #f4f4f5;
  --color-text-sec: #a1a1aa;
  --font-reader: "Noto Serif SC", Georgia, serif;
}
```

Naive UI 主题映射：

- `primaryColor` 使用 `--color-primary`。
- `primaryColorHover` 使用 `--color-primary-hover`。
- `bodyColor` 使用 `--color-bg-page`。
- `cardColor` 使用 `--color-bg-card`。
- `borderColor` 使用 `--color-border`。
- `textColor1` 使用 `--color-text-main`。
- `textColor2` 使用 `--color-text-sec`。

### 10.3 响应式规则

断点：

- 桌面端：`lg >= 1024px`。
- 手机和平板：`lg < 1024px`。

桌面端：

- `AppLayout` 左侧栏常驻，宽度 256px。
- 页面主区域内边距 32px。
- 顶部操作栏常驻。
- 阅读器目录左侧常驻。
- 阅读器工具栏顶部常驻。

移动端：

- 隐藏桌面侧栏，显示顶部移动导航栏。
- 主导航使用侧滑 Drawer。
- 页面主区域内边距 16px。
- 阅读器目录和排版设置使用底部 Bottom Sheet Drawer。
- 阅读正文左右边距使用 `16px` 起步，平板可逐步增加。
- 所有关键点击区域高度不小于 36px。

### 10.4 通用组件视觉

书卡：

- 统一使用 8px 到 12px 圆角，正式实现优先 8px。
- 必须显示格式标签：TXT、EPUB、PDF。
- 必须显示来源标签：私有上传、引自公共馆。
- 底部显示阅读进度细线，背景灰色，填充主色。
- `readable=false` 时整卡弱化，并显示不可读原因。

格式色：

- TXT：琥珀色。
- EPUB：绿松或翡翠色。
- PDF：玫瑰或粉红色。

按钮：

- 主要动作使用 Naive UI primary。
- 危险动作使用 error。
- 次要动作使用 tertiary/quaternary。
- 已加入、不可读、无权限等状态必须 disabled。

图标：

- 优先使用 `lucide-vue-next`。
- 导航、上传、设置、目录、书签、搜索、筛选、删除、编辑等操作要有图标。
- 不熟悉的图标必须提供 tooltip。

## 11. 页面开发规格

### 11.1 登录页

路由：`/login`

接口：

```http
GET  /api/v1/system/info
POST /api/v1/auth/login
GET  /api/v1/auth/me
```

功能：

- 用户名或邮箱登录。
- 密码输入。
- 根据系统配置决定是否展示注册入口。
- 登录成功后跳转 redirect 或 `/bookshelf`。

状态：

- `loadingSystemInfo`
- `submitting`
- `errorMessage`
- `redirect`

验收：

- 错误账号展示明确错误。
- 禁用用户展示账号已禁用。
- 重复提交时按钮 loading 且 disabled。

### 11.2 注册页

路由：`/register`

接口：

```http
GET  /api/v1/system/info
POST /api/v1/auth/register
GET  /api/v1/auth/me
```

功能：

- 用户名、邮箱、昵称、密码、确认密码。
- 注册关闭时展示不可注册状态。
- 第一个用户注册成功且角色为 admin 时展示首个管理员提示。

验收：

- 两次密码不一致在前端拦截。
- 用户名、邮箱冲突展示后端 message。
- 注册成功后认证状态立即可用。

### 11.3 我的书架

路由：`/bookshelf`

接口：

```http
GET    /api/v1/bookshelf
GET    /api/v1/categories?scope=system
GET    /api/v1/tags?scope=system
DELETE /api/v1/bookshelf/:id
```

功能：

- 列表展示当前用户书架。
- 搜索、格式、分类、标签、来源、收藏筛选。
- 最近阅读、添加时间、标题排序。
- 上传入口和继续阅读入口。
- 移除图书前确认。
- 不可读图书展示状态。

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

视觉要求：

- 使用书卡网格。
- 每张卡必须显示格式、来源、进度和最近阅读时间。
- 空书架展示上传入口。
- 筛选无结果展示清除筛选入口。

### 11.4 图书详情

路由：`/bookshelf/:id`

接口：

```http
GET   /api/v1/bookshelf/:id
PATCH /api/v1/bookshelf/:id
GET   /api/v1/categories?scope=system
GET   /api/v1/tags?scope=system
GET   /api/v1/reader/books/:bookId/meta
```

功能：

- 展示图书元数据、分类、标签、来源、状态。
- 打开阅读器。
- 编辑个人标题、收藏、置顶、分类、标签。
- 返回书架时保留 query 状态。

验收：

- 403 展示无权访问。
- 404 展示图书不存在或已移除。
- `readable=false` 时禁用阅读入口。

### 11.5 上传页

路由：`/upload`

接口：

```http
GET  /api/v1/system/info
GET  /api/v1/categories?scope=system
GET  /api/v1/tags?scope=system
POST /api/v1/bookshelf/upload
POST /api/v1/library/books/upload
```

功能：

- 上传目标：个人私有书架或公共图书馆。
- 拖拽和点击选择文件。
- 客户端校验扩展名和大小。
- 元数据表单：标题、作者、简介、分类、标签。
- 上传进度展示。
- 上传成功后展示结果。

视觉要求：

- 上传目标使用并列选择卡。
- 拖拽区使用 2px 虚线边框。
- 拖入或聚焦时边框变为主题主色。
- 明确解释“个人私有书架”和“全站公共图书馆”的可见性差异。

业务规则：

- 支持 `epub`、`pdf`、`txt`。
- 文件大小不超过 `system.info.max_upload_size_mb`。
- 客户端校验只是提前提示，最终以后端校验为准。
- 公共图书上传成功后不自动加入个人书架。
- `pending` 表示待审核，不是上传中。

### 11.6 公共图书馆

路由：`/library`

接口：

```http
GET  /api/v1/library/books
GET  /api/v1/categories?scope=system
GET  /api/v1/tags?scope=system
POST /api/v1/library/books/:id/add-to-bookshelf
```

功能：

- 登录用户浏览 approved 公共图书。
- 搜索、格式、分类、标签筛选。
- 显示是否已加入书架。
- 加入书架。
- 上传公共图书入口。

视觉要求：

- 顶部常驻存储教育 Banner。
- Banner 文案必须传达：加入书架只建立引用，共享同一物理文件，不占用个人存储空间配额。
- 未加入显示主色“加入书架”按钮。
- 已加入显示弱化 disabled “已在书架”按钮。

验收：

- 重复点击加入不会重复提交。
- `book_already_in_bookshelf` 应更新本地加入状态。
- 加入成功后按钮立即变为已加入。

### 11.7 公共图书详情

路由：`/library/:id`

接口：

```http
GET  /api/v1/library/books/:id
POST /api/v1/library/books/:id/add-to-bookshelf
```

功能：

- 展示公共图书元数据。
- 展示加入状态。
- 加入书架。
- 上传者本人可查看自己 pending 的公共图书。

验收：

- 普通用户不能查看未 approved 的他人公共图书。
- 已加入后可跳转书架详情或阅读器。
- 返回公共图书馆时保留 query。

### 11.8 阅读器页

路由：`/reader/:bookId`

接口：

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

通用组件：

```text
ReaderView.vue
  ReaderToolbar.vue
  ReaderToc.vue
  ReaderSettingsPanel.vue
  ReaderProgressBar.vue
  ReaderBookmarkButton.vue
  EpubReader.vue
  PdfReader.vue
  TxtReader.vue
```

阅读设置：

- 主题：modern、sepia、dark。
- 字号：移动端默认 16px，桌面端默认 18px。
- 字号范围：12px 到 26px。
- 字号调整步长：1px 或 2px。
- 行高：1.5、1.8、2.2。
- 字体：sans、serif。

TXT：

- 不拉取完整 TXT。
- 先请求章节目录，再请求当前章节内容。
- 正文使用 `text-align: justify`。
- 后端返回 `content_type='text'` 时按纯文本段落渲染，注意转义。
- 后端返回 `content_type='html'` 时只渲染可信后端 HTML，必要时做白名单净化。

EPUB：

- 使用 `epub.js` 加载 `/api/v1/reader/books/:bookId/file`。
- 请求必须带 Authorization。
- 若 epub.js 资源请求无法稳定携带认证头，改为后端签发短期阅读 URL 或前端 fetch blob 后交给 epub.js。
- 页面目录优先展示后端章节。
- 进度使用 EPUB CFI，`progress_type='epub_cfi'`。

PDF：

- 使用 `pdf.js`。
- 文件来源为 `/api/v1/reader/books/:bookId/file`。
- 后端必须支持 Range。
- 若 pdf.js 无法自动携带认证头，使用 `httpHeaders` 注入 Authorization。
- 进度保存当前页码，`progress_type='pdf_page'`。

进度保存：

- 滚动停止后节流保存。
- 切换章节立即保存。
- 页面隐藏或组件卸载时尝试保存。
- 常规阅读中 5 到 10 秒最多保存一次。
- 后端保存失败时可把最近进度临时保存到 `localStorage`，下次进入提示恢复。

视觉要求：

- 桌面端左侧目录常驻。
- 移动端目录和设置为底部抽屉。
- 顶部工具栏固定。
- 正文宽度桌面端限制在舒适阅读宽度，建议 640px 到 760px。
- TXT/EPUB 正文必须两端对齐。
- 书签 icon 静默展示，激活后黄色高亮并 Toast 提示云端保存。

### 11.9 设置页

路由：`/settings`

接口：

```http
GET   /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password
```

功能：

- 查看当前用户资料。
- 修改昵称、邮箱。
- 修改密码。
- 查看角色、状态、存储配额和已用空间。

验收：

- 修改资料成功后同步 auth store。
- 修改密码成功后按后端策略重新登录或刷新认证状态。

### 11.10 管理首页

路由：`/admin`

接口：

```http
GET /api/v1/admin/system/storage
GET /api/v1/admin/system/settings
GET /api/v1/admin/library/books?page=1&page_size=5
GET /api/v1/admin/users?page=1&page_size=5
```

功能：

- 展示存储概览。
- 展示待审核公共图书。
- 展示最近用户。
- 提供用户管理、公共图书审核、分类标签、系统设置入口。

视觉要求：

- 指标卡使用大数字。
- 待审核数使用橙色状态灯。
- 页面是工作台，不做营销式展示。

### 11.11 管理员用户页

路由：`/admin/users`

接口：

```http
GET   /api/v1/admin/users
GET   /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id
PATCH /api/v1/admin/users/:id/status
PATCH /api/v1/admin/users/:id/role
```

功能：

- 用户列表、搜索、筛选。
- 查看用户详情。
- 禁用或启用用户。
- 调整角色。
- 调整存储配额。

验收：

- 修改状态、角色、配额前确认。
- 管理员不能禁用自己时展示后端错误。
- 操作后只刷新受影响行或当前页。

### 11.12 管理员图书页

路由：`/admin/books`

第一版可以跳转或内嵌 `/admin/library`，后续后端增加私有图书管理接口后再扩展全站图书管理。

### 11.13 管理员公共图书页

路由：`/admin/library`

接口：

```http
GET    /api/v1/admin/library/books
PATCH  /api/v1/admin/library/books/:id/status
DELETE /api/v1/admin/library/books/:id
DELETE /api/v1/admin/library/books/:id?delete_file=true
```

功能：

- 查看 pending、approved、rejected、hidden、deleted 公共图书。
- 审核通过。
- 审核拒绝。
- 隐藏。
- 恢复。
- 删除公共图书。
- 删除时选择是否物理删除文件。

验收：

- 物理删除必须二次确认。
- 默认删除为软删除。
- 操作成功后刷新当前行状态。

### 11.14 管理员分类页

路由：`/admin/categories`

接口：

```http
GET    /api/v1/categories?scope=system
POST   /api/v1/admin/categories
PATCH  /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

功能：

- 分类列表。
- 新增分类。
- 修改名称、slug、描述、父级。
- 删除未使用分类。

视觉要求：

- 分类可用列表或树形结构。
- 删除操作要有红色过渡和确认。

### 11.15 管理员标签页

路由：`/admin/tags`

接口：

```http
GET    /api/v1/tags?scope=system
POST   /api/v1/admin/tags
PATCH  /api/v1/admin/tags/:id
DELETE /api/v1/admin/tags/:id
```

功能：

- 标签列表。
- 新增标签。
- 修改名称和 slug。
- 删除未使用标签。

视觉要求：

- 标签使用可删除 Chip。
- 点击删除时有明显红色危险态。

### 11.16 管理员系统设置页

路由：`/admin/settings`

接口：

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

验收：

- 保存成功后刷新 system store。
- 超出后端限制时展示后端 message。

### 11.17 Not Found 页面

路由：`/:pathMatch(.*)*`

功能：

- 展示页面不存在。
- 已登录用户提供返回书架入口。
- 匿名用户提供返回登录入口。

## 12. 公共图书馆引用模型的前端表达

公共图书馆必须表达清楚：加入书架只建立引用，不复制文件。

前端表现：

- `LibraryView` 顶部常驻说明 Banner。
- 书架卡片中公共来源显示「引自公共馆」。
- 公共图书详情页显示“加入后不占用个人存储空间”。
- 已加入公共图书的按钮必须 disabled。
- 从书架移除公共图书时，确认文案说明只移除个人入口，不删除公共文件。

禁止表现：

- 不要把“加入书架”描述为“下载”。
- 不要暗示会复制一份文件到个人空间。
- 不要在公共上传成功后自动加入个人书架。

## 13. 权限处理

规则：

- 前端根据用户角色控制入口展示，但安全边界以后端为准。
- 普通用户不显示管理后台入口。
- 管理员入口需要后端接口再次校验。
- 403 统一展示无权限状态。
- 401 尝试 refresh，失败后跳转登录。
- 公共图书馆仅登录用户可访问。
- 匿名用户只能访问登录、注册和系统信息所需页面。

组件级权限：

- `RequireAdmin` 或路由 meta 控制管理员页面。
- 删除、审核、禁用、物理删除等按钮只在有权限时显示。
- 如果后端仍返回 403，页面必须展示错误，不可静默失败。

## 14. 数据加载策略

列表页面：

- 进入页面加载第一页。
- 筛选变化后重置到第一页。
- 搜索输入 debounce，建议 300ms。
- 分页、排序、筛选同步到 URL query。
- 列表请求失败保留筛选条件。

详情页面：

- 根据路由 id 加载详情。
- 404 展示不存在或已删除。
- 403 展示无权限。

阅读器页面：

- 先加载 meta。
- 再加载 chapters 和 progress。
- 当前章节内容加载失败允许重试。
- 可预取下一章内容。
- 离开页面前保存最近进度。

上传页面：

- 进入页面加载 system info、分类、标签。
- 选择文件后立即校验扩展名和大小。
- 上传中禁止切换目标和重复提交。

## 15. 本地持久化

可持久化：

- Access Token，建议 `sessionStorage`。
- 阅读设置：主题、字号、行高、字体。
- 最近一次未成功上传的表单草稿，不包含文件对象。
- 阅读进度保存失败时的临时兜底。

不可持久化：

- Refresh Token。
- 用户密码。
- 大段章节内容。
- 图书文件 Blob。
- 管理后台敏感操作草稿。

## 16. 可访问性和交互底线

所有页面必须满足：

- 表单字段有明确 label。
- 关键按钮有 loading 和 disabled 状态。
- 错误提示能被用户看到。
- 需要确认的危险操作不能直接执行。
- 移动端能完成浏览、加入书架、阅读、目录跳转和设置调整。
- 阅读器支持键盘基本操作：上一章、下一章或翻页。
- Toast 不作为唯一错误信息来源，关键错误必须留在页面中。
- 文字不得溢出按钮、卡片、表格单元格。

## 17. 测试计划

单元测试建议覆盖：

- API client token 注入。
- 401 refresh 队列。
- 分页参数构造。
- query 同步。
- 阅读进度节流。
- TXT 章节内容加载状态。
- 上传文件校验。

组件测试建议覆盖：

- 登录表单。
- 注册表单。
- 上传表单校验。
- 书架筛选。
- 公共图书加入书架按钮状态。
- 阅读器目录切换。
- 阅读器设置面板。

端到端测试建议覆盖：

- 首个用户注册成为管理员。
- 登录和刷新登录状态。
- 上传私有图书。
- 打开阅读器并保存进度。
- 上传公共图书后显示 pending。
- 管理员审核公共图书。
- 普通用户加入公共图书到书架。
- 管理员维护分类和标签。
- 管理员调整用户存储配额。

视觉验收：

- 对照 `frontend-style-demo.html` 检查桌面侧栏、移动导航抽屉、阅读器底部目录抽屉、三套主题。
- 检查 375px、768px、1024px、1440px 宽度下文本不重叠、不溢出。
- 检查深色主题下边框、文字和按钮对比度。
- 检查 Sepia 主题下阅读器字体和背景是否切换。

## 18. 开发阶段

### 阶段一：前端骨架

交付：

- Vite + Vue 3 + TypeScript 项目。
- Router。
- Pinia。
- Axios client。
- Naive UI Provider。
- 基础布局。
- 主题变量。
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
- 引用模型视觉说明。

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
- 响应式视觉验收。

## 19. 构建和部署

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
- 构建产物复制到 Go 后端静态资源目录。
- Go 后端最终托管前端静态资源。
- 浏览器访问 `http://server:8080`，API 和前端同源。

前端部署要求：

- 路由 history fallback 必须由后端静态资源服务支持。
- API 请求默认走 `/api/v1`。
- 不额外依赖 Nginx。
- 不要求 Redis、Elasticsearch、MQ 或对象存储。

## 20. 已确认实现决策

1. 前端不实现书籍源文件解析规则。
2. TXT、EPUB、PDF 的章节和内容规则以后端为准。
3. TXT 阅读优先按章节请求内容，不拉取整本书。
4. EPUB 第一期可用 epub.js 读取鉴权文件流。
5. PDF 使用 pdf.js，后端文件接口必须支持 Range。
6. Refresh Token 使用 HttpOnly Cookie，前端只管理 Access Token。
7. 公共图书馆不允许匿名访问。
8. 公共图书加入书架只建立引用，不复制文件。
9. 公共图书上传成功后不自动加入个人书架。
10. 公共图书上传后的 `pending` 是待审核状态，不是上传中。
11. 分类和标签由管理员维护，普通用户只能选择已有项。
12. 管理员可调整用户存储配额。
13. 正式 UI 使用 Naive UI，视觉参考 `frontend-style-demo.html`。
14. 三套主题 modern、sepia、dark 必须在阅读器和主布局中生效。
15. 后续开发若改变路由、接口、视觉规范或状态模型，必须先更新本文档。
