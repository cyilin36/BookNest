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
const progressLabel = computed(() => `${Math.round(progress.value)}% 已读`)
</script>

<template>
  <article class="book-card" :class="{ 'is-unreadable': mode === 'bookshelf' && 'readable' in book && !book.readable }">
    <div class="book-cover-wrap">
      <button class="cover-button" type="button" @click="mode === 'bookshelf' && 'readable' in book && book.readable ? emit('read') : emit('open')">
        <BookCover :src="book.cover_url" :title="book.title" />
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
      <div class="book-author">{{ book.author || '未知作者' }}</div>
      <p v-if="'description' in book && book.description" class="book-description">{{ book.description }}</p>
      <template v-if="mode === 'bookshelf' && 'readable' in book">
        <div class="progress-line" aria-hidden="true">
          <span :style="{ width: `${progress}%` }" />
        </div>
        <div class="progress-label">{{ book.readable ? progressLabel : '暂时不可读' }}</div>
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
      <template v-else-if="mode === 'library' && 'in_bookshelf' in book">
        <div class="library-actions">
          <n-button class="book-primary-action" size="small" type="primary" :disabled="book.in_bookshelf || book.library_status !== 'approved'" @click="emit('join')">
            {{ book.in_bookshelf ? '已在书架' : book.library_status === 'hidden' ? '已下架' : '加入书架' }}
          </n-button>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button size="small" secondary circle :disabled="book.library_status !== 'approved'" @click="emit('download')">
                <Download :size="16" />
              </n-button>
            </template>
            下载
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button size="small" secondary circle @click="emit('open')">
                <Info :size="16" />
              </n-button>
            </template>
            详情
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
}

.book-cover-wrap {
  position: relative;
  min-width: 0;
  width: 100%;
}

.cover-button {
  padding: 0;
  color: inherit;
  background: transparent;
  border: 0;
  text-align: left;
  cursor: pointer;
}

.cover-button {
  display: block;
  width: 100%;
  filter: drop-shadow(0 8px 12px rgba(17, 24, 39, 0.16));
}

.cover-button :deep(.book-cover) {
  border-radius: 8px;
}

.source-chip {
  position: absolute;
  top: 8px;
  right: 8px;
  display: inline-grid;
  width: 26px;
  height: 26px;
  place-items: center;
  color: var(--color-text-sec);
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(209, 213, 219, 0.9);
  border-radius: 999px;
  box-shadow: 0 4px 10px rgba(17, 24, 39, 0.12);
}

.source-chip.private {
  color: #2563eb;
}

.source-chip.library {
  color: #4b5563;
}

.book-info {
  display: grid;
  width: 100%;
  min-width: 0;
  gap: 4px;
  padding-top: 10px;
}

.book-title {
  --book-title-height: 42px;
  --book-title-font-size: 15px;
  --book-title-font-weight: 700;
  --book-title-line-height: 21px;
  --book-title-text-align: center;
}

.book-author,
.book-description {
  overflow: hidden;
  color: var(--color-text-sec);
  font-size: 15px;
  line-height: 1.25;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-description {
  display: none;
  margin: 0;
}

.progress-line {
  height: 7px;
  margin-top: 6px;
  overflow: hidden;
  background: #d7dce5;
  border-radius: 999px;
}

.progress-line span {
  display: block;
  height: 100%;
  background: #17366f;
  border-radius: inherit;
}

.progress-label {
  color: var(--color-text-main);
  font-size: 14px;
  line-height: 1.25;
  text-align: center;
}

.book-actions,
.library-actions {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  gap: 4px;
  min-height: 32px;
  margin-top: 2px;
}

.book-actions {
  justify-content: center;
}

.library-actions {
  justify-content: flex-start;
  margin-top: 6px;
}

.icon-action {
  --n-width: 28px !important;
  --n-height: 28px !important;
}

.read-action:not(:disabled) {
  --n-color: #e8f5ee !important;
  --n-color-hover: #d5f0e2 !important;
  --n-color-pressed: #c7e8d6 !important;
  --n-text-color: var(--color-primary) !important;
  --n-text-color-hover: var(--color-primary) !important;
  --n-border: 1px solid rgba(24, 160, 88, 0.22) !important;
  --n-border-hover: 1px solid rgba(24, 160, 88, 0.34) !important;
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

.is-unreadable .cover-button {
  opacity: 0.72;
}

@media (max-width: 520px) {
  .source-chip {
    top: 6px;
    right: 6px;
    width: 22px;
    height: 22px;
  }

  .book-info {
    gap: 3px;
    padding-top: 8px;
  }

  .book-title {
    --book-title-height: 34px;
    --book-title-font-size: 14px;
    --book-title-line-height: 17px;
  }

  .book-author {
    font-size: 13px;
  }

  .progress-line {
    height: 6px;
    margin-top: 5px;
  }

  .progress-label {
    font-size: 13px;
  }

  .book-actions {
    gap: 1px;
    min-height: 26px;
  }

  .icon-action {
    --n-width: 23px !important;
    --n-height: 23px !important;
  }
}
</style>
