# 阅读器前端开发文档

本文档是本项目唯一的前端开发依据，由原 `reader-frontend-build.md` 合并而来，并按当前 `frontend/` 代码状态更新。

以后所有前端结构、接口消费、状态管理、视觉交互、构建运行和实现约束，都以本文档为准。若前端代码发生变化，必须同步更新本文档；若本文档与 `dev/reader-api-contract.md` 冲突，以 API 契约为准，并先更新契约再改代码。

`frontend-style-demo.html` 和 `VISUAL_STYLE_GUIDE.md` 只作为视觉与交互参考，不是正式技术栈或源码结构约束。

## 1. 当前状态

正式前端位于：

```text
frontend/
```

当前已实现：

- 用户登录、注册、退出、认证恢复。
- 401 refresh 单次刷新与请求重放。
- 我的书架列表、筛选、详情、收藏、置顶、移除、下载、图书信息编辑、封面编辑、继续阅读。
- 图书馆列表、筛选、详情、加入书架、下载、公共上传入口、我的图书馆、上传者上架/下架和编辑自己的公共图书信息与封面。
- 私有/公共图书上传，支持上传时附带封面。
- 分类和标签读取、上传选择、列表筛选。
- 阅读器基础能力：元数据、目录、章节正文、章节切换、阅读设置、进度保存。
- 管理后台：概览、用户管理、公共图书管理、分类管理、标签管理、系统设置（包括站点图标和登录页背景上传）。
- 三套主题：Modern、Sepia、Dark。
- 受保护封面图片前端鉴权加载。
- 登录页自定义背景图片支持。
- 用户头像功能：支持上传自定义头像（PNG/JPEG/WebP/GIF，最大 5MB）、选择默认头像（6 个预设：default1 至 default6），导航栏显示用户头像，头像更新后自动同步到导航栏。
- 个人资料管理：编辑昵称、邮箱，修改密码（成功后自动退出登录），查看存储空间占用和配额百分比。

当前未完成或后续增强：

- EPUB 仍主要消费后端章节内容接口，尚未接入完整 epub.js 翻页体验。
- PDF 尚未接入完整 pdf.js 渲染器，当前按后端章节/HTML 内容展示。
- 书签 UI 和书签管理接口尚未完整接入。
- 阅读器内 EPUB 正文图片资源仍需结合后端资源改写策略进一步完善。
- 管理后台高级编辑能力仍可继续增强，例如用户资料编辑、批量管理公共图书。

## 2. 技术栈

固定技术栈：

- Vue 3
- Vite
- TypeScript
- Vue Router
- Pinia
- Naive UI
- Axios
- @vueuse/core
- epubjs
- pdfjs-dist
- lucide-vue-next

当前 `frontend/package.json` 脚本：

```json
{
  "dev": "vite",
  "build": "vue-tsc --noEmit && vite build",
  "preview": "vite preview"
}
```

本地开发：

```bash
cd frontend
npm install
npm run dev
```

默认地址：

```text
http://localhost:5173
```

本地 Vite proxy：

```ts
server: {
  port: 5173,
  proxy: {
    '/api': 'http://localhost:8080'
  }
}
```

生产构建输出：

```text
frontend/dist
```

后端默认托管目录来自 `FRONTEND_DIST_DIR`，默认值为：

```text
./frontend/dist
```

## 3. 职责边界

前端负责：

- 页面布局、导航、表单、列表、筛选、分页和交互反馈。
- 认证状态维护、access token 注入、refresh 队列。
- 请求后端提供的元数据、目录、章节正文、文件流和图片资源。
- EPUB、PDF、TXT 的浏览器端展示体验。
- 阅读设置、本地 UI 偏好、阅读进度节流保存。
- 管理后台交互。
- 业务错误、权限错误、空状态和加载状态展示。

后端负责：

- 用户认证、角色权限、禁用状态。
- refresh token HttpOnly Cookie。
- 图书上传、格式校验、文件落盘、文件鉴权。
- 元数据、封面、章节目录和章节正文解析。
- TXT 编码识别、章节规则匹配和字节偏移读取。
- EPUB 目录解析、章节正文读取、资源图片读取。
- PDF 文件 Range 读取和章节/页段生成。
- 阅读进度、书签、分类、标签、系统设置持久化。

前端不得实现：

- TXT 章节正则解析。
- EPUB fragment 截取和源文件内部解析规则。
- PDF 分段解析。
- refresh token 存储。
- 大段章节内容持久化。
- 图书文件 Blob 长期持久化。

## 4. API 契约

API 基础路径：

```text
/api/v1
```

所有 API 类型统一维护在：

```text
frontend/src/api/types.ts
```

所有请求函数统一从：

```text
frontend/src/api/
```

