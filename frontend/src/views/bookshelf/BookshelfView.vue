<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { Plus, Search } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import BookCard from '@/components/book/BookCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { useBookshelfStore } from '@/stores/bookshelf'
import { bookshelfApi } from '@/api/bookshelf'
import { createBookDownloadName, downloadBlob } from '@/utils/download'
import type { BookFormat, BookshelfItem, BookshelfSourceType } from '@/api/types'

const store = useBookshelfStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()
const { categoryOptions, tagOptions } = useTaxonomyOptions()
const filterDrawer = ref(false)
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
  { label: '私人图书', value: 'uploaded' },
  { label: '公共图书', value: 'library' }
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
  filterDrawer.value = false
}

function confirmRemove(id: number) {
  dialog.warning({
    title: '移除图书',
    content: '私人图书会同时删除图书文件和相关数据；公共图书只移除你的书架引用。',
    positiveText: '移除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await store.removeBookshelfItem(id)
        message.success('已移除')
      } catch (error) {
        message.error(error instanceof Error ? error.message : '移除失败')
      }
    }
  })
}

async function downloadBook(book: BookshelfItem) {
  if (!book.readable) {
    message.warning('当前图书暂时不可下载')
    return
  }
  try {
    const response = await bookshelfApi.download(book.id)
    downloadBlob(response, createBookDownloadName(book.title, book.format))
    message.success('已开始下载')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '下载失败')
  }
}

onMounted(() => {
  if (route.query.forbidden) message.warning('当前账号没有管理权限')
  fetchPage()
})
</script>

<template>
  <PageShell title="我的书架">
    <template #actions>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle @click="filterDrawer = true">
            <template #icon>
              <Search :size="18" />
            </template>
          </n-button>
        </template>
        搜索和筛选
      </n-tooltip>
      <n-button type="primary" @click="router.push('/upload')">
        <template #icon>
          <Plus :size="18" />
        </template>
        上传图书
      </n-button>
    </template>

    <n-drawer v-model:show="filterDrawer" placement="right" :width="320">
      <n-drawer-content title="搜索和筛选">
        <div class="filter-drawer-body">
          <n-input
            v-model:value="filters.keyword"
            size="large"
            clearable
            placeholder="搜索书名、作者"
            @keyup.enter="fetchPage(1)"
          >
            <template #prefix><Search :size="16" /></template>
          </n-input>
          <n-select v-model:value="filters.format" :options="formatOptions" />
          <n-select v-model:value="filters.source_type" :options="sourceOptions" />
          <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
          <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
          <n-select v-model:value="filters.favorite" :options="favoriteOptions" placeholder="收藏" />
        </div>
        <template #footer>
          <n-button block type="primary" @click="fetchPage(1)">应用筛选</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>

    <n-spin :show="store.loading">
      <EmptyState
        v-if="!store.items.length && !store.loading"
        title="书架空空如也"
        description="上传一本图书，或从图书馆加入公共图书到书架。"
      >
        <n-button type="primary" @click="router.push('/upload')">
          <template #icon>
            <Plus :size="18" />
          </template>
          上传图书
        </n-button>
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
          @download="downloadBook(book)"
          @remove="confirmRemove(book.id)"
        />
      </div>
    </n-spin>

    <PaginationBar :pagination="store.pagination" @change="fetchPage" />
  </PageShell>
</template>

<style scoped>
.filter-drawer-body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.filter-drawer-body :deep(.n-input) {
  border-radius: var(--radius-large);
}

.filter-drawer-body :deep(.n-input__input-el) {
  font-size: var(--font-size-lg);
}

.filter-drawer-body :deep(.n-input:focus-within) {
  box-shadow: 0 0 0 4px var(--color-primary-light);
}

.filter-drawer-body :deep(.n-select) {
  border-radius: var(--radius-large);
}
</style>
