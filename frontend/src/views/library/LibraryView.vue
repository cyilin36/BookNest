<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import { Library, Plus, Search } from 'lucide-vue-next'
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
const filterDrawer = ref(false)
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

onMounted(() => fetchPage())
</script>

<template>
  <PageShell class="library-page" title="图书馆">
    <template #actions>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle title="我的公共图书馆" @click="router.push('/library/mine')">
            <Library :size="18" />
          </n-button>
        </template>
        我的公共图书馆
      </n-tooltip>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle title="搜索和筛选" @click="filterDrawer = true">
            <Search :size="18" />
          </n-button>
        </template>
        搜索和筛选
      </n-tooltip>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle title="上传公共图书" @click="router.push('/upload?target=public')">
            <Plus :size="18" />
          </n-button>
        </template>
        上传公共图书
      </n-tooltip>
    </template>
    <n-drawer v-model:show="filterDrawer" placement="right" :width="320">
      <n-drawer-content title="搜索和筛选">
        <div class="filter-drawer-body">
          <n-input v-model:value="filters.keyword" clearable placeholder="搜索公共图书" @keyup.enter="fetchPage(1)">
            <template #prefix><Search :size="16" /></template>
          </n-input>
          <n-select v-model:value="filters.format" :options="formatOptions" />
          <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
          <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
        </div>
        <template #footer>
          <n-button block type="primary" @click="fetchPage(1)">筛选</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>
    <n-spin :show="store.loading">
      <EmptyState v-if="!store.books.length && !store.loading" title="还没有公共图书" description="上传一本公共图书，大家就能引用阅读。" />
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
.filter-drawer-body {
  display: grid;
  gap: 12px;
}

@media (max-width: 720px) {
  .library-page :deep(.page-header) {
    flex-direction: row;
    align-items: center;
  }

  .library-page :deep(.toolbar) {
    margin-left: auto;
  }
}
</style>
