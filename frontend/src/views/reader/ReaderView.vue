<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { ChevronLeft, ChevronRight, List, Settings2 } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useReaderStore } from '@/stores/reader'
import { useSettingsStore } from '@/stores/settings'
import { useReaderProgress } from '@/composables/useReaderProgress'
import type { ReaderLineHeight, ThemeName } from '@/api/types'

const route = useRoute()
const message = useMessage()
const reader = useReaderStore()
const settings = useSettingsStore()
const chapterDrawer = ref(false)
const settingsDrawer = ref(false)
const bookId = Number(route.params.bookId)
const { save } = useReaderProgress(bookId)

const activeChapter = computed(() => reader.chapters.find((chapter) => chapter.id === reader.activeChapterId) || null)
const activeContent = ref<string>('')

function chapterIndexById(id: number | null) {
  return id ? reader.chapters.findIndex((chapter) => chapter.id === id) : -1
}

function prevChapter() {
  const index = chapterIndexById(reader.activeChapterId)
  if (index > 0) {
    loadChapter(reader.chapters[index - 1].id)
  }
}

function nextChapter() {
  const index = chapterIndexById(reader.activeChapterId)
  if (index >= 0 && index < reader.chapters.length - 1) {
    loadChapter(reader.chapters[index + 1].id)
  }
}

function updateTheme(value: ThemeName) {
  settings.setTheme(value)
}

function updateLineHeight(value: ReaderLineHeight) {
  settings.updateReaderSettings({ line_height: value })
}

function updateFontSize(value: number) {
  settings.updateReaderSettings({ font_size: value })
}

async function loadChapter(chapterId: number) {
  reader.activeChapterId = chapterId
  const content = await reader.loadChapterContent(bookId, chapterId)
  activeContent.value = content.content
  const progressType = reader.bookMeta?.format === 'pdf' ? 'pdf_page' : reader.bookMeta?.format === 'txt' ? 'txt_offset' : 'epub_cfi'
  save(progressType, String(chapterId), Math.round(((reader.chapters.findIndex((item) => item.id === chapterId) + 1) / Math.max(reader.chapters.length, 1)) * 100))
}

onMounted(async () => {
  try {
    await Promise.all([reader.loadMeta(bookId), reader.loadChapters(bookId), reader.loadProgress(bookId)])
    if (reader.chapters[0]) await loadChapter(reader.chapters[0].id)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '阅读器加载失败')
  }
})

watch(
  () => settings.reader,
  () => {
    document.documentElement.style.setProperty('--reader-font-size', `${settings.reader.font_size}px`)
    document.documentElement.style.setProperty('--reader-line-height', String(settings.reader.line_height))
  },
  { deep: true, immediate: true }
)
</script>

<template>
  <main class="reader-shell">
    <PageShell v-if="reader.bookMeta" :title="reader.bookMeta.title" :subtitle="reader.bookMeta.author || '在线阅读'">
      <template #actions>
        <n-button secondary @click="chapterDrawer = true">
          <template #icon><List :size="16" /></template>
          目录
        </n-button>
        <n-button secondary @click="settingsDrawer = true">
          <template #icon><Settings2 :size="16" /></template>
          设置
        </n-button>
      </template>
      <div class="reader-topline surface">
        <div>{{ activeChapter?.title || '未选择章节' }}</div>
        <div class="muted">{{ reader.progress?.progress_type || reader.bookMeta.format }}</div>
      </div>
      <section class="reader-panel surface">
        <div v-if="activeContent" class="reader-content" v-html="activeContent" />
        <EmptyState v-else title="暂无章节内容" description="后端未返回正文时可重试加载。">
          <n-button secondary @click="reader.loadChapters(bookId)">重试</n-button>
        </EmptyState>
      </section>
      <div class="reader-bottom toolbar">
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) <= 0" @click="prevChapter()">
          <template #icon><ChevronLeft :size="16" /></template>
          上一章
        </n-button>
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) < 0 || chapterIndexById(reader.activeChapterId) >= reader.chapters.length - 1" @click="nextChapter()">
          <template #icon><ChevronRight :size="16" /></template>
          下一章
        </n-button>
      </div>
    </PageShell>
    <EmptyState v-else title="阅读器未加载" description="请返回书架后重新进入。">
      <n-button @click="$router.push('/bookshelf')">返回书架</n-button>
    </EmptyState>
  </main>

  <n-drawer v-model:show="chapterDrawer" placement="left" :width="320">
    <n-drawer-content title="目录">
      <div class="chapter-list">
        <button v-for="chapter in reader.chapters" :key="chapter.id" type="button" class="chapter-item" :class="{ active: chapter.id === reader.activeChapterId }" @click="chapterDrawer = false; loadChapter(chapter.id)">
          <span>{{ chapter.title }}</span>
          <small v-if="chapter.is_volume">卷</small>
        </button>
      </div>
    </n-drawer-content>
  </n-drawer>

  <n-drawer v-model:show="settingsDrawer" placement="right" :width="320">
    <n-drawer-content title="阅读设置">
      <div class="settings-panel">
        <n-radio-group v-model:value="settings.reader.theme" @update:value="updateTheme">
          <n-space vertical>
            <n-radio value="modern">Modern</n-radio>
            <n-radio value="sepia">Sepia</n-radio>
            <n-radio value="dark">Dark</n-radio>
          </n-space>
        </n-radio-group>
        <n-slider v-model:value="settings.reader.font_size" :min="14" :max="26" :step="1" @update:value="updateFontSize" />
        <n-select
          v-model:value="settings.reader.line_height"
          :options="[
            { label: '紧凑 1.5', value: 1.5 },
            { label: '适中 1.8', value: 1.8 },
            { label: '宽松 2.2', value: 2.2 }
          ]"
          @update:value="updateLineHeight"
        />
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.reader-topline,
.reader-panel,
.reader-bottom {
  padding: 14px;
}

.reader-topline {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.reader-panel {
  min-height: 54vh;
}

.reader-bottom {
  justify-content: center;
}

.chapter-list {
  display: grid;
  gap: 6px;
}

.chapter-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  color: var(--color-text-main);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  text-align: left;
}

.chapter-item.active {
  color: var(--color-primary);
  border-color: var(--color-primary);
  background: var(--color-primary-suppl);
}

.settings-panel {
  display: grid;
  gap: 18px;
}
</style>