页面组件不得直接拼写 Axios 请求 URL，除非是通过 API 层或专用组件封装后的内部实现。

系统信息和管理员系统设置依赖接口：

```http
GET /api/v1/system/info
GET /api/v1/system/login-background
GET /api/v1/admin/system/settings
PUT /api/v1/admin/system/settings
POST /api/v1/admin/system/icon
DELETE /api/v1/admin/system/icon
POST /api/v1/admin/system/login-background
DELETE /api/v1/admin/system/login-background
```

站点图标和登录页背景通过 `site_icon_url` 和 `login_background_url` 在 `GET /api/v1/system/info` 返回；管理员上传/删除图标和背景必须走独立接口，不能把这些 URL 作为系统设置修改字段提交。

用户头像和个人资料接口：

```http
GET /api/v1/users/me
PATCH /api/v1/users/me
PATCH /api/v1/users/me/password
POST /api/v1/users/me/avatar/upload
POST /api/v1/users/me/avatar/default
GET /api/v1/users/:userId/avatar
```

阅读器依赖接口：

```http
GET /api/v1/reader/books/:bookId/meta
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/file
GET /api/v1/reader/books/:bookId/cover
GET /api/v1/reader/books/:bookId/resources?href=...
GET /api/v1/reader/books/:bookId/progress
PUT /api/v1/reader/books/:bookId/progress
```

若接口契约未变化，前端内部展示或状态调整不需要更新 `dev/reader-api-contract.md`。

## 5. 目录结构

当前实际目录：

```text
frontend/
  index.html
  package.json
  package-lock.json
  tsconfig.json
  vite.config.ts
  src/
    App.vue
    env.d.ts
    main.ts
    api/
      admin.ts
      auth.ts
      bookshelf.ts
      category.ts
      client.ts
      library.ts
      reader.ts
      system.ts
      tag.ts
      types.ts
      user.ts
    assets/
      styles/
        app.css
        reader.css
        variables.css
    components/
      admin/
        TaxonomyManager.vue
      book/
        BookCard.vue
        BookCover.vue
        CoverUploadField.vue
        BookInfoEditModal.vue
        BookTitle.vue
        FormatTag.vue
      common/
        EmptyState.vue
        PageShell.vue
        PaginationBar.vue
    composables/
      useReaderProgress.ts
      useTaxonomyOptions.ts
      useUpload.ts
    layouts/
      AdminLayout.vue
      AppLayout.vue
      AuthLayout.vue
      ReaderLayout.vue
    router/
      guards.ts
      index.ts
    stores/
      auth.ts
      bookshelf.ts
      library.ts
      reader.ts
      settings.ts
      system.ts
    utils/
      download.ts
      format.ts
    views/
      NotFoundView.vue
      admin/
        AdminCategoriesView.vue
        AdminHomeView.vue
        AdminLibraryView.vue
        AdminSettingsView.vue
        AdminTagsView.vue
        AdminUsersView.vue
      auth/
        LoginView.vue
        RegisterView.vue
      bookshelf/
        BookshelfDetailView.vue
        BookshelfView.vue
      library/
        LibraryDetailView.vue
        LibraryView.vue
        MyLibraryView.vue
      reader/
        ReaderView.vue
      settings/
        SettingsView.vue
        ProfileView.vue
      upload/
        UploadView.vue
```

目录约定：

- `api/`：请求函数、请求类型、响应类型、错误归一化。
- `stores/`：跨页面状态，不承载复杂 UI 细节。
- `views/`：页面级组合和路由参数读取。
- `layouts/`：页面骨架、导航和响应式区域。
- `components/book/`：书卡、封面、封面上传、格式标签等图书复用组件。
- `components/common/`：空状态、分页、页面壳等基础组件。
- `components/admin/`：后台复用组件。
- `composables/`：上传、阅读进度、分类标签选项等复用交互逻辑。
- `assets/styles/`：全局样式、主题变量、阅读器样式。

## 6. 认证和 Axios

唯一 Axios 实例：

```text
frontend/src/api/client.ts
```

默认行为：

- `baseURL = import.meta.env.VITE_API_BASE || '/api/v1'`
- `withCredentials = true`
- access token 保存在 Pinia auth store 和 `sessionStorage`
- 请求拦截器添加 `Authorization: Bearer <token>`
- 响应拦截器统一处理后端错误

401 处理：

```text
API 返回 401
  |
若当前没有 refresh 请求，调用 /auth/refresh
  |
刷新成功，更新 access token
  |
重放原请求
  |
刷新失败，清空认证状态并跳转登录页
```

约束：

- 同一时间只允许一个 refresh 请求。
- 多个 401 请求等待同一个 refresh 结果。
- refresh 请求自身失败后不能无限重试。
- 退出登录必须清空本地认证状态。

