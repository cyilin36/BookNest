<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import PageShell from '@/components/common/PageShell.vue'
import { userApi } from '@/api/user'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/api/types'

const message = useMessage()
const auth = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const avatarUploading = ref(false)
const passwordChanging = ref(false)
const avatarInputRef = ref<HTMLInputElement | null>(null)
const showPasswordDialog = ref(false)

const avatarMaxBytes = 5 * 1024 * 1024
const avatarAllowedExtensions = ['png', 'jpg', 'jpeg', 'webp', 'gif']
const avatarAllowedTypes = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']
const avatarLimitText = '支持 PNG、JPEG、WebP、GIF，最大 5MB。'

const defaultAvatars = [
  { name: 'default1', label: '默认头像 1' },
  { name: 'default2', label: '默认头像 2' },
  { name: 'default3', label: '默认头像 3' },
  { name: 'default4', label: '默认头像 4' },
  { name: 'default5', label: '默认头像 5' },
  { name: 'default6', label: '默认头像 6' }
]

const profile = ref<User | null>(null)
const editForm = ref({
  email: '',
  nickname: ''
})
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const avatarUrl = computed(() => {
  if (!profile.value?.avatar_url) return null
  return `${profile.value.avatar_url}?t=${Date.now()}`
})

const storagePercentage = computed(() => {
  if (!profile.value || !profile.value.storage_quota_bytes) return 0
  return Math.round((profile.value.storage_used_bytes / profile.value.storage_quota_bytes) * 100)
})

const storageUsedText = computed(() => {
  if (!profile.value) return ''
  return formatBytes(profile.value.storage_used_bytes)
})

const storageQuotaText = computed(() => {
  if (!profile.value || !profile.value.storage_quota_bytes) return '无限制'
  return formatBytes(profile.value.storage_quota_bytes)
})

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

async function loadProfile() {
  loading.value = true
  try {
    profile.value = await userApi.me()
    editForm.value.email = profile.value.email || ''
    editForm.value.nickname = profile.value.nickname || ''
  } finally {
    loading.value = false
  }
}

async function saveProfile() {
  saving.value = true
  try {
    profile.value = await userApi.updateMe({
      email: editForm.value.email || null,
      nickname: editForm.value.nickname || null
    })
    message.success('个人信息已保存')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function uploadAvatar(files: FileList | null) {
  const file = files?.[0]
  if (!file) return

  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const typeAllowed = !file.type || avatarAllowedTypes.includes(file.type)
  const extensionAllowed = avatarAllowedExtensions.includes(extension)

  if (!extensionAllowed || !typeAllowed) {
    message.error(`头像格式不支持，${avatarLimitText}`)
    if (avatarInputRef.value) avatarInputRef.value.value = ''
    return
  }

  if (file.size > avatarMaxBytes) {
    message.error(`头像过大，最大允许 5MB。当前文件约 ${(file.size / 1024 / 1024).toFixed(2)}MB。`)
    if (avatarInputRef.value) avatarInputRef.value.value = ''
    return
  }

  avatarUploading.value = true
  try {
    profile.value = await userApi.uploadAvatar(file)
    auth.updateUser(profile.value)
    message.success('头像已更新')
  } catch (error) {
    const reason = error instanceof Error ? error.message : '上传失败'
    message.error(`${reason}。${avatarLimitText}`)
  } finally {
    avatarUploading.value = false
    if (avatarInputRef.value) avatarInputRef.value.value = ''
  }
}

async function setDefaultAvatar(avatarName: string) {
  avatarUploading.value = true
  try {
    profile.value = await userApi.setDefaultAvatar({ avatar_name: avatarName })
    auth.updateUser(profile.value)
    message.success('已设置默认头像')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '设置失败')
  } finally {
    avatarUploading.value = false
  }
}

async function changePassword() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    message.error('两次输入的新密码不一致')
    return
  }

  passwordChanging.value = true
  try {
    await userApi.changePassword({
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    message.success('密码修改成功，请重新登录')
    showPasswordDialog.value = false
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
    setTimeout(() => {
      auth.logout()
    }, 1500)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '修改失败')
  } finally {
    passwordChanging.value = false
  }
}

function cancelPasswordChange() {
  showPasswordDialog.value = false
  passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
}

onMounted(loadProfile)
</script>

