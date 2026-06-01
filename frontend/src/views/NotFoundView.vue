<script setup lang="ts">
import { useRouter } from 'vue-router'
import EmptyState from '@/components/common/EmptyState.vue'

const router = useRouter()
</script>

<template>
  <main class="not-found">
    <EmptyState title="页面不存在" description="这个地址没有对应的内容。">
      <n-button type="primary" @click="router.push('/bookshelf')">返回书架</n-button>
    </EmptyState>
  </main>
</template>

<style scoped>
.not-found {
  min-height: 100vh;
  padding: 24px;
  background: var(--color-bg-page);
}
</style>
