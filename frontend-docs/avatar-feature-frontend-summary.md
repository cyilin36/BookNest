# 用户头像功能前端实现总结

## 实现日期
2026年7月3日

## 功能概述
为 BookNest 前端实现用户头像功能，包括头像上传、默认头像选择、头像显示和个人资料管理。

## 前端变更

### 1. API 类型定义 (frontend/src/api/types.ts)

更新 `User` 接口，添加 `avatar_url` 字段：

```typescript
export interface User {
  id: number
  username: string
  email: string | null
  nickname: string | null
  avatar_url: string | null  // 新增
  role: UserRole
  status: UserStatus
  storage_quota_bytes: number | null
  storage_used_bytes: number
  created_at?: string
  last_login_at?: string | null
}
```

### 2. API 请求封装 (frontend/src/api/user.ts)

新增接口请求函数：

```typescript
export interface UpdateMeRequest {
  email?: string | null
  nickname?: string | null
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface SetDefaultAvatarRequest {
  avatar_name: string
}

export const userApi = {
  me() {
    return unwrap<User>(apiClient.get('/users/me'))
  },
  updateMe(payload: UpdateMeRequest) {
    return unwrap<User>(apiClient.patch('/users/me', payload))
  },
  changePassword(payload: ChangePasswordRequest) {
    return unwrap<Record<string, never>>(apiClient.patch('/users/me/password', payload))
  },
  uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return unwrap<User>(apiClient.post('/users/me/avatar/upload', formData, { 
      headers: { 'Content-Type': 'multipart/form-data' } 
    }))
  },
  setDefaultAvatar(payload: SetDefaultAvatarRequest) {
    return unwrap<User>(apiClient.post('/users/me/avatar/default', payload))
  }
}
```

### 3. 状态管理 (frontend/src/stores/auth.ts)

新增 `updateUser` 方法用于同步用户信息：

```typescript
actions: {
  // ...existing actions
  updateUser(user: User) {
    this.user = user
  }
}
```

### 4. 新增个人资料页面 (frontend/src/views/settings/ProfileView.vue)

功能特性：

- **头像管理**
  - 显示当前头像（120x120 圆形）
  - 上传自定义头像（支持 PNG/JPEG/WebP/GIF，最大 5MB）
  - 选择默认头像（6 个预设选项：default1 至 default6）
  - 前端文件类型和大小验证，显示明确错误提示
  - 头像上传后自动同步到 auth store 和导航栏

- **基本信息编辑**
  - 显示用户名（只读）
  - 编辑邮箱（可选）
  - 编辑昵称（可选）
  - 保存后更新本地状态

- **存储空间显示**
  - 显示已用空间 / 配额
  - 进度条可视化占用百分比
  - 字节数格式化为易读格式（B/KB/MB/GB）

- **密码修改**
  - 弹窗式密码修改对话框
  - 输入当前密码、新密码、确认新密码
  - 前端验证两次新密码一致性
  - 密码修改成功后自动退出登录（需重新登录）

### 5. 导航栏头像显示 (frontend/src/layouts/AppLayout.vue)

更新用户按钮显示逻辑：

```vue
<button class="user-button">
  <div v-if="auth.user?.avatar_url" class="user-avatar">
    <img :src="`${auth.user.avatar_url}?t=${Date.now()}`" alt="用户头像" class="avatar-image" />
  </div>
  <div v-else class="user-avatar">
    <User :size="20" />
  </div>
  <span class="user-name">{{ auth.user?.nickname || auth.user?.username }}</span>
</button>
```

样式更新：

```css
.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-round);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  overflow: hidden;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
```

用户下拉菜单新增"个人资料"入口：

```typescript
const userDropdownOptions = computed(() => [
  {
    label: auth.user?.role === 'admin' ? '管理员' : '普通用户',
    key: 'role',
    disabled: true
  },
  {
    type: 'divider',
    key: 'd1'
  },
  {
    label: '个人资料',
    key: 'profile',
    icon: () => h(User, { size: 18 })
  },
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(LogOut, { size: 18 })
  }
])
```

### 6. 路由配置 (frontend/src/router/index.ts)

新增个人资料路由：

```typescript
import ProfileView from '@/views/settings/ProfileView.vue'

// ...
{
  path: '/',
  component: AppLayout,
  meta: { requiresAuth: true },
  children: [
    // ...existing routes
    { path: 'profile', name: 'profile', component: ProfileView }
  ]
}
```

## UI/UX 设计

### 个人资料页面布局

- **容器**：最大宽度 680px，居中显示
- **卡片式设计**：圆角、阴影、内边距 48px
- **分节布局**：头像、基本信息、存储空间、安全设置用分隔线分开

### 头像预览

- **尺寸**：120x120px（桌面），100x100px（移动端）
- **样式**：圆形、边框
- **未设置状态**：灰色背景 + "未设置" 文字

### 默认头像选择

- **布局**：横向排列按钮，支持换行
- **按钮**：次要按钮样式，显示头像名称
- **加载状态**：上传中禁用所有按钮并显示加载动画

### 表单样式

- **标签位置**：顶部对齐
- **输入框**：全宽度
- **按钮**：主按钮（保存）+ 次要按钮（取消/重置）

### 密码修改对话框

- **类型**：模态对话框
- **表单布局**：垂直堆叠三个密码输入框
- **密码输入**：type="password"，隐藏输入内容
- **按钮**：取消 + 确认修改（主按钮）

## 前端验证规则

### 头像上传验证