## 7. 受保护图片资源

当前后端封面接口是登录保护接口：

```http
GET /api/v1/reader/books/:bookId/cover
```

浏览器原生 `<img src="/api/v1/...">` 请求不会自动携带 `Authorization: Bearer ...`。因此封面不能直接把受保护 URL 塞给 `<img>`。

当前实现位置：

```text
frontend/src/components/book/BookCover.vue
```

当前策略：

```text
cover_url
  |
BookCover.vue
  |
apiClient.get(..., { responseType: 'blob' })
  |
URL.createObjectURL(blob)
  |
<img :src="objectUrl">
```

效果：

- 书架卡片封面、书架详情封面、图书馆详情封面统一修复。
- 后端仍保持封面资源鉴权。
- 前端组件卸载或封面切换时释放 object URL。
- 封面更新接口可能返回与旧值相同的 `cover_url`，`BookCover.vue` 支持通过 `revision` 主动重新拉取 blob，避免更新成功后仍显示旧图。
- 上传页和图书信息编辑弹窗使用 `frontend/src/components/book/CoverUploadField.vue` 统一选择、校验和预览封面图片；预览已有受保护封面时复用 `BookCover.vue` 的鉴权加载策略。

后续如在阅读器正文中渲染 EPUB 内部图片，也需要同类策略，或者后端提供可安全访问的短期签名 URL。

## 8. 路由

当前路由：

```text
/
/login
/register
/bookshelf
/bookshelf/:id
/library
/library/mine
/library/:id
/upload
/settings
/reader/:bookId
/admin
/admin/users
/admin/books
/admin/library
/admin/categories
/admin/tags
/admin/settings
/:pathMatch(.*)*
```

守卫规则：

- `/login`、`/register` 匿名可访问。
- 登录用户访问登录/注册页时跳转 `/bookshelf` 或 redirect。
- 书架、图书馆、上传、设置、阅读器需要登录。
- `/admin/*` 需要管理员角色。
- 普通用户访问后台跳转 `/bookshelf?forbidden=1`。

认证恢复：

```text
App 启动
  |
auth.restore()
  |
若 sessionStorage 有 access token，优先 /auth/me
  |
失败则尝试 /auth/refresh
  |
成功进入目标页，失败进入匿名状态
```

## 9. 状态管理

### auth store

位置：

```text
frontend/src/stores/auth.ts
```

状态：

- `accessToken`
- `user`
- `authReady`
- `loading`
- `error`

动作：

- `register`
- `login`
- `fetchMe`
- `refresh`
- `restore`
- `logout`
- `clearAuth`
- `updateUser`：更新当前用户信息（用于上传头像后同步状态）

### system store

位置：

```text
frontend/src/stores/system.ts
```

用于读取匿名系统信息：

- 站点名称
- 是否开放注册
- 公共图书上传策略
- 支持格式
- 最大上传大小
- 默认用户配额

### bookshelf store

位置：

```text
frontend/src/stores/bookshelf.ts
```

负责：

- 书架列表
- 筛选条件
- 分页
- 收藏/置顶更新
- 移除书架项

### library store

位置：

```text
frontend/src/stores/library.ts
```

负责：

- 公共图书列表
- 筛选条件
- 分页
- 加入书架状态

### reader store

位置：

```text
frontend/src/stores/reader.ts
```

负责：

- 图书元数据
- 章节目录
- 当前章节
- 当前章节内容缓存
- 阅读进度

章节缓存规则：

- 只缓存当前访问过的少量章节。
- 缓存上限当前为 5 章。
- 不持久化大段章节内容。

### settings store

位置：

```text
frontend/src/stores/settings.ts
```

负责：

- 主题
- 阅读模式：滚动或分页
- 阅读字号
- 阅读行高
- 阅读宽度
- 阅读字体

持久化：

```text
localStorage: book-nest.settings
```

### system store

位置：

```text
frontend/src/stores/system.ts
```

负责：

- 站点名称
- 站点图标地址和缓存刷新版本
- 登录页背景地址和缓存刷新版本
- 注册开关
- 公共图书审核开关
- 支持上传格式
- 上传大小和默认用户配额

站点图标：

- 全站品牌图标使用 `frontend/src/components/common/SiteBrandMark.vue` 渲染。
- 登录页、侧边栏、移动端导航和系统设置预览均使用同一组件。
- 后端未配置图标或图标加载失败时，回退显示 `BN`。
- 上传或删除站点图标后，前端需要刷新 `system` store 并同步浏览器 favicon。
- 管理员上传站点图标前需要做前端校验，格式仅允许 PNG/JPG/WEBP/SVG/ICO，大小最大 2MB；格式或大小不符合时必须给明确错误提示。

