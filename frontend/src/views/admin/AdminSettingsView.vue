<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import PageShell from '@/components/common/PageShell.vue'
import { adminApi } from '@/api/admin'
import type { SystemSettings } from '@/api/types'

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const form = reactive<SystemSettings>({
  site_name: 'Book Reader',
  allow_registration: true,
  library_review_required: false,
  max_upload_size_mb: 100,
  default_user_storage_quota_mb: 1024
})

async function loadSettings() {
  loading.value = true
  try {
    Object.assign(form, await adminApi.settings())
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    Object.assign(form, await adminApi.updateSettings({ ...form }))
    message.success('系统设置已保存')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <PageShell title="系统设置" subtitle="编辑站点名称、注册开关、公共馆审核和上传限制。">
    <n-spin :show="loading">
      <n-form class="surface settings-form" label-placement="left" label-width="150" @submit.prevent="saveSettings">
        <n-form-item label="站点名称">
          <n-input v-model:value="form.site_name" placeholder="Book Reader" />
        </n-form-item>
        <n-form-item label="开放注册">
          <n-switch v-model:value="form.allow_registration" />
        </n-form-item>
        <n-form-item label="公共馆需要审核">
          <n-switch v-model:value="form.library_review_required" />
        </n-form-item>
        <n-form-item label="最大上传 MB">
          <n-input-number v-model:value="form.max_upload_size_mb" :min="1" />
        </n-form-item>
        <n-form-item label="默认配额 MB">
          <n-input-number v-model:value="form.default_user_storage_quota_mb" :min="0" />
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
</style>
