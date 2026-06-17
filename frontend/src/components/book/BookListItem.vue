<script setup lang="ts">
import { Download, Info, Plus } from 'lucide-vue-next'
import BookCover from './BookCover.vue'
import type { LibraryBook } from '@/api/types'

const props = defineProps<{
  book: LibraryBook
}>()

const emit = defineEmits<{
  open: []
  download: []
  join: []
}>()
</script>

<template>
  <article class="book-list-item">
    <button class="item-cover-button" type="button" @click="emit('open')">
      <BookCover :src="book.cover_url" :title="book.title" />
    </button>

    <div class="item-content">
      <div class="item-header">
        <h3 class="item-title" @click="emit('open')">{{ book.title }}</h3>
        <div class="item-meta-badges">
          <span v-if="book.format" class="meta-badge format-badge">{{ book.format.toUpperCase() }}</span>
          <span v-if="book.library_status === 'hidden'" class="meta-badge status-badge">已下架</span>
        </div>
      </div>

      <p v-if="book.author" class="item-author">{{ book.author }}</p>

      <p v-if="book.description" class="item-description">{{ book.description }}</p>

      <div class="item-footer">
        <div class="item-stats">
          <span v-if="book.category_ids && book.category_ids.length > 0" class="stat-item">
            <span class="stat-icon">📚</span>
            分类
          </span>
        </div>

        <div class="item-actions">
          <n-button
            size="small"
            type="primary"
            :disabled="book.in_bookshelf || book.library_status !== 'approved'"
            @click="emit('join')"
          >
            <template #icon>
              <Plus :size="16" />
            </template>
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
      </div>
    </div>
  </article>
</template>

<style scoped>
.book-list-item {
  display: flex;
  gap: var(--spacing-xl);
  padding: var(--spacing-xl);
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-large);
  box-shadow: var(--shadow-light);
  transition: all var(--transition-slow) ease;
}

.book-list-item:hover {
  box-shadow: var(--shadow-medium);
  transform: translateY(-2px);
}

.item-cover-button {
  flex-shrink: 0;
  width: 80px;
  padding: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: transform var(--transition-base) ease, filter var(--transition-base) ease;
  filter: drop-shadow(var(--shadow-medium));
}

.item-cover-button:hover {
  transform: translateY(-2px);
  filter: drop-shadow(var(--shadow-heavy));
}

.item-cover-button :deep(.book-cover) {
  width: 80px;
  height: 120px;
  border-radius: var(--radius-small);
}

.item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.item-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-lg);
}

.item-title {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-tight);
  color: var(--color-text-main);
  cursor: pointer;
  transition: color var(--transition-base);
}

.item-title:hover {
  color: var(--color-primary);
}

.item-meta-badges {
  display: flex;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.meta-badge {
  padding: 3px var(--spacing-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: var(--letter-spacing-wider);
  border-radius: var(--radius-xsmall);
  white-space: nowrap;
}

.format-badge {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.status-badge {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.item-author {
  margin: 0;
  font-size: var(--font-size-md);
  color: var(--color-text-sec);
  line-height: var(--line-height-tight);
}

.item-description {
  margin: 0;
  font-size: var(--font-size-base);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.item-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-lg);
  margin-top: auto;
  padding-top: var(--spacing-xs);
}

.item-stats {
  display: flex;
  gap: var(--spacing-xl);
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-sec);
}

.stat-icon {
  font-size: var(--font-size-base);
}

.item-actions {
  display: flex;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.item-actions :deep(.n-button) {
  transition: transform var(--transition-base) ease;
}

.item-actions :deep(.n-button:hover:not(:disabled)) {
  transform: translateY(-1px);
}

.item-actions :deep(.n-button--primary-type) {
  box-shadow: var(--shadow-button-primary);
}

.item-actions :deep(.n-button--primary-type:hover:not(:disabled)) {
  box-shadow: var(--shadow-button-primary-hover);
}

@media (max-width: 768px) {
  .book-list-item {
    flex-direction: column;
    gap: var(--spacing-lg);
    padding: var(--spacing-lg);
  }

  .item-cover-button {
    width: 100%;
  }

  .item-cover-button :deep(.book-cover) {
    width: 100%;
    height: auto;
    aspect-ratio: 2 / 3;
  }

  .item-header {
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .item-meta-badges {
    align-self: flex-start;
  }

  .item-footer {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .item-actions {
    width: 100%;
  }

  .item-actions :deep(.n-button:first-child) {
    flex: 1;
  }
}
</style>
