<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { Plus, Search } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
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
  filterDrawer.value = false
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
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle title="搜索和筛选" @click="filterDrawer = true">
            <Search :size="18" />
          </n-button>
        </template>
        搜索和筛选
      </n-tooltip>
    </template>
    <n-drawer v-model:show="filterDrawer" placement="right" :width="320">
      <n-drawer-content title="搜索和筛选">
        <div class="filter-drawer-body">
          <n-input v-model:value="filters.keyword" clearable placeholder="搜索书名、作者" @keyup.enter="fetchPage(1)">
            <template #prefix><Search :size="16" /></template>
          </n-input>
          <n-select v-model:value="filters.format" :options="formatOptions" />
          <n-select v-model:value="filters.source_type" :options="sourceOptions" />
          <n-select v-model:value="filters.category_id" clearable :options="categoryOptions" placeholder="分类" />
          <n-select v-model:value="filters.tag_id" clearable :options="tagOptions" placeholder="标签" />
          <n-select v-model:value="filters.favorite" :options="favoriteOptions" placeholder="收藏" />
        </div>
        <template #footer>
          <n-button block type="primary" @click="fetchPage(1)">筛选</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>
    <n-spin :show="store.loading">
      <div v-if="!store.loading" class="grid-books">
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
        <button class="upload-book-card surface" type="button" aria-label="上传图书" @click="router.push('/upload')">
          <Plus :size="42" stroke-width="1.8" />
        </button>
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

.upload-book-card {
  display: grid;
  min-height: 214px;
  place-items: center;
  padding: 12px;
  color: var(--color-primary);
  background: var(--color-bg-card);
  cursor: pointer;
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease,
    transform 0.16s ease;
}

.upload-book-card:hover {
  border-color: var(--color-primary);
  box-shadow: inset 0 0 0 1px var(--color-primary);
  transform: translateY(-1px);
}
</style>
