<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { Search, Upload } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import BookCard from '@/components/book/BookCard.vue'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { useBookshelfStore } from '@/stores/bookshelf'
import type { BookFormat, BookshelfSourceType } from '@/api/types'

const store = useBookshelfStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()
const { categoryOptions, tagOptions } = useTaxonomyOptions()
const filters = reactive({
  keyword: '',
  format: null as BookFormat | null,
  source_type: null as BookshelfSourceType | null,
  category_id: null as number | null,
  tag_id: null as number | null,
  favorite: null as boolean | null
})

const formatOptions = [
  { label: '全部格式', value: null },
  { label: 'EPUB', value: 'epub' },
  { label: 'PDF', value: 'pdf' },
  { label: 'TXT', value: 'txt' }
]

const sourceOptions = [
  { label: '全部来源', value: null },
  { label: '私有上传', value: 'uploaded' },
  { label: '引自公共馆', value: 'library' }
]

const favoriteOptions = [
  { label: '全部收藏', value: null },
  { label: '仅收藏', value: true },
  { label: '仅未收藏', value: false }
]

async function fetchPage(page = 1) {
  await store.fetchBookshelf({
    page,
    keyword: filters.keyword || undefined,
    format: filters.format || undefined,
    source_type: filters.source_type || undefined,
    category_id: filters.category_id || undefined,
    tag_id: filters.tag_id || undefined,
    favorite: filters.favorite === null ? undefined : filters.favorite
  })
}

function confirmRemove(id: number) {
  dialog.warning({
    title: '移除图书',
    content: '私有上传图书会同步删除文件；公共图书只移除你的书架引用。',
    positiveText: '移除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await store.removeBookshelfItem(id)
      message.success('已移除')
    }
  })
}

onMounted(() => {
  if (route.query.forbidden) message.warning('当前账号没有管理权限')
  fetchPage()
})
</script>

<template>
  <PageShell title="我的书架" subtitle="管理个人上传和从公共图书馆引入的图书">
    <template #actions>
      <n-button type="primary" @click="router.push('/upload')">
        <template #icon><Upload :size="16" /></template>
        上传图书
      </n-button>
    </template>
    <div class="surface filter-bar">
      <n-input v-model:value="filters.keyword" clearable placeholder="搜索书名、作者" @keyup.enter="fetchPage(1)">
        <template #prefix><Search :size="16" /></template>
      </n-input>
      <n-select v-model:value="filters.format" :options="formatOptions" />
      <n-select v-model:value="filters.source_type" :options="sourceOptions" />
      <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
      <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
      <n-select v-model:value="filters.favorite" :options="favoriteOptions" placeholder="收藏" />
      <n-button type="primary" secondary @click="fetchPage(1)">筛选</n-button>
    </div>
    <n-spin :show="store.loading">
      <EmptyState v-if="!store.items.length && !store.loading" title="书架还是空的" description="上传一本私有图书，或者从公共图书馆加入一本。">
        <n-button type="primary" @click="router.push('/upload')">去上传</n-button>
      </EmptyState>
      <div v-else class="grid-books">
        <BookCard
          v-for="book in store.items"
          :key="book.id"
          :book="book"
          mode="bookshelf"
          @open="router.push(`/bookshelf/${book.id}`)"
          @read="router.push(`/reader/${book.book_id}`)"
          @toggle-favorite="store.updateBookshelfItem(book.id, { favorite: !book.favorite })"
          @toggle-pinned="store.updateBookshelfItem(book.id, { pinned: !book.pinned })"
          @remove="confirmRemove(book.id)"
        />
      </div>
    </n-spin>
    <PaginationBar :pagination="store.pagination" @change="fetchPage" />
  </PageShell>
</template>

<style scoped>
.filter-bar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 120px 150px 140px 140px auto auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
}

@media (max-width: 820px) {
  .filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>
