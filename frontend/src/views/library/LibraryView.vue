<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import { Library, Plus, Search } from 'lucide-vue-next'
import { useWindowSize } from '@vueuse/core'
import PageShell from '@/components/common/PageShell.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import BookCard from '@/components/book/BookCard.vue'
import BookListItem from '@/components/book/BookListItem.vue'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { useLibraryStore } from '@/stores/library'
import { libraryApi } from '@/api/library'
import { createBookDownloadName, downloadBlob } from '@/utils/download'
import type { BookFormat, LibraryBook } from '@/api/types'

const store = useLibraryStore()
const router = useRouter()
const message = useMessage()
const { categoryOptions, tagOptions } = useTaxonomyOptions()
const filterDrawer = ref(false)
const filters = reactive({ keyword: '', format: null as BookFormat | null, category_id: null as number | null, tag_id: null as number | null })

const { width } = useWindowSize()
const isMobile = computed(() => width.value < 768)

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
  filterDrawer.value = false
}

async function join(bookId: number) {
  try {
    await store.addToBookshelf(bookId)
    message.success('已加入书架')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加入失败')
  }
}

async function downloadBook(book: LibraryBook) {
  if (book.library_status !== 'approved') {
    message.warning('已下架图书不能下载')
    return
  }
  try {
    const response = await libraryApi.download(book.id)
    downloadBlob(response, createBookDownloadName(book.title, book.format))
    message.success('已开始下载')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '下载失败')
  }
}

onMounted(() => fetchPage())
</script>

<template>
  <PageShell title="图书馆">
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
      <n-button type="primary" @click="router.push('/upload?target=public')">
        <template #icon>
          <Plus :size="18" />
        </template>
        上传公共图书
      </n-button>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle @click="router.push('/library/mine')">
            <template #icon>
              <Library :size="18" />
            </template>
          </n-button>
        </template>
        我的图书馆
      </n-tooltip>
    </template>

    <n-drawer v-model:show="filterDrawer" placement="right" :width="320">
      <n-drawer-content title="搜索和筛选">
        <div class="filter-drawer-body">
          <n-input
            v-model:value="filters.keyword"
            size="large"
            clearable
            placeholder="搜索图书馆中的书籍、作者、分类..."
            @keyup.enter="fetchPage(1)"
          >
            <template #prefix><Search :size="16" /></template>
          </n-input>
          <n-select v-model:value="filters.format" :options="formatOptions" />
          <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
          <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
        </div>
        <template #footer>
          <n-button block type="primary" @click="fetchPage(1)">应用筛选</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>

    <n-spin :show="store.loading">
      <EmptyState
        v-if="!store.books.length && !store.loading"
        title="还没有公共图书"
        description="上传一本公共图书，大家就能引用阅读。"
      />

      <!-- Desktop: List Layout -->
      <div v-else-if="!isMobile" class="books-list">
        <BookListItem
          v-for="book in store.books"
          :key="book.id"
          :book="book"
          @open="router.push(`/library/${book.id}`)"
          @join="join(book.id)"
          @download="downloadBook(book)"
        />
      </div>

      <!-- Mobile: Grid Layout -->
      <div v-else class="grid-books">
        <BookCard
          v-for="book in store.books"
          :key="book.id"
          :book="book"
          mode="library"
          @open="router.push(`/library/${book.id}`)"
          @join="join(book.id)"
          @download="downloadBook(book)"
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

.books-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}
</style>
