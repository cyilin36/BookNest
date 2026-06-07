# 阅读器前端开发文档

本文档是本项目唯一的前端开发依据，由原 `reader-frontend-build.md` 合并而来，并按当前 `frontend/` 代码状态更新。

以后所有前端结构、接口消费、状态管理、视觉交互、构建运行和实现约束，都以本文档为准。若前端代码发生变化，必须同步更新本文档；若本文档与 `reader-api-contract.md` 冲突，以 API 契约为准，并先更新契约再改代码。

`frontend-style-demo.html` 和 `VISUAL_STYLE_GUIDE.md` 只作为视觉与交互参考，不是正式技术栈或源码结构约束。

## 1. 当前状态

正式前端位于：

```text
frontend/
```

当前已实现：

- 用户登录、注册、退出、认证恢复。
- 401 refresh 单次刷新与请求重放。
- 我的书架列表、筛选、详情、收藏、置顶、移除、继续阅读。
- 公共图书馆列表、筛选、详情、加入书架、公共上传入口、我的公共图书馆、上传者上架/下架自己的公共图书。
- 私有/公共图书上传。
- 分类和标签读取、上传选择、列表筛选。
- 阅读器基础能力：元数据、目录、章节正文、章节切换、阅读设置、进度保存。
- 管理后台：概览、用户管理、公共图书管理、分类管理、标签管理、系统设置。
- 三套主题：Modern、Sepia、Dark。
- 受保护封面图片前端鉴权加载。

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

本次封面修复没有变更接口，因此无需更新 `reader-api-contract.md`。

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
      reader/
        ReaderView.vue
      settings/
        SettingsView.vue
      upload/
        UploadView.vue
```

目录约定：

- `api/`：请求函数、请求类型、响应类型、错误归一化。
- `stores/`：跨页面状态，不承载复杂 UI 细节。
- `views/`：页面级组合和路由参数读取。
- `layouts/`：页面骨架、导航和响应式区域。
- `components/book/`：书卡、封面、格式标签等图书复用组件。
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
- 阅读字号
- 阅读行高
- 阅读字体

持久化：

```text
localStorage: book-reader.settings
```

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
- 图书馆标题行右上角提供“我的公共图书馆”图标入口，位于搜索图标左侧。
- 我的公共图书馆使用 `mine=true` 拉取当前用户自己上传的公共图书，显示已上架和已下架图书，并提供全部/已上架/已下架筛选。
- 上传者可在自己的公共图书详情页或我的公共图书馆页面上架/下架图书；下架只改为 `hidden`，上架改回 `approved`，均不删除文件、书架引用、阅读进度或书签。我的公共图书馆中只有已上架图书提供下载入口。

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

上传目标：

- `private` -> `POST /bookshelf/upload`
- `public` -> `POST /library/books/upload`

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
- 目录抽屉
- 打开目录时自动定位到当前阅读章节
- 目录中当前阅读章节使用颜色和边框高亮提示
- 上一章/下一章
- 切换章节后自动回到章节顶部
- 进入阅读器时按已保存阅读进度恢复章节和滚动位置
- 正文区域点击交互：左侧切上一章，中间呼出阅读菜单，右侧切下一章；阅读菜单已打开时，再点击正文区域只关闭菜单，不触发翻章。
- 主题切换
- 字号调整
- 行高调整
- 节流保存阅读进度

进度策略：

- 前端继续使用 `GET/PUT /api/v1/reader/books/:bookId/progress`。
- `progress_value` 兼容历史纯章节 ID 字符串；旧值只能恢复到章节顶部。
- 新保存的 `progress_value` 是 JSON 字符串：`{"chapterId":number,"scrollRatio":number}`。
- `chapterId` 用于恢复章节，`scrollRatio` 用于恢复页面滚动比例。
- `percentage` 根据章节序号和当前章节内滚动比例估算，用于书架进度条展示。

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
- 系统设置编辑

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

设计原则：

- 工具型页面保持紧凑、清晰、低装饰。
- 阅读器正文优先，减少视觉干扰。
- 书架和图书馆使用书卡展示。
- 书架和图书馆列表书卡采用竖向封面书架样式：封面为视觉主体，书名、作者、进度条和阅读百分比排列在封面下方。
- 书卡标题统一使用 `frontend/src/components/book/BookTitle.vue` 做两行测量截断和省略号生成，后续不要在各页面重复写 `line-clamp` 或手动省略号逻辑。
- 书架和图书馆列表书卡不显示 EPUB/TXT/PDF 格式标签；来源、上传归属等业务信息用图标标识，与移除等操作按钮并排，避免抬高卡片顶部信息。
- 书卡可用主操作按钮使用实心高对比样式；后端返回 `readable=false` 时，“阅读”按钮必须呈现明确禁用态。
- 窄屏和平板/窄桌面宽度下使用顶部栏和抽屉导航，避免固定侧边栏挤压内容区。
- 书架页和图书馆页不直接铺开完整筛选表单，使用右上角搜索图标和抽屉承载复杂筛选项。

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

- 修改接口字段、路径或语义时，先更新 `reader-api-contract.md`。
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
- 若只是前端内部实现变化且不影响 API，`reader-api-contract.md` 不需要更新。
- 若接口契约变化，必须同步更新 `reader-api-contract.md`。