登录页背景：

- 登录页背景使用 `AuthLayout.vue` 动态渲染。
- 后端已配置背景时，使用 `background-image` 覆盖整个登录页面（cover 模式）。
- 后端未配置背景时，显示默认浅蓝色渐变 `linear-gradient(135deg, #E0F2FE 0%, #BAE6FD 100%)`。
- 上传或删除登录页背景后，前端需要刷新 `system` store 更新背景 URL 和版本号。
- 管理员上传登录页背景前需要做前端校验，格式仅允许 PNG/JPG/WEBP/GIF，大小最大 10MB；格式或大小不符合时必须给明确错误提示。
- 背景图片通过 `GET /api/v1/system/login-background` 匿名访问（无需认证）。

## 10. 页面说明

### 登录和注册

位置：

```text
frontend/src/views/auth/
```

能力：

- 登录
- 注册
- 注册前校验密码至少 6 位，避免只展示后端泛化的校验失败提示。
- 登录后跳转 redirect 或 `/bookshelf`
- 注册后进入书架
- 动态背景：管理员配置的登录页背景图片或默认浅蓝色渐变

### 我的书架

位置：

```text
frontend/src/views/bookshelf/
```

能力：

- 书架列表
- 关键词搜索
- 格式筛选
- 来源筛选
- 分类筛选
- 标签筛选
- 收藏筛选
- 筛选条件不常驻展示，统一由右上角搜索图标打开抽屉操作。
- 书架页不显示副标题，搜索图标和 `+` 上传图标保持在标题行右侧，两个图标按钮使用一致的弱化样式。
- 收藏/置顶
- 移除；私人图书会删除文件和相关数据，公共图书只移除当前用户书架引用，失败时展示后端错误。
- 下载；书架项通过 `GET /api/v1/bookshelf/:id/download` 下载源文件，文件名以服务端 `Content-Disposition` 为准，前端仅提供展示书名加格式后缀的 fallback。后端返回 `readable=false` 的书架项不提供可用下载入口。
- 编辑图书信息；书架详情页通过 `PATCH /api/v1/bookshelf/:id` 编辑标题、作者、简介、个人分类、个人标签。私人上传图书会修改图书本体信息，公共图书引用只修改当前用户自己的书架展示信息；分类和标签始终作为当前书架项的个人分类/标签处理。
- 编辑封面；书架详情页的编辑弹窗通过 `CoverUploadField.vue` 选择封面图片，并通过 `PUT /api/v1/bookshelf/:id/cover` 上传。私人上传图书更新图书本体封面；公共图书引用更新当前书架项个人封面，只影响当前用户自己的书架展示，不修改图书馆公共封面，也不影响其他用户。封面更新后递增 `BookCover` 的 `revision` 以重新拉取可能同 URL 的受保护封面。
- 进入详情
- 继续阅读

### 图书馆

位置：

```text
frontend/src/views/library/
```

能力：

- 公共图书列表
- 关键词搜索
- 格式筛选
- 分类筛选
- 标签筛选
- 筛选条件不常驻展示，统一由右上角搜索图标打开抽屉操作。
- 图书馆页不显示副标题，搜索图标和 `+` 上传图标保持在标题行右侧，两个图标按钮使用一致的弱化样式。
- 加入书架
- 下载；仅 `library_status='approved'` 的公共图书通过 `GET /api/v1/library/books/:id/download` 下载源文件，已下架图书不可下载。
- 详情查看
- 公共上传入口
- 图书馆标题行右上角提供“我的图书馆”图标入口，位于搜索图标左侧。
- 我的图书馆使用 `mine=true` 拉取当前用户自己上传的公共图书，显示已上架和已下架图书，并提供全部/已上架/已下架筛选。
- 上传者可在自己的公共图书详情页或我的图书馆页面上架/下架图书；下架只改为 `hidden`，上架改回 `approved`，均不删除文件、书架引用、阅读进度或书签。我的图书馆中只有已上架图书提供下载入口。
- 上传者可在自己的公共图书详情页和我的图书馆列表编辑标题、作者、简介、公共分类、公共标签；通过 `PATCH /api/v1/library/books/:id` 修改公共图书本体和公共分类/标签，只影响图书馆展示、图书馆筛选和后续新加入书架的快照，不影响已经加入书架的图书展示或个人分类/标签。
- 上传者可在自己的公共图书详情页和“我的图书馆”的编辑信息弹窗中通过 `CoverUploadField.vue` 更改公共封面；通过 `PUT /api/v1/library/books/:id/cover` 修改图书馆公共封面，影响图书馆展示和后续新加入书架的默认封面，不影响已经在私人书架内设置过个人封面的书架项。封面更新后递增对应 `BookCover` 的 `revision` 以重新拉取可能同 URL 的受保护封面。

