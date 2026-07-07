<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import PageShell from '@/components/common/PageShell.vue'
import SiteBrandMark from '@/components/common/SiteBrandMark.vue'
import { adminApi } from '@/api/admin'
import { useSystemStore } from '@/stores/system'
import type { SystemSettings } from '@/api/types'

const message = useMessage()
const system = useSystemStore()
const loading = ref(false)
const saving = ref(false)
const iconUploading = ref(false)
const iconDeleting = ref(false)
const bgUploading = ref(false)
const bgDeleting = ref(false)
const iconInputRef = ref<HTMLInputElement | null>(null)
const bgInputRef = ref<HTMLInputElement | null>(null)
const iconMaxBytes = 2 * 1024 * 1024
const bgMaxBytes = 10 * 1024 * 1024
const iconAllowedExtensions = ['png', 'jpg', 'jpeg', 'webp', 'svg', 'ico']
const bgAllowedExtensions = ['png', 'jpg', 'jpeg', 'webp', 'gif']
const iconAllowedTypes = ['image/png', 'image/jpeg', 'image/webp', 'image/svg+xml', 'image/x-icon', 'image/vnd.microsoft.icon']
const bgAllowedTypes = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']
const iconLimitText = '仅支持 PNG、JPG、WEBP、SVG、ICO，最大 2MB。'
const bgLimitText = '仅支持 PNG、JPG、WEBP、GIF，最大 10MB。'
const form = reactive<SystemSettings>({
  site_name: 'BookNest',
  site_icon_url: null,
  login_background_url: null,
  allow_registration: true,
  library_review_required: false,
  max_upload_size_mb: 100,
  default_user_storage_quota_mb: 1024
})