<template>
  <PageShell title="个人资料" subtitle="管理您的头像、昵称、邮箱和密码。">
    <n-spin :show="loading">
      <div class="profile-container">
        <div class="profile-card surface">
          <div class="profile-section">
            <div class="section-header">
              <h3 class="section-title">头像</h3>
              <p class="section-description">{{ avatarLimitText }}</p>
            </div>
            <div class="avatar-setting">
              <div class="avatar-preview">
                <img v-if="avatarUrl" :src="avatarUrl" alt="用户头像" class="avatar-image" />
                <div v-else class="avatar-placeholder">
                  <span>未设置</span>
                </div>
              </div>
              <div class="avatar-actions">
                <input ref="avatarInputRef" class="sr-only" type="file" accept=".png,.jpg,.jpeg,.webp,.gif,image/png,image/jpeg,image/webp,image/gif" @change="uploadAvatar(($event.target as HTMLInputElement).files)" />
                <div class="toolbar">
                  <n-button secondary :loading="avatarUploading" @click="avatarInputRef?.click()">上传自定义头像</n-button>
                </div>
                <div class="default-avatars">
                  <n-space>
                    <n-button
                      v-for="avatar in defaultAvatars"
                      :key="avatar.name"
                      secondary
                      size="small"
                      :loading="avatarUploading"
                      @click="setDefaultAvatar(avatar.name)"
                    >
                      {{ avatar.label }}
                    </n-button>
                  </n-space>
                </div>
              </div>
            </div>
          </div>

          <n-divider />

          <div class="profile-section">
            <div class="section-header">
              <h3 class="section-title">基本信息</h3>
            </div>
            <n-form label-placement="top" @submit.prevent="saveProfile">
              <n-form-item label="用户名">
                <n-input :value="profile?.username" disabled />
              </n-form-item>
              <n-form-item label="邮箱">
                <n-input v-model:value="editForm.email" placeholder="可选" />
              </n-form-item>
              <n-form-item label="昵称">
                <n-input v-model:value="editForm.nickname" placeholder="可选" />
              </n-form-item>
              <div class="toolbar">
                <n-button type="primary" attr-type="submit" :loading="saving">保存信息</n-button>
              </div>
            </n-form>
          </div>

          <n-divider />

          <div class="profile-section">
            <div class="section-header">
              <h3 class="section-title">存储空间</h3>
              <p class="section-description">{{ storageUsedText }} / {{ storageQuotaText }}</p>
            </div>
            <n-progress type="line" :percentage="storagePercentage" :show-indicator="false" />
          </div>

          <n-divider />

          <div class="profile-section">
            <div class="section-header">
              <h3 class="section-title">安全</h3>
            </div>
            <n-button secondary @click="showPasswordDialog = true">修改密码</n-button>
          </div>
        </div>
      </div>
    </n-spin>

    <n-modal v-model:show="showPasswordDialog" preset="dialog" title="修改密码">
      <n-form label-placement="top" @submit.prevent="changePassword">
        <n-form-item label="当前密码" required>
          <n-input v-model:value="passwordForm.old_password" type="password" placeholder="输入当前密码" />
        </n-form-item>
        <n-form-item label="新密码" required>
          <n-input v-model:value="passwordForm.new_password" type="password" placeholder="输入新密码" />
        </n-form-item>
        <n-form-item label="确认新密码" required>
          <n-input v-model:value="passwordForm.confirm_password" type="password" placeholder="再次输入新密码" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="cancelPasswordChange">取消</n-button>
          <n-button type="primary" :loading="passwordChanging" @click="changePassword">确认修改</n-button>
        </n-space>
      </template>
    </n-modal>
  </PageShell>
</template>

<style scoped>
.profile-container {
  max-width: 680px;
  margin: 0 auto;
}

.profile-card {
  padding: var(--spacing-3xl);
  border-radius: var(--radius-xlarge);
  box-shadow: var(--shadow-medium);
}

.profile-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.section-header {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.section-title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-main);
  letter-spacing: -0.01em;
}

.section-description {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
}

.avatar-setting {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-xl);
}

.avatar-preview {
  flex-shrink: 0;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--color-border);
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: var(--color-bg-page);
  color: var(--color-text-sec);
  font-size: var(--font-size-sm);
}

.avatar-actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  flex: 1;
}

.default-avatars {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 768px) {
  .profile-card {
    padding: var(--spacing-2xl);
  }

  .avatar-setting {
    flex-direction: column;
    align-items: center;
  }

  .avatar-preview {
    width: 100px;
    height: 100px;
  }

  .avatar-actions {
    width: 100%;
  }
}
</style>
