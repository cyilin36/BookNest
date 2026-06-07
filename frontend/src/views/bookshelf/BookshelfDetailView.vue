<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Download } from 'lucide-vue-next'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import PageShell from '@/components/common/PageShell.vue'
import BookCover from '@/components/book/BookCover.vue'
import FormatTag from '@/components/book/FormatTag.vue'
import { bookshelfApi } from '@/api/bookshelf'
import { createBookDownloadName, downloadBlob } from '@/utils/download'
import type { BookshelfItem } from '@/api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const message = useMessage()
const book = ref<BookshelfItem | null>(null)
const loading = ref(false)
const downloading = ref(false)

function formatUnreadableReason(reason: BookshelfItem['unreadable_reason']) {
  const labels: Record<NonNullable<BookshelfItem['unreadable_reason']>, string> = {
    library_hidden: '公共图书已下架',
    library_deleted: '公共图书已删除',
    file_missing: '图书文件丢失',
    permission_denied: '暂无阅读权限'
  }
  return reason ? labels[reason] : '暂时无法阅读'
}

async function downloadBook() {
  if (!book.value || !book.value.readable) return
  downloading.value = true
  try {
    const response = await bookshelfApi.download(book.value.id)
    downloadBlob(response, createBookDownloadName(book.value.title, book.value.format))
    message.success('已开始下载')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '下载失败')
  } finally {
    downloading.value = false
  }
}

onMounted(async () => {
  loading.value = true
  try {
    book.value = await bookshelfApi.detail(Number(props.id))
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <PageShell title="书架详情" subtitle="查看个人书架项状态和继续阅读">
    <n-spin :show="loading">
      <div v-if="book" class="detail surface">
        <BookCover :src="book.cover_url" :title="book.title" />
        <section class="detail-info">
          <div class="toolbar">
            <FormatTag :format="book.format" />
            <n-tag round>{{ book.source_type === 'uploaded' ? '私人图书' : '公共图书' }}</n-tag>
            <n-tag v-if="!book.readable" type="error" round>{{ formatUnreadableReason(book.unreadable_reason) }}</n-tag>
          </div>
          <h1>{{ book.title }}</h1>
          <p class="muted">{{ book.author || '未知作者' }}</p>
          <n-progress type="line" :percentage="Math.round(book.progress_percentage || 0)" :height="4" :show-indicator="false" />
          <div class="toolbar">
            <n-button type="primary" :disabled="!book.readable" @click="router.push(`/reader/${book.book_id}`)">继续阅读</n-button>
            <n-button secondary :disabled="!book.readable" :loading="downloading" @click="downloadBook">
              <template #icon><Download :size="16" /></template>
              下载
            </n-button>
            <n-button secondary @click="router.push('/bookshelf')">返回书架</n-button>
          </div>
        </section>
      </div>
    </n-spin>
  </PageShell>
</template>

<style scoped>
.detail {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 24px;
  padding: 18px;
}

.detail-info h1 {
  margin: 12px 0 6px;
  font-size: 28px;
}

@media (max-width: 640px) {
  .detail {
    grid-template-columns: 1fr;
  }
}
</style>
