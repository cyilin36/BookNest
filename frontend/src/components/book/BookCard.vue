<script setup lang="ts">
import { computed } from 'vue'
import { BookMarked, BookOpenText, Download, Info, Library, Pin, Star, Trash2 } from 'lucide-vue-next'
import BookCover from './BookCover.vue'
import BookTitle from './BookTitle.vue'
import type { BookshelfItem, LibraryBook } from '@/api/types'

const props = defineProps<{
  book: BookshelfItem | LibraryBook
  mode: 'bookshelf' | 'library'
}>()

const emit = defineEmits<{
  open: []
  read: []
  download: []
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
const progressLabel = computed(() => `${Math.round(progress.value)}%`)
</script>

<template>
  <article class="book-card" :class="{ 'is-unreadable': mode === 'bookshelf' && 'readable' in book && !book.readable }">
    <div class="book-cover-wrap">
      <button class="cover-button" type="button" @click="mode === 'bookshelf' && 'readable' in book && book.readable ? emit('read') : emit('open')">
        <BookCover :src="book.cover_url" :title="book.title" />
        <!-- Progress bar overlay for bookshelf -->
        <div v-if="mode === 'bookshelf' && 'readable' in book" class="progress-overlay">
          <div class="progress-bar" :style="{ width: `${progress}%` }" />
        </div>
      </button>
      <n-tooltip v-if="isShelf && 'source_type' in book" trigger="hover">
        <template #trigger>
          <span class="source-chip" :class="book.source_type === 'uploaded' ? 'private' : 'library'">
            <BookMarked v-if="book.source_type === 'uploaded'" :size="13" />
            <Library v-else :size="13" />
          </span>
        </template>
        {{ book.source_type === 'uploaded' ? '私人图书' : '公共图书' }}
      </n-tooltip>
    </div>
    <div class="book-info">
      <BookTitle class="book-title" :title="book.title" @click="emit('open')" />
      <div v-if="book.author" class="book-author">{{ book.author }}</div>

      <template v-if="mode === 'bookshelf' && 'readable' in book">
        <div class="book-actions">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button class="icon-action read-action" size="small" circle secondary :disabled="!book.readable" @click="emit('read')">
                <BookOpenText :size="16" />
              </n-button>
            </template>
            阅读
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button class="icon-action" size="small" quaternary circle :disabled="!book.readable" @click="emit('download')">
                <Download :size="16" />
              </n-button>
            </template>
            下载
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button class="icon-action" size="small" quaternary circle @click="emit('toggleFavorite')">
                <Star :size="16" :fill="book.favorite ? 'currentColor' : 'none'" />
              </n-button>
            </template>
            收藏
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button class="icon-action" size="small" quaternary circle @click="emit('togglePinned')">
                <Pin :size="16" :fill="book.pinned ? 'currentColor' : 'none'" />
              </n-button>
            </template>
            置顶
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button class="icon-action" size="small" quaternary circle type="error" @click="emit('remove')">
                <Trash2 :size="16" />
              </n-button>
            </template>
            移除
          </n-tooltip>
        </div>
      </template>
    </div>
  </article>
</template>

<style scoped>
.book-card {
  min-width: 0;
  width: 100%;
  cursor: pointer;
}

.book-card:hover .cover-button {
  transform: translateY(-4px);
}

.book-card:hover .cover-button::before {
  box-shadow: var(--shadow-cover-hover);
}

.book-cover-wrap {
  position: relative;
  min-width: 0;
  width: 100%;
  margin-bottom: var(--spacing-md);
}

.cover-button {
  position: relative;
  display: block;
  width: 100%;
  padding: 0;
  background: transparent;
  border: 0;
  cursor: pointer;
  transition: transform var(--transition-slow) ease;
}

.cover-button::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: var(--radius-large);
  box-shadow: var(--shadow-cover);
  transition: box-shadow var(--transition-slow) ease;
  pointer-events: none;
}

.cover-button :deep(.book-cover) {
  border-radius: var(--radius-large);
  aspect-ratio: var(--book-cover-aspect-ratio);
}

/* Progress overlay */
.progress-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 0 0 var(--radius-large) var(--radius-large);
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: var(--color-primary);
  transition: width var(--transition-slow) ease;
}

/* Source chip */
.source-chip {
  position: absolute;
  top: var(--spacing-sm);
  right: var(--spacing-sm);
  display: inline-grid;
  width: 28px;
  height: 28px;
  place-items: center;
  background: var(--glass-bg);
  backdrop-filter: blur(var(--backdrop-blur-sm));
  -webkit-backdrop-filter: blur(var(--backdrop-blur-sm));
  border-radius: var(--radius-round);
  box-shadow: var(--shadow-light);
  z-index: 2;
}

.source-chip.private {
  color: var(--color-primary);
}

.source-chip.library {
  color: var(--color-text-sec);
}

/* Book info */
.book-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  width: 100%;
  min-width: 0;
}

.book-title {
  --book-title-height: auto;
  --book-title-font-size: var(--font-size-md);
  --book-title-font-weight: var(--font-weight-semibold);
  --book-title-line-height: var(--line-height-tight);
  --book-title-text-align: left;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: calc(var(--font-size-md) * var(--line-height-tight) * 2);
}

.book-author {
  font-size: var(--font-size-sm);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Actions */
.book-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-xs);
  width: 100%;
}

.icon-action {
  --n-width: var(--height-button-icon-sm) !important;
  --n-height: var(--height-button-icon-sm) !important;
  transition: transform var(--transition-base);
}

.icon-action:hover:not(:disabled) {
  transform: translateY(-1px);
}

.read-action:not(:disabled) {
  --n-color: var(--color-primary-suppl) !important;
  --n-color-hover: var(--color-primary-light) !important;
  --n-text-color: var(--color-primary) !important;
  --n-text-color-hover: var(--color-primary-hover) !important;
  --n-border: 1px solid transparent !important;
}

/* Unreadable state */
.is-unreadable .cover-button {
  opacity: 0.6;
}

/* Responsive */
@media (max-width: 768px) {
  .source-chip {
    width: 24px;
    height: 24px;
  }

  .source-chip :deep(svg) {
    width: 12px;
    height: 12px;
  }

  .book-title {
    --book-title-font-size: var(--font-size-base);
  }

  .book-author {
    font-size: 12px;
  }

  .icon-action {
    --n-width: 28px !important;
    --n-height: 28px !important;
  }

  .book-actions {
    gap: var(--spacing-xs);
  }
}
</style>
