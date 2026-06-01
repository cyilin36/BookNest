<script setup lang="ts">
import type { BookFormat } from '@/api/types'

defineProps<{ format: BookFormat }>()

const tagType = {
  epub: 'success',
  pdf: 'error',
  txt: 'warning'
} as const
</script>

<template>
  <n-tag size="small" :type="tagType[format]" round>{{ format.toUpperCase() }}</n-tag>
</template>