```typescript
const avatarMaxBytes = 5 * 1024 * 1024
const avatarAllowedExtensions = ['png', 'jpg', 'jpeg', 'webp', 'gif']
const avatarAllowedTypes = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']
```

验证逻辑：

1. 检查文件扩展名是否在允许列表中
2. 检查 MIME 类型是否在允许列表中
3. 检查文件大小是否不超过 5MB
4. 验证失败显示具体错误消息和限制说明

### 密码修改验证

1. 当前密码、新密码、确认新密码均为必填
2. 新密码和确认新密码必须一致
3. 不一致时显示错误提示，不发送请求

## 用户交互流程

### 上传自定义头像

1. 用户进入个人资料页面
2. 点击"上传自定义头像"按钮
3. 选择图片文件
4. 前端验证格式和大小
5. 验证通过，显示上传中状态
6. 上传成功，更新头像预览和导航栏
7. 显示成功提示消息

### 选择默认头像

1. 用户进入个人资料页面
2. 点击任一默认头像按钮（如"默认头像 1"）
3. 显示加载状态
4. 设置成功，更新头像预览和导航栏
5. 显示成功提示消息

### 修改密码

1. 用户进入个人资料页面
2. 点击"修改密码"按钮
3. 弹出密码修改对话框
4. 输入当前密码、新密码、确认新密码
5. 点击"确认修改"
6. 前端验证两次密码一致
7. 提交到后端
8. 成功后显示提示并在 1.5 秒后自动退出登录
9. 跳转到登录页

## 错误处理

### 头像上传错误

- **avatar_format_not_supported**：显示"头像格式不支持，支持 PNG、JPEG、WebP、GIF，最大 5MB。"
- **avatar_too_large**：显示"头像过大，最大允许 5MB。当前文件约 X.XX MB。"
- **avatar_content_invalid**：显示后端错误消息 + 格式限制说明
- **网络错误**：显示"上传失败" + 格式限制说明

### 默认头像设置错误

- **invalid_avatar_name**：显示后端返回的错误消息
- **网络错误**：显示"设置失败"

### 密码修改错误

- **客户端验证失败**：显示"两次输入的新密码不一致"
- **old_password 错误**：显示后端返回的错误消息
- **网络错误**：显示"修改失败"

## 缓存策略

### 头像 URL 缓存破坏

在显示头像时添加时间戳参数：

```typescript
const avatarUrl = computed(() => {
  if (!profile.value?.avatar_url) return null
  return `${profile.value.avatar_url}?t=${Date.now()}`
})
```

导航栏同样添加时间戳：

```vue
<img :src="`${auth.user.avatar_url}?t=${Date.now()}`" />
```

这确保头像更新后浏览器立即重新加载最新图片。

## 响应式设计

### 桌面端（> 768px）

- 头像和操作按钮横向排列
- 头像 120x120px
- 默认头像按钮横向排列，自动换行

### 移动端（≤ 768px）

- 头像和操作按钮纵向堆叠，居中对齐
- 头像 100x100px
- 表单内边距减小
- 按钮全宽度

## 文档更新

### 前端开发文档 (frontend-docs/reader-frontend-development-doc.md)

1. **当前状态**部分：新增"用户头像功能"和"个人资料管理"实现说明
2. **目录结构**部分：新增 `ProfileView.vue`
3. **API 契约**部分：新增用户头像和个人资料接口列表
4. **状态管理**部分：auth store 新增 `updateUser` 方法说明
5. **页面说明**部分：新增"设置"章节，详细说明 `SettingsView` 和 `ProfileView` 功能

### API 契约文档 (reader-api-contract.md)

已由后端开发更新，前端直接参考以下接口：

- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`
- `PATCH /api/v1/users/me/password`
- `POST /api/v1/users/me/avatar/upload`
- `POST /api/v1/users/me/avatar/default`
- `GET /api/v1/users/:userId/avatar`

## 测试建议

### 功能测试

1. 访问 `/profile` 页面，验证能正常加载当前用户信息
2. 上传 PNG/JPEG/WebP/GIF 格式头像，验证成功并在导航栏显示
3. 选择 6 个不同的默认头像，验证都能成功设置
4. 上传头像后再选默认头像，验证自定义头像被替换
5. 选默认头像后再上传自定义头像，验证能正常上传
6. 修改邮箱和昵称，验证保存成功
7. 修改密码，验证成功后自动退出
8. 尝试上传 10MB 文件，验证前端拦截并显示错误
9. 尝试上传 TXT 文件，验证前端拦截并显示错误
10. 清空邮箱和昵称，验证能保存为空

### UI/UX 测试

1. 在桌面端和移动端测试响应式布局
2. 验证头像更新后导航栏立即刷新
3. 验证存储空间进度条正确显示
4. 验证密码修改对话框打开/关闭交互
5. 验证所有按钮的加载状态显示

### 边界测试

1. 上传恰好 5MB 的文件
2. 上传稍大于 5MB 的文件
3. 密码修改时输入不一致的新密码
4. 网络断开时尝试上传头像
5. 查看未设置头像的用户资料

## 后续优化方向

1. **头像裁剪**：上传前在前端提供裁剪和预览功能
2. **头像缩略图**：提供不同尺寸的头像 URL
3. **默认头像可视化**：在选择前显示默认头像的实际样式预览
4. **拖拽上传**：支持拖拽图片到头像区域上传
5. **头像历史**：保留最近几个头像版本，支持回退
6. **批量操作**：支持同时编辑多个资料字段
7. **实时验证**：输入邮箱时实时验证格式
8. **更好的错误提示**：针对不同错误码显示更友好的提示和解决方案