async function loadSettings() {
  loading.value = true
  try {
    Object.assign(form, await adminApi.settings())
    system.applySystemInfo(form)
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    const { site_icon_url: _siteIconUrl, login_background_url: _loginBgUrl, ...payload } = form
    Object.assign(form, await adminApi.updateSettings(payload))
    system.applySystemInfo(form)
    message.success('系统设置已保存')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function uploadIcon(files: FileList | null) {
  const file = files?.[0]
  if (!file) return
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const typeAllowed = !file.type || iconAllowedTypes.includes(file.type)
  const extensionAllowed = iconAllowedExtensions.includes(extension)
  if (!extensionAllowed || !typeAllowed) {
    message.error(`站点图标格式不支持，${iconLimitText}`)
    if (iconInputRef.value) iconInputRef.value.value = ''
    return
  }
  if (file.size > iconMaxBytes) {
    message.error(`站点图标过大，最大允许 2MB。当前文件约 ${(file.size / 1024 / 1024).toFixed(2)}MB。`)
    if (iconInputRef.value) iconInputRef.value.value = ''
    return
  }
  iconUploading.value = true
  try {
    const result = await adminApi.uploadSystemIcon(file)
    form.site_icon_url = result.site_icon_url
    system.setSiteIconUrl(result.site_icon_url)
    message.success('站点图标已更新')
  } catch (error) {
    const reason = error instanceof Error ? error.message : '上传失败'
    message.error(`${reason}。${iconLimitText}`)
  } finally {
    iconUploading.value = false
    if (iconInputRef.value) iconInputRef.value.value = ''
  }
}

async function deleteIcon() {
  iconDeleting.value = true
  try {
    await adminApi.deleteSystemIcon()
    form.site_icon_url = null
    system.setSiteIconUrl(null)
    message.success('站点图标已删除')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除失败')
  } finally {
    iconDeleting.value = false
  }
}

async function uploadBackground(files: FileList | null) {
  const file = files?.[0]
  if (!file) return
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const typeAllowed = !file.type || bgAllowedTypes.includes(file.type)
  const extensionAllowed = bgAllowedExtensions.includes(extension)
  if (!extensionAllowed || !typeAllowed) {
    message.error(`登录页背景格式不支持，${bgLimitText}`)
    if (bgInputRef.value) bgInputRef.value.value = ''
    return
  }
  if (file.size > bgMaxBytes) {
    message.error(`登录页背景过大，最大允许 10MB。当前文件约 ${(file.size / 1024 / 1024).toFixed(2)}MB。`)
    if (bgInputRef.value) bgInputRef.value.value = ''
    return
  }
  bgUploading.value = true
  try {
    const result = await adminApi.uploadLoginBackground(file)
    form.login_background_url = result.login_background_url
    system.setLoginBackgroundUrl(result.login_background_url)
    message.success('登录页背景已更新')
  } catch (error) {
    const reason = error instanceof Error ? error.message : '上传失败'
    message.error(`${reason}。${bgLimitText}`)
  } finally {
    bgUploading.value = false
    if (bgInputRef.value) bgInputRef.value.value = ''
  }
}

async function deleteBackground() {
  bgDeleting.value = true
  try {
    await adminApi.deleteLoginBackground()
    form.login_background_url = null
    system.setLoginBackgroundUrl(null)
    message.success('登录页背景已删除')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除失败')
  } finally {
    bgDeleting.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <PageShell title="系统设置" subtitle="编辑站点名称、注册开关、图书馆审核和上传限制。">
    <n-spin :show="loading">
      <n-form class="surface settings-form" label-placement="left" label-width="150" @submit.prevent="saveSettings">
        <n-form-item label="站点名称">
          <n-input v-model:value="form.site_name" placeholder="BookNest" />
        </n-form-item>
        <n-form-item label="站点图标">
          <div class="icon-setting">
            <SiteBrandMark :size="56" />
            <div class="icon-actions">
              <input ref="iconInputRef" class="sr-only" type="file" accept=".png,.jpg,.jpeg,.webp,.svg,.ico,image/png,image/jpeg,image/webp,image/svg+xml,image/x-icon" @change="uploadIcon(($event.target as HTMLInputElement).files)" />
              <div class="toolbar">
                <n-button secondary :loading="iconUploading" @click="iconInputRef?.click()">上传图标</n-button>
                <n-button secondary type="error" :disabled="!form.site_icon_url" :loading="iconDeleting" @click="deleteIcon">删除图标</n-button>
              </div>
              <p>{{ iconLimitText }}</p>
            </div>
          </div>
        </n-form-item>
        <n-form-item label="登录页背景">
          <div class="bg-setting">
            <div v-if="form.login_background_url" class="bg-preview">
              <img :src="system.loginBackgroundSrc || ''" alt="登录页背景预览" />
            </div>
            <div v-else class="bg-preview bg-preview-empty">
              <span>未设置背景</span>
            </div>
            <div class="bg-actions">
              <input ref="bgInputRef" class="sr-only" type="file" accept=".png,.jpg,.jpeg,.webp,.gif,image/png,image/jpeg,image/webp,image/gif" @change="uploadBackground(($event.target as HTMLInputElement).files)" />
              <div class="toolbar">
                <n-button secondary :loading="bgUploading" @click="bgInputRef?.click()">上传背景</n-button>
                <n-button secondary type="error" :disabled="!form.login_background_url" :loading="bgDeleting" @click="deleteBackground">删除背景</n-button>
              </div>
              <p>{{ bgLimitText }}</p>
            </div>
          </div>
        </n-form-item>
        <n-form-item label="开放注册">
          <n-switch v-model:value="form.allow_registration" />
        </n-form-item>
        <n-form-item label="图书馆需要审核">
          <n-switch v-model:value="form.library_review_required" />
        </n-form-item>
        <n-form-item label="最大上传 MB">
          <n-input-number v-model:value="form.max_upload_size_mb" :min="1" />
        </n-form-item>
        <n-form-item label="默认用户配额 MB">
          <div class="quota-field">
            <n-input-number v-model:value="form.default_user_storage_quota_mb" :min="0" :step="1024" />
            <p>普通用户未设置专属配额时套用此默认值，只统计私人书籍。0 表示默认不限制。</p>
          </div>
        </n-form-item>
        <div class="toolbar">
          <n-button type="primary" attr-type="submit" :loading="saving">保存设置</n-button>
          <n-button secondary @click="loadSettings">重新加载</n-button>
        </div>
      </n-form>
    </n-spin>
  </PageShell>
</template>

<style scoped>
.settings-form {
  max-width: 720px;
  padding: 18px;
}

.quota-field {
  display: grid;
  gap: 6px;
  width: 100%;
}

.quota-field p {
  margin: 0;
  color: var(--color-text-sec);
  font-size: 13px;
  line-height: 1.6;
}

.icon-setting {
  display: flex;
  align-items: center;
  gap: 14px;
}

.icon-actions {
  display: grid;
  gap: 6px;
}

.icon-actions p {
  margin: 0;
  color: var(--color-text-sec);
  font-size: 13px;
}

.bg-setting {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  width: 100%;
}

.bg-preview {
  flex-shrink: 0;
  width: 160px;
  height: 90px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.bg-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.bg-preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-page);
  color: var(--color-text-sec);
  font-size: 13px;
}

.bg-actions {
  display: grid;
  gap: 6px;
  flex: 1;
}

.bg-actions p {
  margin: 0;
  color: var(--color-text-sec);
  font-size: 13px;
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
</style>
