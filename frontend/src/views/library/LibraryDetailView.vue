<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import PageShell from '@/components/common/PageShell.vue'
import BookCover from '@/components/book/BookCover.vue'
import FormatTag from '@/components/book/FormatTag.vue'
import { libraryApi } from '@/api/library'
import type { LibraryBook } from '@/api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const message = useMessage()
const book = ref<LibraryBook | null>(null)
const loading = ref(false)
const joining = ref(false)

async function join() {
  if (!book.value) return
  joining.value = true
  try {
    await libraryApi.addToBookshelf(book.value.id)
    book.value.in_bookshelf = true
    message.success('已加入书架')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加入失败')
  } finally {
    joining.value = false
  }
}

onMounted(async () => {
  loading.value = true
  try {
    book.value = await libraryApi.detail(Number(props.id))
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <PageShell title="公共图书详情" subtitle="公共馆图书加入书架时只建立引用关系">
    <n-spin :show="loading">
      <div v-if="book" class="detail surface">
        <BookCover :src="book.cover_url" :title="book.title" />
        <section class="detail-info">
          <div class="toolbar">
            <FormatTag :format="book.format" />
            <n-tag round :type="book.library_status === 'approved' ? 'success' : 'warning'">{{ book.library_status }}</n-tag>
          </div>
          <h1>{{ book.title }}</h1>
          <p class="muted">{{ book.author || '未知作者' }} · 上传者 {{ book.owner_username || book.owner_user_id }}</p>
          <p>{{ book.description || '暂无简介' }}</p>
          <div class="toolbar">
            <n-button type="primary" :disabled="book.in_bookshelf" :loading="joining" @click="join">
              {{ book.in_bookshelf ? '已在书架' : '加入书架' }}
            </n-button>
            <n-button v-if="book.bookshelf_id" secondary @click="router.push(`/reader/${book.id}`)">阅读</n-button>
            <n-button secondary @click="router.push('/library')">返回公共馆</n-button>
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