### 上传

位置：

```text
frontend/src/views/upload/UploadView.vue
```

能力：

- 私人图书
- 公共上传
- 拖拽/点击选择文件
- 标题、作者、简介
- 分类、标签
- 可选封面图片，支持 PNG、JPG、WebP，最大 5MB；前端通过 `CoverUploadField.vue` 统一校验并预览

上传目标：

- `private` -> `POST /bookshelf/upload`
- `public` -> `POST /library/books/upload`

上传字段：

- `file`：图书源文件。
- `cover`：可选封面文件。
- `title`、`author`、`description`、`category_ids`、`tag_ids`：可选元数据和分类标签。

### 设置

位置：

```text
frontend/src/views/settings/
```

能力：

- **SettingsView.vue**：阅读设置页面
  - 主题切换
  - 阅读模式切换（滚动/分页）
  - 字号调整
  - 行距调整
  - 阅读宽度调整
  - 恢复默认设置
  
- **ProfileView.vue**：个人资料页面
  - **头像管理**：显示当前头像（120x120px 圆形预览，桌面端；100x100px 移动端），上传自定义头像（点击"上传自定义头像"按钮触发文件选择，格式限制 PNG/JPEG/WebP/GIF，最大 5MB），选择默认头像（6 个按钮：default1 至 default6）。前端先校验文件扩展名、MIME 类型和大小，校验失败显示明确错误提示和格式限制说明。上传或选择成功后，递增 `avatarVersion` 响应式变量触发预览刷新（通过 `avatarUrl` computed 属性附加时间戳参数破坏缓存），同时调用 `auth.updateUser(profile.value)` 同步到 auth store，导航栏立即显示更新后头像。未设置头像时预览区域显示灰色占位符和"未设置"文字。
  - **基本信息编辑**：显示用户名（只读），编辑邮箱（可选），编辑昵称（可选）。通过 `PATCH /api/v1/users/me` 提交 `{email, nickname}`，字段值为空时传 `null`。保存成功后更新本地 `profile` 状态并显示成功消息。
  - **存储空间显示**：显示已用空间 / 配额（字节数格式化为 B/KB/MB/GB，通过 `formatBytes` 工具函数），进度条可视化占用百分比（基于 `storage_used_bytes / storage_quota_bytes * 100`）。配额为 null 时显示"无限制"。
  - **密码修改**：点击"修改密码"按钮弹出模态对话框，输入当前密码、新密码、确认新密码。前端先验证两次新密码一致性，不一致时显示"两次输入的新密码不一致"错误提示，不发送请求。通过 `PATCH /api/v1/users/me/password` 提交 `{old_password, new_password}`，成功后显示"密码修改成功，请重新登录"提示，并在 1.5 秒后自动调用 `auth.logout()` 清空认证状态并跳转登录页（用户需重新登录验证新密码）。

路由：

- `/settings`：阅读设置
- `/profile`：个人资料

### 阅读器

位置：

```text
frontend/src/views/reader/ReaderView.vue
```

当前能力：

