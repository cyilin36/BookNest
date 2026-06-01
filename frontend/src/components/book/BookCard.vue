<script setup lang="ts">
import { computed } from 'vue'
import { BookMarked, Library, Pin, Star } from 'lucide-vue-next'
import BookCover from './BookCover.vue'
import FormatTag from './FormatTag.vue'
import type { BookshelfItem, LibraryBook } from '@/api/types'

const props = defineProps<{
  book: BookshelfItem | LibraryBook
  mode: 'bookshelf' | 'library'
}>()

const emit = defineEmits<{
  open: []
  read: []
  join: []
  remove: []
  toggleFavorite: []
  togglePinned: []
}>()

const progress = computed(() => {
  if ('progress_percentage' in props.book) return Math.max(0, Math.min(100, props.book.progress_percentage || 0))
  return 0
})

const isShelf = computed(() => props.mode === 'bookshelf' && 'source_type' in props.book)
</script>

<template>
  <article class="book-card surface">
    <button class="cover-button" type="button" @click="emit('open')">
      <BookCover :src="book.cover_url" :title="book.title" />
    </button>
    <div class="book-info">
      <div class="book-tags">
        <FormatTag :format="book.format" />
        <n-tag v-if="isShelf && 'source_type' in book" size="small" :type="book.source_type === 'uploaded' ? 'info' : 'default'" round>
          <template #icon>
            <BookMarked v-if="book.source_type === 'uploaded'" :size="13" />
            <Library v-else :size="13" />
          </template>
          {{ book.source_type === 'uploaded' ? '私有上传' : '引自公共馆' }}
        </n-tag>
      </div>
      <button class="book-title" type="button" @click="emit('open')">{{ book.title }}</button>
      <div class="book-author">{{ book.author || '未知作者' }}</div>
      <p v-if="'description' in book && book.description" class="book-description">{{ book.description }}</p>
      <div v-if="mode === 'bookshelf'" class="progress-line" aria-hidden="true">
        <span :style="{ width: `${progress}%` }" />
      </div>
      <div class="book-actions">
        <template v-if="mode === 'bookshelf' && 'readable' in book">
          <n-button class="book-primary-action book-read-action" size="small" type="primary" :disabled="!book.readable" @click="emit('read')">阅读</n-button>
          <n-button size="small" quaternary circle @click="emit('toggleFavorite')">
            <Star :size="16" :fill="book.favorite ? 'currentColor' : 'none'" />
          </n-button>
          <n-button size="small" quaternary circle @click="emit('togglePinned')">
            <Pin :size="16" :fill="book.pinned ? 'currentColor' : 'none'" />
          </n-button>
          <n-button size="small" quaternary type="error" @click="emit('remove')">移除</n-button>
        </template>
        <template v-else-if="mode === 'library' && 'in_bookshelf' in book">
          <n-button class="book-primary-action" size="small" type="primary" :disabled="book.in_bookshelf" @click="emit('join')">
            {{ book.in_bookshelf ? '已在书架' : '加入书架' }}
          </n-button>
          <n-button size="small" secondary @click="emit('open')">详情</n-button>
        </template>
      </div>
    </div>
  </article>
</template>

<style scoped>
.book-card {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
  min-height: 158px;
  padding: 12px;
}

.cover-button,
.book-title {
  padding: 0;
  color: inherit;
  background: transparent;
  border: 0;
  text-align: left;
  cursor: pointer;
}

.book-info {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.book-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.book-title {
  display: -webkit-box;
  overflow: hidden;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.book-author,
.book-description {
  overflow: hidden;
  color: var(--color-text-sec);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-description {
  margin: 0;
}

.progress-line {
  height: 3px;
  overflow: hidden;
  background: var(--color-border);
  border-radius: 999px;
}

.progress-line span {
  display: block;
  height: 100%;
  background: var(--color-primary);
}

.book-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: auto;
}

.book-primary-action {
  --n-color: var(--color-primary) !important;
  --n-color-hover: var(--color-primary-hover) !important;
  --n-color-pressed: var(--color-primary) !important;
  --n-color-focus: var(--color-primary-hover) !important;
  --n-border: 1px solid var(--color-primary) !important;
  --n-border-hover: 1px solid var(--color-primary-hover) !important;
  --n-border-pressed: 1px solid var(--color-primary) !important;
  --n-border-focus: 1px solid var(--color-primary-hover) !important;
  --n-text-color: #ffffff !important;
  --n-text-color-hover: #ffffff !important;
  --n-text-color-pressed: #ffffff !important;
  --n-text-color-focus: #ffffff !important;
  font-weight: 700;
}

.book-primary-action:not(:disabled) {
  box-shadow: 0 1px 2px rgba(24, 160, 88, 0.18);
}

.book-read-action:disabled {
  --n-color-disabled: var(--color-primary) !important;
  --n-border-disabled: 1px solid var(--color-primary) !important;
  --n-text-color-disabled: #ffffff !important;
  --n-opacity-disabled: 1 !important;
  cursor: not-allowed;
}
</style>
