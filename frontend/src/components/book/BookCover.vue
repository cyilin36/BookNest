<script setup lang="ts">
import { BookOpenText } from 'lucide-vue-next'

defineProps<{
  src?: string | null
  title: string
}>()
</script>

<template>
  <div class="book-cover">
    <img v-if="src" :src="src" :alt="title" />
    <div v-else class="cover-fallback">
      <BookOpenText :size="32" />
      <span>{{ title.slice(0, 12) }}</span>
    </div>
  </div>
</template>

<style scoped>
.book-cover {
  position: relative;
  aspect-ratio: 3 / 4;
  overflow: hidden;
  background: linear-gradient(145deg, rgba(24, 160, 88, 0.18), rgba(31, 41, 55, 0.08));
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.book-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  display: grid;
  height: 100%;
  place-items: center;
  align-content: center;
  gap: 10px;
  padding: 14px;
  color: var(--color-text-sec);
  text-align: center;
}

.cover-fallback span {
  color: var(--color-text-main);
  font-weight: 700;
  line-height: 1.35;
}
</style>
