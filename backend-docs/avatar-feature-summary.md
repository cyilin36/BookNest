# 用户头像功能实现总结

## 实现日期
2026年7月3日

## 功能概述
为 BookNest 项目新增用户头像功能，支持默认头像选择和自定义头像上传。

## 数据库变更

### 新增迁移文件
- `000006_add_user_avatar.up.sql`：为 `users` 表添加 `avatar_path` 字段
- `000006_add_user_avatar.down.sql`：回滚迁移

### 字段说明
- `users.avatar_path` (VARCHAR(512), nullable)：存储头像文件相对路径
  - 默认头像：`default/default1.svg` 到 `default/default6.svg`
  - 自定义头像：`avatars/<user_id>/avatar-<uuid>.<ext>`

## 文件存储

### 默认头像
在 `backend/assets/default/` 目录下预置了 6 个简单的 SVG 头像：
- default1.svg - 粉色
- default2.svg - 绿色
- default3.svg - 蓝色
- default4.svg - 橙色
- default5.svg - 紫色
- default6.svg - 米黄色

### 自定义头像
- 存储位置：`backend/assets/avatars/<user_id>/avatar-<uuid>.<ext>`
- 支持格式：PNG、JPEG、WebP、GIF
- 大小限制：5 MB
- 文件验证：扩展名、Content-Type、魔数（magic number）三重校验

## API 接口

### 1. 上传自定义头像
```http
POST /api/v1/users/me/avatar/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

字段：
- file: 图片文件

响应：User 对象（包含 avatar_url）

错误码：
- avatar_format_not_supported: 不支持的格式
- avatar_content_invalid: 文件内容与声明格式不匹配
- avatar_too_large: 文件超过 5MB
```

### 2. 设置默认头像
```http
POST /api/v1/users/me/avatar/default
Content-Type: application/json
Authorization: Bearer <token>

请求体：
{
  "avatar_name": "default1"  // default1 到 default6
}

响应：User 对象（包含 avatar_url）

错误码：
- invalid_avatar_name: 无效的默认头像名称
```

### 3. 获取用户头像
```http
GET /api/v1/users/:userId/avatar
Authorization: Bearer <token>

响应：图片文件流
- Content-Type: 根据文件格式返回正确的 MIME 类型
- Cache-Control: public, max-age=3600
- 未设置头像时返回 404
```

## 代码变更

### 模型层 (internal/model/user.go)
- 新增 `AvatarPath *string` 字段（数据库字段，JSON 忽略）
- 新增 `AvatarURL *string` 字段（计算字段，不存数据库）

### 存储层 (internal/storage/local.go)
- 新增 `SaveUserAvatar()` 方法：保存用户上传的头像文件

### 应用层 (internal/app/app.go)
- 新增路由：
  - `POST /api/v1/users/me/avatar/upload`
  - `POST /api/v1/users/me/avatar/default`
  - `GET /api/v1/users/:userId/avatar`
  - `HEAD /api/v1/users/:userId/avatar`

- 新增处理函数：
  - `uploadAvatar()`: 处理自定义头像上传
  - `setDefaultAvatar()`: 设置默认头像
  - `userAvatar()`: 返回用户头像文件流

- 新增辅助函数：
  - `avatarURL()`: 生成头像 URL
  - `avatarFormat()`: 验证头像格式
  - `validAvatarBytes()`: 验证文件魔数

- 更新 `FindUserByID()`: 填充 `avatar_url` 字段

## 业务规则

### 上传自定义头像
1. 验证文件格式（扩展名 + Content-Type + 魔数）
2. 验证文件大小（≤ 5MB）
3. 保存文件到 `assets/avatars/<user_id>/avatar-<uuid>.<ext>`
4. 更新数据库 `avatar_path` 字段
5. 删除旧的自定义头像文件（默认头像不删除）

### 设置默认头像
1. 验证头像名称在允许列表中
2. 更新数据库 `avatar_path` 为 `default/<name>.svg`
3. 删除旧的自定义头像文件（默认头像不删除）

### 获取头像
1. 任何登录用户都可以获取任何用户的头像
2. 根据文件扩展名返回正确的 Content-Type
3. 设置缓存头 `Cache-Control: public, max-age=3600`
4. 未设置头像时返回 404

## 文档更新

### API 契约文档 (reader-api-contract.md)
- 更新 `User` 类型定义，添加 `avatar_url` 字段
- 新增"上传自定义头像"接口文档
- 新增"设置默认头像"接口文档
- 新增"获取用户头像"接口文档

### 后端开发文档 (backend-docs/reader-backend-development-doc.md)
- 更新第 6 节"数据库设计"，说明 `avatar_path` 字段
- 新增第 10.4 节"用户头像"，详细说明头像功能实现
- 更新第 11 节"文件存储和上传"，添加 `assets/avatars/` 和 `assets/default/` 目录
- 更新第 19 节"路由总表"，添加头像相关路由
- 更新第 20 节"权限矩阵"，添加头像相关权限

## 安全特性

1. **文件格式验证**：三重验证（扩展名、Content-Type、魔数）防止恶意文件上传
2. **大小限制**：5MB 硬限制，通过中间件和代码双重校验
3. **路径安全**：所有路径通过 `storage` 模块生成，防止路径穿越
4. **权限控制**：
   - 上传/设置头像：只能修改自己的头像
   - 查看头像：登录用户可以查看任何用户的头像
5. **文件隔离**：每个用户的自定义头像存储在独立目录
6. **旧文件清理**：上传新头像或设置默认头像时自动删除旧的自定义头像

## 测试建议

### 功能测试
1. 注册新用户，验证 `avatar_url` 为 `null`
2. 设置默认头像，验证返回正确的 `avatar_url`
3. 访问头像 URL，验证返回 SVG 文件
4. 上传自定义头像（PNG、JPEG、WebP、GIF），验证成功
5. 上传超大文件，验证返回 `avatar_too_large`
6. 上传不支持格式，验证返回 `avatar_format_not_supported`
7. 上传伪造格式文件（如改扩展名的 txt），验证返回 `avatar_content_invalid`
8. 切换不同默认头像，验证 URL 更新
9. 切换默认头像后再上传自定义头像，验证成功
10. 查看其他用户头像，验证可以访问

### 边界测试
1. 上传恰好 5MB 的文件
2. 上传稍大于 5MB 的文件
3. 尝试设置不存在的默认头像名称
4. 尝试访问不存在的用户头像
5. 未登录访问头像接口

### 性能测试
1. 并发上传多个头像
2. 验证缓存头是否正确设置

## 前端集成建议

1. 在用户设置页面展示当前头像
2. 提供 6 个默认头像的缩略图供选择
3. 提供文件上传控件，限制客户端文件类型
4. 上传前进行客户端文件大小验证
5. 显示上传进度
6. 在导航栏、评论区等位置展示用户头像
7. 头像使用 `<img>` 标签直接引用 `avatar_url`
8. 未设置头像时显示默认占位符（如首字母圆圈）

## 后续优化方向

1. **图片压缩**：服务端自动压缩和调整图片尺寸（如统一为 200x200）
2. **CDN 支持**：头像文件迁移到 CDN，提升加载速度
3. **头像裁剪**：前端上传前提供裁剪功能，确保正方形头像
4. **更多默认头像**：增加更多风格的默认头像选择
5. **头像审核**：对于公共社区功能，可能需要头像内容审核
6. **WebP 优化**：优先返回 WebP 格式以节省带宽
