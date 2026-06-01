<script setup lang="ts">
import { BookOpen } from 'lucide-vue-next'

defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="empty-state surface">
    <BookOpen :size="34" />
    <h3>{{ title }}</h3>
    <p v-if="description">{{ description }}</p>
    <slot />
  </div>
</template>

<style scoped>
.empty-state {
  display: grid;
  place-items: center;
  min-height: 220px;
  padding: 30px;
  text-align: center;
  color: var(--color-text-sec);
}

.empty-state h3 {
  margin: 12px 0 4px;
  color: var(--color-text-main);
  font-size: 18px;
}

.empty-state p {
  margin: 0 0 16px;
}
</style>