- 加载 meta
- 加载 chapters
- 加载 progress
- 加载章节正文
- 阅读器页面不常驻显示书名或普通页面标题栏，减少正文上方干扰；当前章节名由正文或阅读控制栏承担展示。
- TXT 等纯文本章节的当前章节名渲染在正文区域开头，使用大号加粗标题；HTML/EPUB 章节优先使用后端返回正文中的标题，避免重复标题。阅读器正文顶部不再单独显示章节信息栏，也不显示 EPUB/TXT/PDF 等格式标签。
- 目录抽屉
- 打开目录时自动定位到当前阅读章节
- 目录中当前阅读章节使用颜色和边框高亮提示
- 上一章/下一章
- 切换章节后自动回到章节顶部
- 进入阅读器时按已保存阅读进度恢复章节和滚动位置
- 正文区域点击交互：滚动模式下左侧切上一章，中间呼出阅读控制栏，右侧切下一章；分页模式下左侧切上一页，右侧切下一页，在章节第一页继续向左翻会进入上一章最后一页，在章节末页继续向右翻会进入下一章第一页，中间仍呼出阅读控制栏；控制栏已打开时，再点击正文区域只关闭控制栏，不触发翻页或翻章。
- 分页模式跨章节切换必须延续同一套横向翻页动画：正文点击区按“连续翻页”处理，末页右翻落到下一章第一页，首页左翻落到上一章最后一页；跨章时临时渲染当前页和目标章节页组成的双页轨道并整体横向滑动，不能先滑空再瞬切内容。明确的“上一章/下一章”按钮按“章节跳转”处理，动画方向一致，但落点始终是目标章节第一页。
- `ReaderView.vue` 使用 `crossChapterPageTurn` 临时状态渲染跨章双页轨道；目标章节页数必须以覆盖层真实渲染后的分页宽度测量，确保上一页跨章直接落到上一章尾页。轨道先落到起点，再用浏览器原生动画完整播放横向 `transform` 翻页；动画完成后，底层真实章节在覆盖层下完成重算和目标页定位，再移除覆盖层并保存阅读进度，避免跨章时先露出目标章节开头再跳到尾页。
- 分页模式刷新或重新进入阅读器恢复进度时，必须先静默完成章节内容渲染、页数计算和目标页定位，并在定位期间禁用正文轨道的 `transform` 过渡，再显示正文与点击区，避免先露出章节开头或从其他页滑回保存页。
- 阅读器去除普通页面标题栏后，分页阅读面板需要按视口高度拉长，正文区域应填满视口内可用高度，只保留正常操作间距。
- 阅读控制栏采用贴边栏式显示，不使用卡片式浮层；顶部状态栏左侧为退出阅读，中间显示当前章节名，右侧为阅读设置。
- 正文下方的常驻章节按钮栏只在滚动模式显示；分页模式不渲染该常驻底部栏，正文区域直接填满视口内可用高度。
- 中间点击呼出的阅读控制栏在滚动模式和分页模式都保留，底部浮现菜单提供上一章、目录、下一章三个入口，目录位于上一章和下一章之间；阅读设置不在底部重复出现。
- 主题切换
- 字号调整
- 行高调整
- 阅读宽度调整；默认桌面端居中显示为约 760px，移动端默认铺满，可在阅读设置中选择窄、适中、宽或铺满。
- 分页阅读模式使用前端多栏排版计算页数，页数会随字号、行高、阅读宽度和设备尺寸变化重新计算；进度仍复用章节内比例保存，跨设备恢复为近似阅读位置。
- 节流保存阅读进度

进度策略：

- 前端继续使用 `GET/PUT /api/v1/reader/books/:bookId/progress`。
- `progress_value` 兼容历史纯章节 ID 字符串；旧值只能恢复到章节顶部。
- 新保存的 `progress_value` 是 JSON 字符串：`{"chapterId":number,"scrollRatio":number}`。
- `chapterId` 用于恢复章节，`scrollRatio` 用于恢复当前章节内比例；滚动模式表示页面滚动比例，分页模式表示当前页在章节页数中的比例。
- `percentage` 根据章节序号和当前章节内比例估算，用于书架进度条展示。

当前渲染策略：

- TXT/EPUB/PDF 均优先消费后端章节接口。
- 章节正文按后端返回的 `content_type` 分支渲染。
- `content_type='html'` 使用 `v-html` 展示 HTML 内容。
- `content_type='text'` 使用 `v-text` 纯文本展示，并通过 `white-space: pre-wrap` 保留 TXT 换行和段落空白。
- PDF 当前依赖后端生成的章节 HTML，不是完整 pdf.js 页面渲染。
- 阅读器正文全局样式必须防止横向撑破阅读区域：`.reader-content` 及其后代统一使用 `max-width: 100%`、`overflow-wrap: anywhere`、`word-break: break-word`；`pre/code` 使用 `pre-wrap`；`table` 限制在容器内横向滚动；图片和嵌入媒体不得超过正文宽度。

### 管理后台

位置：

```text
frontend/src/views/admin/
```

能力：

- 概览指标
- 用户筛选、角色调整、启用/禁用、删除普通用户账号
- 公共图书筛选、`approved`/`hidden` 状态流转、真正删除；筛选条件由右上角搜索图标打开抽屉操作，状态流转使用直接操作按钮，删除使用独立按钮。
- 分类新增、编辑、删除
- 标签新增、编辑、删除
- 系统设置编辑，包括站点名称、站点图标上传/删除、登录页背景上传/删除、注册开关、图书馆审核和上传限制。

## 11. 视觉和响应式

全局样式入口：

```text
frontend/src/assets/styles/app.css
```

主题变量：

```text
frontend/src/assets/styles/variables.css
```

阅读器样式：

```text
frontend/src/assets/styles/reader.css
```

当前主题：

- `theme-modern`
- `theme-sepia`
- `theme-dark`

### 设计系统（2024 精致化重构）

#### 主色调
- **主色**：#4A90E2（蓝色系）
- **主色悬浮**：#3A7BC8
- **主色浅色**：rgba(74, 144, 226, 0.15)

#### 背景色
- **页面背景**：#F8F9FA
- **卡片背景**：#FFFFFF
- **悬浮背景**：rgba(0, 0, 0, 0.04)

