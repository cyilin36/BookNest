<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import PageShell from '@/components/common/PageShell.vue'
import BookCover from '@/components/book/BookCover.vue'
import FormatTag from '@/components/book/FormatTag.vue'
import { libraryApi } from '@/api/library'
import { useAuthStore } from '@/stores/auth'
import type { LibraryBook } from '@/api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()
const book = ref<LibraryBook | null>(null)
const loading = ref(false)
const joining = ref(false)
const statusChanging = ref(false)

const isOwner = computed(() => Boolean(book.value && auth.user?.id === book.value.owner_user_id))
const canJoin = computed(() => Boolean(book.value && book.value.library_status === 'approved' && !book.value.in_bookshelf))
const canChangeStatus = computed(() => Boolean(book.value && isOwner.value && book.value.library_status !== 'deleted'))

async function join() {
  if (!book.value || !canJoin.value) return
  joining.value = true
  try {
    const shelfItem = await libraryApi.addToBookshelf(book.value.id)
    book.value.in_bookshelf = true
    book.value.bookshelf_id = shelfItem.id
    message.success('已加入书架')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加入失败')
  } finally {
    joining.value = false
  }
}

function confirmStatusChange() {
  if (!book.value || !canChangeStatus.value) return
  const currentBook = book.value
  const isHidden = currentBook.library_status === 'hidden'
  dialog.warning({
    title: `${isHidden ? '上架' : '下架'}「${currentBook.title}」`,
    content: isHidden
      ? '上架后该书会重新出现在图书馆，其他用户可以重新加入书架。'
      : '下架后其他用户将无法在图书馆看到该书，也不能再新加入书架；已加入书架的引用会保留，但暂时不可阅读。管理员仍可恢复或删除。',
    positiveText: isHidden ? '上架' : '下架',
    negativeText: '取消',
    onPositiveClick: async () => {
      statusChanging.value = true
      try {
        book.value = isHidden ? await libraryApi.show(currentBook.id) : await libraryApi.hide(currentBook.id)
        message.success(isHidden ? '已上架' : '已下架')
      } catch (error) {
        message.error(error instanceof Error ? error.message : isHidden ? '上架失败' : '下架失败')
      } finally {
        statusChanging.value = false
      }
    }
  })
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
  <PageShell title="图书详情">
    <n-spin :show="loading">
      <div v-if="book" class="detail surface">
        <BookCover :src="book.cover_url" :title="book.title" />
        <section class="detail-info">
          <div class="toolbar">
            <FormatTag :format="book.format" />
          </div>
          <h1>{{ book.title }}</h1>
          <p class="muted">{{ book.author || '未知作者' }} · 上传者 {{ book.owner_username || book.owner_user_id }}</p>
          <p>{{ book.description || '暂无简介' }}</p>
          <div class="toolbar">
            <n-button type="primary" :disabled="!canJoin" :loading="joining" @click="join">
              {{ book.in_bookshelf ? '已在书架' : book.library_status === 'hidden' ? '已下架' : '加入书架' }}
            </n-button>
            <n-button v-if="book.bookshelf_id" secondary @click="router.push(`/reader/${book.id}`)">阅读</n-button>
            <n-button v-if="canChangeStatus" secondary type="warning" :loading="statusChanging" @click="confirmStatusChange">
              {{ book.library_status === 'hidden' ? '上架' : '下架' }}
            </n-button>
            <n-button secondary @click="router.push('/library')">返回图书馆</n-button>
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
