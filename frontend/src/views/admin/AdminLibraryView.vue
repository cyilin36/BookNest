<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { Search } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { adminApi, type AdminLibraryEditableStatus } from '@/api/admin'
import type { BookFormat, LibraryBook, LibraryStatus, Pagination } from '@/api/types'
import { formatBytes, formatDate } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()
const books = ref<LibraryBook[]>([])
const loading = ref(false)
const filterDrawer = ref(false)
const pagination = ref<Pagination>({ page: 1, page_size: 20, total: 0 })
const filters = reactive({
  keyword: '',
  status: null as LibraryStatus | null,
  format: null as BookFormat | null
})

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '已通过', value: 'approved' },
  { label: '已隐藏', value: 'hidden' },
  { label: '已删除', value: 'deleted' }
]

const formatOptions = [
  { label: '全部格式', value: null },
  { label: 'EPUB', value: 'epub' },
  { label: 'PDF', value: 'pdf' },
  { label: 'TXT', value: 'txt' }
]

const transitionOptions: Record<LibraryStatus, { label: string; status: AdminLibraryEditableStatus }[]> = {
  approved: [{ label: '隐藏', status: 'hidden' }],
  hidden: [{ label: '恢复', status: 'approved' }],
  deleted: []
}

async function fetchBooks(page = 1) {
  loading.value = true
  try {
    const result = await adminApi.library({
      page,
      page_size: pagination.value.page_size,
      keyword: filters.keyword || undefined,
      status: filters.status || undefined,
      format: filters.format || undefined
    })
    books.value = result.data
    pagination.value = result.pagination
  } finally {
    loading.value = false
    filterDrawer.value = false
  }
}

async function changeStatus(book: LibraryBook, status: AdminLibraryEditableStatus) {
  try {
    Object.assign(book, await adminApi.updateLibraryStatus(book.id, { status }))
    message.success('状态已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新失败')
  }
}

function confirmDelete(book: LibraryBook) {
  dialog.warning({
    title: `删除「${book.title}」`,
    content: '删除后会移除公共图书、物理文件、书架引用、阅读进度和书签，且不可恢复。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminApi.deleteLibraryBook(book.id)
        books.value = books.value.filter((row) => row.id !== book.id)
        pagination.value.total = Math.max(0, pagination.value.total - 1)
        message.success('已删除')
      } catch (error) {
        message.error(error instanceof Error ? error.message : '删除失败')
      }
    }
  })
}

onMounted(() => fetchBooks())
</script>

<template>
  <PageShell title="公共图书管理" subtitle="管理公共图书的状态和删除">
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
          <n-input v-model:value="filters.keyword" clearable placeholder="搜索书名或作者" @keyup.enter="fetchBooks(1)">
            <template #prefix><Search :size="16" /></template>
          </n-input>
          <n-select v-model:value="filters.status" :options="statusOptions" />
          <n-select v-model:value="filters.format" :options="formatOptions" />
        </div>
        <template #footer>
          <n-button block type="primary" @click="fetchBooks(1)">筛选</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>
    <n-spin :show="loading">
      <div class="surface table">
        <div class="row header">
          <strong>图书</strong>
          <strong>状态</strong>
          <strong>上传者</strong>
          <strong>大小</strong>
          <strong>上传时间</strong>
          <strong>操作</strong>
        </div>
        <div v-for="book in books" :key="book.id" class="row">
          <div>
            <strong>{{ book.title }}</strong>
            <span>{{ book.author || '未知作者' }} · {{ book.format.toUpperCase() }}</span>
          </div>
          <n-tag size="small" :type="book.library_status === 'approved' ? 'success' : book.library_status === 'hidden' ? 'warning' : 'default'">
            {{ book.library_status }}
          </n-tag>
          <span>{{ book.owner_username || book.owner_user_id }}</span>
          <span>{{ formatBytes(book.file_size) }}</span>
          <span>{{ formatDate(book.created_at) }}</span>
          <div class="toolbar action-buttons">
            <n-button
              v-for="option in transitionOptions[book.library_status]"
              :key="option.status"
              size="small"
              secondary
              @click="changeStatus(book, option.status)"
            >
              {{ option.label }}
            </n-button>
            <n-button size="small" quaternary type="error" :disabled="book.library_status === 'deleted'" @click="confirmDelete(book)">删除</n-button>
          </div>
        </div>
      </div>
    </n-spin>
    <PaginationBar :pagination="pagination" @change="fetchBooks" />
  </PageShell>
</template>

<style scoped>
.filter-drawer-body {
  display: grid;
  gap: 10px;
}

.table {
  display: grid;
  padding: 12px;
  overflow-x: auto;
}

.row {
  display: grid;
  grid-template-columns: minmax(220px, 2fr) 90px 110px 100px 160px 150px;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--color-border);
}

.row.header {
  color: var(--color-text-sec);
  font-size: 13px;
}

.row.header strong:last-child {
  text-align: center;
}

.row div {
  min-width: 0;
}

.row div span {
  display: block;
  overflow: hidden;
  color: var(--color-text-sec);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-buttons {
  justify-content: center;
  flex-wrap: nowrap;
}

</style>
