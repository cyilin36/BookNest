<script setup lang="ts">
import PageShell from '@/components/common/PageShell.vue'
import TaxonomyManager from '@/components/admin/TaxonomyManager.vue'
import { categoryApi } from '@/api/category'
</script>

<template>
  <PageShell title="分类管理" subtitle="维护系统分类，供上传、图书馆和书架筛选使用。">
    <TaxonomyManager
      title="分类"
      description="分类适合表达题材、学科或业务分区。"
      :load-items="categoryApi.list"
      :create-item="categoryApi.create"
      :update-item="categoryApi.update"
      :remove-item="categoryApi.remove"
    />
  </PageShell>
</template>
