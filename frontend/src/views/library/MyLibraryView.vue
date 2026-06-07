<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { ArrowLeft, Download, Eye, EyeOff, Pencil, Plus } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import BookCover from '@/components/book/BookCover.vue'
import BookInfoEditModal, { type BookInfoEditPayload } from '@/components/book/BookInfoEditModal.vue'
import BookTitle from '@/components/book/BookTitle.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PageShell from '@/components/common/PageShell.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { libraryApi } from '@/api/library'
import { createBookDownloadName, downloadBlob } from '@/utils/download'
import type { LibraryBook, Pagination } from '@/api/types'

type OwnLibraryStatusFilter = 'all' | 'approved' | 'hidden'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const books = ref<LibraryBook[]>([])
const loading = ref(false)
const statusChangingId = ref<number | null>(null)
const downloadingId = ref<number | null>(null)
const editingBook = ref<LibraryBook | null>(null)
const savingInfo = ref(false)
const pagination = ref<Pagination>({ page: 1, page_size: 12, total: 0 })
const filters = reactive({
  status: 'all' as OwnLibraryStatusFilter
})

const statusOptions = [
  { label: '全部', value: 'all' },
  { label: '已上架', value: 'approved' },
  { label: '已下架', value: 'hidden' }
]

async function fetchPage(page = 1) {
  loading.value = true
  try {
    const result = await libraryApi.list({
      page,
      page_size: pagination.value.page_size,
      mine: true,
      status: filters.status === 'all' ? undefined : filters.status,
      sort: 'created_at',
      order: 'desc'
    })
    books.value = result.data
    pagination.value = result.pagination
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function replaceBook(nextBook: LibraryBook) {
  const index = books.value.findIndex((book) => book.id === nextBook.id)
  if (index >= 0) {
    if (filters.status !== 'all' && nextBook.library_status !== filters.status) {
      books.value.splice(index, 1)
      pagination.value.total = Math.max(0, pagination.value.total - 1)
    } else {
      books.value[index] = nextBook
    }
  }
}

function confirmStatusChange(book: LibraryBook) {
  const isHidden = book.library_status === 'hidden'
  dialog.warning({
    title: `${isHidden ? '上架' : '下架'}「${book.title}」`,
    content: isHidden
      ? '上架后该书会重新出现在图书馆，其他用户可以重新加入书架。'
      : '下架后该书会从图书馆普通列表中隐藏，不能再被新加入书架；已加入书架的引用会保留但暂时不可阅读。',
    positiveText: isHidden ? '上架' : '下架',
    negativeText: '取消',
    onPositiveClick: async () => {
      statusChangingId.value = book.id
      try {
        replaceBook(isHidden ? await libraryApi.show(book.id) : await libraryApi.hide(book.id))
        message.success(isHidden ? '已上架' : '已下架')
      } catch (error) {
        message.error(error instanceof Error ? error.message : isHidden ? '上架失败' : '下架失败')
      } finally {
        statusChangingId.value = null
      }
    }
  })
}

async function downloadBook(book: LibraryBook) {
  if (book.library_status !== 'approved') {
    message.warning('已下架图书不能下载')
    return
  }
  downloadingId.value = book.id
  try {
    const response = await libraryApi.download(book.id)
    downloadBlob(response, createBookDownloadName(book.title, book.format))
    message.success('已开始下载')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '下载失败')
  } finally {
    downloadingId.value = null
  }
}

async function saveBookInfo(payload: BookInfoEditPayload) {
  if (!editingBook.value) return
  savingInfo.value = true
  try {
    const updated = await libraryApi.update(editingBook.value.id, payload)
    replaceBook(updated)
    editingBook.value = null
    message.success('图书信息已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    savingInfo.value = false
  }
}

onMounted(() => fetchPage())
</script>

<template>
  <PageShell class="my-library-page" title="我的公共图书馆">
    <template #actions>
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button secondary circle title="返回图书馆" @click="router.push('/library')">
            <ArrowLeft :size="18" />
          </n-button>
        </template>
        返回图书馆
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

    <n-radio-group v-model:value="filters.status" size="small" @update:value="fetchPage(1)">
      <n-radio-button v-for="option in statusOptions" :key="option.value" :value="option.value">
        {{ option.label }}
      </n-radio-button>
    </n-radio-group>

    <n-spin :show="loading">
      <EmptyState v-if="!books.length && !loading" title="还没有公共图书" description="上传公共图书后，可以在这里管理上架和下架。" />
      <div v-else class="grid-books my-library-grid">
        <article v-for="book in books" :key="book.id" class="own-book-card surface">
          <button class="cover-button" type="button" @click="router.push(`/library/${book.id}`)">
            <BookCover :src="book.cover_url" :title="book.title" />
          </button>
          <div class="own-book-info">
            <BookTitle class="book-title" :title="book.title" @click="router.push(`/library/${book.id}`)" />
            <div class="book-author">{{ book.author || '未知作者' }}</div>
            <div class="own-book-actions">
              <n-tag size="small" :type="book.library_status === 'approved' ? 'success' : 'warning'" round>
                {{ book.library_status === 'approved' ? '已上架' : '已下架' }}
              </n-tag>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button
                    size="small"
                    secondary
                    circle
                    :disabled="book.library_status !== 'approved'"
                    :loading="downloadingId === book.id"
                    @click="downloadBook(book)"
                  >
                    <Download :size="15" />
                  </n-button>
                </template>
                下载
              </n-tooltip>
              <n-button
                class="status-button"
                size="small"
                secondary
                :type="book.library_status === 'approved' ? 'warning' : 'primary'"
                :loading="statusChangingId === book.id"
                @click="confirmStatusChange(book)"
              >
                <template #icon>
                  <EyeOff v-if="book.library_status === 'approved'" :size="15" />
                  <Eye v-else :size="15" />
                </template>
                {{ book.library_status === 'approved' ? '下架' : '上架' }}
              </n-button>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button size="small" secondary circle @click="editingBook = book">
                    <Pencil :size="15" />
                  </n-button>
                </template>
                编辑信息
              </n-tooltip>
              <n-button size="small" secondary @click="router.push(`/library/${book.id}`)">详情</n-button>
            </div>
          </div>
        </article>
      </div>
    </n-spin>
    <PaginationBar :pagination="pagination" @change="fetchPage" />
    <BookInfoEditModal
      :show="Boolean(editingBook)"
      title="编辑公共图书信息"
      :initial="editingBook ? { title: editingBook.title, author: editingBook.author, description: editingBook.description } : null"
      :saving="savingInfo"
      @update:show="($event) => { if (!$event) editingBook = null }"
      @save="saveBookInfo"
    />
  </PageShell>
</template>

<style scoped>
.my-library-page :deep(.page-header) {
  align-items: center;
}

.my-library-grid {
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
}

.own-book-card {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
  min-height: 142px;
  padding: 12px;
}

.cover-button {
  padding: 0;
  color: inherit;
  background: transparent;
  border: 0;
  text-align: left;
  cursor: pointer;
}

.own-book-info {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.book-title {
  --book-title-height: 42px;
  --book-title-font-size: 15px;
  --book-title-font-weight: 700;
  --book-title-line-height: 21px;
  --book-title-text-align: center;
}

.book-author {
  overflow: hidden;
  color: var(--color-text-sec);
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.own-book-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: auto;
}

.status-button {
  font-weight: 700;
}

@media (max-width: 720px) {
  .my-library-page :deep(.page-header) {
    flex-direction: row;
  }

  .my-library-page :deep(.toolbar) {
    margin-left: auto;
  }
}
</style>