#### 文字色
- **主文字**：#1A1A1A
- **次要文字**：#6B6B6B
- **禁用文字**：#A0A0A0

#### 边框色
- **默认边框**：#E8EAED
- **悬浮边框**：rgba(74, 144, 226, 0.3)

#### 阴影系统
```css
--shadow-light: 0 1px 3px rgba(0, 0, 0, 0.08)
--shadow-medium: 0 2px 8px rgba(0, 0, 0, 0.12)
--shadow-heavy: 0 4px 16px rgba(0, 0, 0, 0.12)
--shadow-cover: 0 4px 12px rgba(0, 0, 0, 0.1)
--shadow-cover-hover: 0 8px 24px rgba(0, 0, 0, 0.18)
--shadow-button-primary: 0 2px 8px rgba(74, 144, 226, 0.25)
--shadow-button-primary-hover: 0 4px 16px rgba(74, 144, 226, 0.35)
```

#### 间距系统（8px 基准）
```css
--spacing-xs: 4px
--spacing-sm: 8px
--spacing-md: 12px
--spacing-lg: 16px
--spacing-xl: 20px
--spacing-2xl: 24px
--spacing-3xl: 32px
--spacing-4xl: 40px
--spacing-5xl: 48px
```

#### 圆角
```css
--radius-xlarge: 16px  /* 大型容器 */
--radius-large: 12px   /* 卡片、模态框 */
--radius-medium: 10px  /* 输入框、按钮 */
--radius-small: 8px    /* 小型元素 */
--radius-xsmall: 6px   /* 标签 */
--radius-round: 50%    /* 圆形 */
```

#### 字号系统
```css
--font-size-xs: 12px
--font-size-sm: 13px
--font-size-base: 14px
--font-size-md: 15px
--font-size-lg: 16px
--font-size-xl: 18px
--font-size-2xl: 20px
--font-size-3xl: 24px
```

#### 控件尺寸
- **Large 控件**：52px 高度（主要表单输入）
- **Medium 控件**：44px 高度（次要表单、普通按钮）
- **Small 控件**：32px 高度（紧凑布局）

#### 微交互动画
- **按钮悬浮**：`translateY(-2px)` + 增强阴影
- **卡片悬浮**：`translateY(-4px)` + 阴影变化
- **目录项悬浮**：`translateX(4px)` + 背景色变化
- **过渡时长**：200-300ms ease

#### Focus 状态
所有可交互元素的 focus 状态：
```css
box-shadow: 0 0 0 4px var(--color-primary-light);
```

### Naive UI 主题覆盖

在 `App.vue` 中通过 `GlobalThemeOverrides` 统一覆盖 Naive UI 组件样式：

```typescript
const themeOverrides = {
  common: {
    primaryColor: '#4A90E2',
    primaryColorHover: '#3A7BC8',
    primaryColorPressed: '#2E6BAC',
    borderRadius: '12px'
  },
  Button: {
    heightMedium: '44px',
    heightLarge: '52px',
    borderRadiusMedium: '10px',
    borderRadiusLarge: '12px'
  },
  Input: {
    heightLarge: '52px',
    borderRadius: '12px'
  }
  // ...
}
```

### 页面级布局规范

#### 认证页面（AuthLayout）
- 渐变紫色背景：`linear-gradient(135deg, #667eea 0%, #764ba2 100%)`
- 中央白色卡片：480px 宽，40px 内边距，16px 圆角
- 大阴影：`0 20px 60px rgba(0, 0, 0, 0.3)`
- Brand 区域底部边框分隔

#### 主应用页面（AppLayout）
- 顶部水平导航栏：64px 高度（桌面），56px（移动）
- 导航项 48px 间距
- 激活状态：蓝色背景 + 白色文字

#### 阅读器（ReaderLayout）
- 全屏沉浸式布局
- 顶部状态栏：透明背景 + 毛玻璃效果（`backdrop-filter: blur(12px)`）
- 底部菜单：3 列均分按钮网格

#### 通用内容页（PageShell）
- 页面标题：32px 字号，-0.03em 字距
- 操作按钮组：右对齐，12px 间距
- 内容区域：32px 上边距

### 组件级视觉规范

#### 书籍卡片（BookCard）
- 封面圆角：12px
- 封面阴影：默认 medium，悬浮 cover-hover
- 悬浮效果：`translateY(-4px)` + 增强阴影
- 进度条：4px 高度，底部覆盖，蓝色填充
- 来源标识：右上角 28px 圆形，毛玻璃背景

#### 列表项（BookListItem）
- 横向布局：80×120px 封面 + 内容区
- 悬浮效果：`translateY(-2px)` + 阴影提升
- Meta 标签：3px 10px 内边距，6px 圆角，大写

