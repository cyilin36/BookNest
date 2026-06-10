<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { AlertCircle, BookOpen, ShieldCheck, Users } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import { adminApi } from '@/api/admin'
import type { AdminStorageStats, SystemSettings } from '@/api/types'

const storage = ref<AdminStorageStats | null>(null)
const settings = ref<SystemSettings | null>(null)

onMounted(async () => {
  storage.value = await adminApi.storage()
  settings.value = await adminApi.settings()
})
</script>

<template>
  <PageShell title="管理后台" subtitle="用户、公共图书、分类标签和系统设置">
    <div class="metrics">
      <div class="metric surface">
        <Users :size="20" />
        <strong>{{ storage?.total_books ?? 0 }}</strong>
        <span>总图书</span>
      </div>
      <div class="metric surface">
        <ShieldCheck :size="20" />
        <strong>{{ settings?.library_review_required ? '待审' : '直发' }}</strong>
        <span>图书馆审核</span>
      </div>
      <div class="metric surface">
        <AlertCircle :size="20" />
        <strong>{{ settings?.allow_registration ? '开放' : '关闭' }}</strong>
        <span>注册状态</span>
      </div>
      <div class="metric surface">
        <BookOpen :size="20" />
        <strong>{{ storage?.public_books ?? 0 }}</strong>
        <span>公共图书</span>
      </div>
    </div>
  </PageShell>
</template>

<style scoped>
.metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric {
  display: grid;
  gap: 6px;
  padding: 18px;
}

.metric strong {
  font-size: 34px;
  line-height: 1;
}

.metric span {
  color: var(--color-text-sec);
}

@media (max-width: 720px) {
  .metrics {
    grid-template-columns: 1fr;
  }
}
</style>
