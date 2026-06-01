<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import { Search, Upload } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import BookCard from '@/components/book/BookCard.vue'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { useLibraryStore } from '@/stores/library'
import type { BookFormat } from '@/api/types'

const store = useLibraryStore()
const router = useRouter()
const message = useMessage()
const { categoryOptions, tagOptions } = useTaxonomyOptions()
const filters = reactive({ keyword: '', format: null as BookFormat | null, category_id: null as number | null, tag_id: null as number | null })

const formatOptions = [
  { label: '全部格式', value: null },
  { label: 'EPUB', value: 'epub' },
  { label: 'PDF', value: 'pdf' },
  { label: 'TXT', value: 'txt' }
]

async function fetchPage(page = 1) {
  await store.fetchLibraryBooks({
    page,
    keyword: filters.keyword || undefined,
    format: filters.format || undefined,
    category_id: filters.category_id || undefined,
    tag_id: filters.tag_id || undefined
  })
}

async function join(bookId: number) {
  try {
    await store.addToBookshelf(bookId)
    message.success('已加入书架')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加入失败')
  }
}

onMounted(() => fetchPage())
</script>

<template>
  <PageShell title="公共图书馆" subtitle="所有登录用户可浏览公共图书，并以引用方式加入个人书架">
    <template #actions>
      <n-button type="primary" @click="router.push('/upload?target=public')">
        <template #icon><Upload :size="16" /></template>
        上传公共图书
      </n-button>
    </template>
    <div class="library-banner">
      加入书架时仅在后端建立引用关系，共享同一物理文件，不占用您的个人存储空间配额。
    </div>
    <div class="surface filter-bar">
      <n-input v-model:value="filters.keyword" clearable placeholder="搜索公共图书" @keyup.enter="fetchPage(1)">
        <template #prefix><Search :size="16" /></template>
      </n-input>
      <n-select v-model:value="filters.format" :options="formatOptions" />
      <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
      <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
      <n-button type="primary" secondary @click="fetchPage(1)">筛选</n-button>
    </div>
    <n-spin :show="store.loading">
      <EmptyState v-if="!store.books.length && !store.loading" title="还没有公共图书" description="上传一本公共图书，审核通过后大家都能引用阅读。" />
      <div v-else class="grid-books">
        <BookCard
          v-for="book in store.books"
          :key="book.id"
          :book="book"
          mode="library"
          @open="router.push(`/library/${book.id}`)"
          @join="join(book.id)"
        />
      </div>
    </n-spin>
    <PaginationBar :pagination="store.pagination" @change="fetchPage" />
  </PageShell>
</template>

<style scoped>
.library-banner {
  padding: 14px 16px;
  color: #11613c;
  background: linear-gradient(90deg, rgba(24, 160, 88, 0.14), rgba(24, 160, 88, 0.04));
  border: 1px solid rgba(24, 160, 88, 0.18);
  border-radius: 8px;
}

.filter-bar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 120px 140px 140px auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
}

@media (max-width: 720px) {
  .filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>