#### 模态框（BookInfoEditModal）
- 宽度：580px（最大 calc(100vw - 32px)）
- 输入框：large 尺寸，12px 圆角
- 按钮：large 尺寸，最小宽度 100px

#### 空状态（EmptyState）
- 图标圆形背景：80px，蓝色，0.8 透明度
- 标题：20px semibold
- 描述：15px 次要色
- 操作按钮居中

#### Reader 抽屉
- 目录列表：无边框，背景色区分
- 激活章节：白色文字 + 蓝色背景 + 轻阴影
- 悬浮动画：`translateX(4px)`
- 设置面板：32px 垂直间距

### 响应式断点

```css
/* 主要断点 */
@media (max-width: 768px) {
  /* 平板和移动设备 */
}

@media (max-width: 520px) {
  /* 移动设备优化 */
}

/* 容器最大宽度 */
.container {
  max-width: 1200px;
  margin: 0 auto;
}

/* 详情页 */
.detail-container {
  max-width: 960px;
}

/* 设置页 */
.settings-container {
  max-width: 680px;
}

/* 上传页 */
.upload-container {
  max-width: 800px;
}
```

设计原则：

- 工具型页面保持紧凑、清晰、低装饰。
- 阅读器正文优先，减少视觉干扰。
- 书架和图书馆使用书卡展示。
- 书架和图书馆列表书卡采用竖向封面书架样式：封面为视觉主体，书名、作者、进度条和阅读百分比排列在封面下方。
- 书卡标题统一使用 `frontend/src/components/book/BookTitle.vue` 做两行测量截断和省略号生成，后续不要在各页面重复写 `line-clamp` 或手动省略号逻辑。
- 书架和图书馆列表书卡不显示 EPUB/TXT/PDF 格式标签；来源、上传归属等业务信息用图标标识，与移除等操作按钮并排，避免抬高卡片顶部信息。
- 书卡可用主操作按钮使用实心高对比样式；后端返回 `readable=false` 时，”阅读”按钮必须呈现明确禁用态。
- 窄屏和平板/窄桌面宽度下使用顶部栏和抽屉导航，避免固定侧边栏挤压内容区。
- 书架页和图书馆页不直接铺开完整筛选表单，使用右上角搜索图标和抽屉承载复杂筛选项。
- 所有表单输入优先使用 `size=”large”`，提供更好的点击目标和视觉层级。
- 所有悬浮交互必须有平滑过渡（200-300ms）。
- 所有主要操作按钮使用蓝色阴影增强视觉吸引力。

## 12. 构建和部署

本地开发：

```bash
cd frontend
npm run dev
```

生产构建：

```bash
cd frontend
npm run build
```

构建产物：

```text
frontend/dist
```

Go 后端静态托管：

- 非 `/api/` 路径由后端 `serveFrontend` 返回前端静态资源。
- 若 `frontend/dist/index.html` 不存在，后端返回一个简单的 backend running 页面。

Docker/统一部署要求：

- 前端构建产物应复制到后端配置的 `FRONTEND_DIST_DIR`。
- 生产环境 `VITE_API_BASE` 使用 `/api/v1`。
- 不在前端环境变量保存密钥。

## 13. 开发约束

- 修改接口字段、路径或语义时，先更新 `dev/reader-api-contract.md`。
- 修改前端结构、状态、页面、组件、构建方式时，同步更新本文档。
- 所有 API 请求从 `src/api/` 发出。
- 组件内不得重复定义后端响应类型。
- 认证图片、文件等受保护资源不能直接依赖裸 `<img src>` 或裸 URL 下载，除非接口不需要 Bearer token。
- 使用 `apiClient` 请求受保护资源时，注意释放 `URL.createObjectURL` 创建的对象 URL；图书下载统一使用 `frontend/src/utils/download.ts` 解析 `Content-Disposition` 和触发浏览器下载。
- 不要把 refresh token、密码、图书文件 Blob、大段章节内容写入 localStorage。
- 不要在前端复刻后端解析规则。
- 前端页面应提供加载、空状态和错误反馈。
- 后台危险操作必须有确认弹窗。

## 14. 文档维护规则

本文档需要持续保持最新：

- 新增页面：更新目录结构、路由、页面说明。
- 新增 store：更新状态管理章节。
- 新增 API 模块：更新 API 章节。
- 修改认证、封面、文件、资源加载策略：更新认证和受保护资源章节。
- 修改构建命令、端口、代理、部署方式：更新构建和部署章节。
- 若只是前端内部实现变化且不影响 API，`dev/reader-api-contract.md` 不需要更新。
- 若接口契约变化，必须同步更新 `dev/reader-api-contract.md`。
