<script setup lang="ts">
import PageShell from '@/components/common/PageShell.vue'
import TaxonomyManager from '@/components/admin/TaxonomyManager.vue'
import { tagApi } from '@/api/tag'
</script>

<template>
  <PageShell title="标签管理" subtitle="维护系统标签，供上传和书架个性化标记使用。">
    <TaxonomyManager
      title="标签"
      description="标签适合表达短主题、状态或阅读偏好。"
      :load-items="tagApi.list"
      :create-item="tagApi.create"
      :update-item="tagApi.update"
      :remove-item="tagApi.remove"
    />
  </PageShell>
</template>
