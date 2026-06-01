<script setup lang="ts">
import type { Pagination } from '@/api/types'

defineProps<{
  pagination: Pagination
}>()

defineEmits<{
  change: [page: number]
}>()
</script>

<template>
  <div v-if="pagination.total > pagination.page_size" class="pagination-bar">
    <n-pagination
      :page="pagination.page"
      :page-size="pagination.page_size"
      :item-count="pagination.total"
      @update:page="$emit('change', $event)"
    />
  </div>
</template>

<style scoped>
.pagination-bar {
  display: flex;
  justify-content: center;
  padding: 8px 0 4px;
}
</style>
