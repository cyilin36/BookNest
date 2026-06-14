<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Download, Pencil } from 'lucide-vue-next'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import PageShell from '@/components/common/PageShell.vue'
import BookCover from '@/components/book/BookCover.vue'
import BookInfoEditModal, { type BookInfoEditPayload } from '@/components/book/BookInfoEditModal.vue'
import FormatTag from '@/components/book/FormatTag.vue'
import { bookshelfApi } from '@/api/bookshelf'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { createBookDownloadName, downloadBlob } from '@/utils/download'
import type { BookshelfItem } from '@/api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const message = useMessage()
const { categoryOptions, tagOptions, loading: taxonomyLoading } = useTaxonomyOptions()
const book = ref<BookshelfItem | null>(null)
const loading = ref(false)
const downloading = ref(false)
const editOpen = ref(false)
const savingInfo = ref(false)
const coverRevision = ref(0)

const categoryLabel = computed(() => {
  const categoryId = book.value?.personal_category_id
  return categoryId ? categoryOptions.value.find((item) => item.value === categoryId)?.label || '' : ''
})
const tagLabels = computed(() => (book.value?.tag_ids || []).map((id) => tagOptions.value.find((item) => item.value === id)?.label).filter(Boolean))

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

async function saveBookInfo(payload: BookInfoEditPayload) {
  if (!book.value) return
  savingInfo.value = true
  try {
    let updated = await bookshelfApi.update(book.value.id, {
      title: payload.title,
      author: payload.author,
      description: payload.description,
      personal_category_id: payload.category_ids?.[0] ?? null,
      tag_ids: payload.tag_ids || []
    })
    if (payload.cover) {
      updated = await bookshelfApi.updateCover(book.value.id, payload.cover)
      coverRevision.value = Date.now()
    }
    book.value = updated
    editOpen.value = false
    message.success('图书信息已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    savingInfo.value = false
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
        <BookCover :src="book.cover_url" :title="book.title" :revision="coverRevision" />
        <section class="detail-info">
          <div class="toolbar">
            <FormatTag :format="book.format" />
            <n-tag round>{{ book.source_type === 'uploaded' ? '私人图书' : '公共图书' }}</n-tag>
            <n-tag v-if="!book.readable" type="error" round>{{ formatUnreadableReason(book.unreadable_reason) }}</n-tag>
          </div>
          <h1>{{ book.title }}</h1>
          <p class="muted">{{ book.author || '未知作者' }}</p>
          <div v-if="categoryLabel || tagLabels.length" class="taxonomy-line">
            <n-tag v-if="categoryLabel" size="small" round>{{ categoryLabel }}</n-tag>
            <n-tag v-for="label in tagLabels" :key="label" size="small" round>{{ label }}</n-tag>
          </div>
          <p class="description">{{ book.description || '暂无简介' }}</p>
          <n-progress type="line" :percentage="Math.round(book.progress_percentage || 0)" :height="4" :show-indicator="false" />
          <div class="toolbar">
            <n-button type="primary" :disabled="!book.readable" @click="router.push(`/reader/${book.book_id}`)">继续阅读</n-button>
            <n-button secondary :disabled="!book.readable" :loading="downloading" @click="downloadBook">
              <template #icon><Download :size="16" /></template>
              下载
            </n-button>
            <n-button secondary @click="editOpen = true">
              <template #icon><Pencil :size="16" /></template>
              编辑信息
            </n-button>
            <n-button secondary @click="router.push('/bookshelf')">返回书架</n-button>
          </div>
        </section>
      </div>
    </n-spin>
    <BookInfoEditModal
      v-if="book"
      v-model:show="editOpen"
      title="编辑图书信息"
      :initial="{ title: book.title, author: book.author, description: book.description, category_ids: book.personal_category_id ? [book.personal_category_id] : [], tag_ids: book.tag_ids || [] }"
      :category-options="categoryOptions"
      :tag-options="tagOptions"
      :taxonomy-loading="taxonomyLoading"
      single-category
      :saving="savingInfo"
      allow-cover
      :current-cover-url="book.cover_url"
      @save="saveBookInfo"
    />
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

.description {
  max-width: 720px;
  margin: 12px 0;
  color: var(--color-text-main);
  line-height: 1.7;
  white-space: pre-wrap;
}

.taxonomy-line {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

@media (max-width: 640px) {
  .detail {
    grid-template-columns: 1fr;
  }
}
</style>
