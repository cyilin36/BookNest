<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { ChevronLeft, ChevronRight, List, Settings2 } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useReaderStore } from '@/stores/reader'
import { useSettingsStore } from '@/stores/settings'
import { useReaderProgress } from '@/composables/useReaderProgress'
import type { ReaderChapter, ReaderChapterContent, ReaderLineHeight, ThemeName } from '@/api/types'

const route = useRoute()
const message = useMessage()
const reader = useReaderStore()
const settings = useSettingsStore()
const chapterDrawer = ref(false)
const settingsDrawer = ref(false)
const readerMenuOpen = ref(false)
const bookId = Number(route.params.bookId)
const { save, saveNow } = useReaderProgress(bookId)

const activeChapter = computed(() => reader.chapters.find((chapter) => chapter.id === reader.activeChapterId) || null)
const readableChapters = computed(() => reader.chapters.filter((chapter) => !chapter.is_volume))
const bookFormatLabel = computed(() => reader.bookMeta?.format.toUpperCase() || '')
const activeContent = ref<string>('')
const activeContentType = ref<ReaderChapterContent['content_type']>('html')
const readerTopRef = ref<HTMLElement | null>(null)
const isRestoringPosition = ref(false)
const ignoreNextScrollMenuClose = ref(false)
const chapterItemRefs = new Map<number, HTMLElement>()

interface SavedReaderPosition {
  chapterId: number
  scrollRatio: number
}

function isReadableChapter(chapter: ReaderChapter | null | undefined) {
  return !!chapter && !chapter.is_volume
}

function chapterIndexById(id: number | null) {
  return id ? readableChapters.value.findIndex((chapter) => chapter.id === id) : -1
}

function fallbackReadableChapterFrom(id: number) {
  const chapterIndex = reader.chapters.findIndex((chapter) => chapter.id === id)
  if (chapterIndex < 0) return readableChapters.value[0] || null
  return reader.chapters.slice(chapterIndex).find(isReadableChapter) || [...reader.chapters].reverse().find(isReadableChapter) || null
}

function prevChapter() {
  const index = chapterIndexById(reader.activeChapterId)
  if (index > 0) {
    loadChapter(readableChapters.value[index - 1].id)
  }
}

function nextChapter() {
  const index = chapterIndexById(reader.activeChapterId)
  if (index >= 0 && index < readableChapters.value.length - 1) {
    loadChapter(readableChapters.value[index + 1].id)
  }
}

function openReaderMenu() {
  readerMenuOpen.value = true
}

function handleReaderTap(action: () => void) {
  if (readerMenuOpen.value) {
    readerMenuOpen.value = false
    return
  }
  action()
}

async function runWithMenuKept(action: () => void | Promise<void>) {
  ignoreNextScrollMenuClose.value = true
  try {
    await action()
  } finally {
    window.setTimeout(() => {
      ignoreNextScrollMenuClose.value = false
    }, 120)
  }
}

function setChapterItemRef(chapterId: number, element: Element | null) {
  if (element instanceof HTMLElement) {
    chapterItemRefs.set(chapterId, element)
  } else {
    chapterItemRefs.delete(chapterId)
  }
}

async function scrollActiveChapterIntoView() {
  await nextTick()
  if (!reader.activeChapterId) return
  chapterItemRefs.get(reader.activeChapterId)?.scrollIntoView({ block: 'center' })
}

function openChapterDrawer() {
  chapterDrawer.value = true
  scrollActiveChapterIntoView()
}

function clampPercentage(value: number) {
  return Math.max(0, Math.min(100, Math.round(value)))
}

function clampScrollRatio(value: number) {
  return Math.max(0, Math.min(1, value))
}

function progressType() {
  return reader.bookMeta?.format === 'pdf' ? 'pdf_page' : reader.bookMeta?.format === 'txt' ? 'txt_offset' : 'epub_cfi'
}

function chapterProgressPercentage(chapterId: number, scrollRatio: number) {
  const index = Math.max(0, readableChapters.value.findIndex((item) => item.id === chapterId))
  const chapterShare = (index + clampScrollRatio(scrollRatio)) / Math.max(readableChapters.value.length, 1)
  return clampPercentage(chapterShare * 100)
}

function currentScrollRatio() {
  const scrollable = document.documentElement.scrollHeight - window.innerHeight
  if (scrollable <= 0) return 0
  return clampScrollRatio(window.scrollY / scrollable)
}

function encodeProgressValue(chapterId: number, scrollRatio: number) {
  return JSON.stringify({ chapterId, scrollRatio: Number(clampScrollRatio(scrollRatio).toFixed(4)) })
}

function parseProgressValue(value: string | null | undefined): SavedReaderPosition | null {
  if (!value) return null

  try {
    const parsed = JSON.parse(value) as Partial<SavedReaderPosition> | number
    if (typeof parsed === 'number' && Number.isFinite(parsed)) {
      return { chapterId: parsed, scrollRatio: 0 }
    }
    if (typeof parsed === 'object' && Number.isFinite(parsed.chapterId)) {
      return {
        chapterId: Number(parsed.chapterId),
        scrollRatio: Number.isFinite(parsed.scrollRatio) ? clampScrollRatio(Number(parsed.scrollRatio)) : 0
      }
    }
  } catch {
    const chapterId = Number(value)
    if (Number.isFinite(chapterId)) return { chapterId, scrollRatio: 0 }
  }

  return null
}

function savedReadablePosition() {
  const saved = parseProgressValue(reader.progress?.progress_value)
  if (!saved) return null
  const savedChapter = reader.chapters.find((chapter) => chapter.id === saved.chapterId)
  if (isReadableChapter(savedChapter)) return saved
  const fallback = fallbackReadableChapterFrom(saved.chapterId)
  return fallback ? { chapterId: fallback.id, scrollRatio: 0 } : null
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

async function scrollReaderToTop() {
  await nextTick()
  readerTopRef.value?.scrollIntoView({ block: 'start' })
  window.scrollTo({ top: 0, behavior: 'auto' })
}

async function scrollReaderToRatio(scrollRatio: number) {
  await nextTick()
  await new Promise<void>((resolve) => {
    requestAnimationFrame(() => {
      const scrollable = document.documentElement.scrollHeight - window.innerHeight
      window.scrollTo({ top: Math.max(0, scrollable) * clampScrollRatio(scrollRatio), behavior: 'auto' })
      resolve()
    })
  })
}

function saveCurrentPosition(immediate = false) {
  if (!reader.activeChapterId || isRestoringPosition.value) return
  const currentChapter = reader.chapters.find((chapter) => chapter.id === reader.activeChapterId)
  if (!isReadableChapter(currentChapter) || !activeContent.value) return
  const scrollRatio = currentScrollRatio()
  const value = encodeProgressValue(reader.activeChapterId, scrollRatio)
  const percentage = chapterProgressPercentage(reader.activeChapterId, scrollRatio)
  const persist = immediate ? saveNow : save
  persist(progressType(), value, percentage)
}

function handleReaderScroll() {
  if (readerMenuOpen.value) {
    if (ignoreNextScrollMenuClose.value) {
      return
    }
    readerMenuOpen.value = false
  }
  saveCurrentPosition()
}

function handleVisibilityChange() {
  if (document.visibilityState === 'hidden') saveCurrentPosition(true)
}

async function loadChapter(chapterId: number, options: { resetScroll?: boolean; restoreScrollRatio?: number | null; saveProgress?: boolean } = {}) {
  const { resetScroll = true, restoreScrollRatio = null, saveProgress = true } = options
  const chapter = fallbackReadableChapterFrom(chapterId)
  if (!chapter) return
  reader.activeChapterId = chapter.id
  const content = await reader.loadChapterContent(bookId, chapter.id)
  activeContentType.value = content.content_type
  activeContent.value = content.content
  if (restoreScrollRatio !== null) {
    await scrollReaderToRatio(restoreScrollRatio)
  } else if (resetScroll) {
    await scrollReaderToTop()
  }
  if (saveProgress) saveCurrentPosition(true)
}

onMounted(async () => {
  try {
    reader.clearReaderState()
    await Promise.all([reader.loadMeta(bookId), reader.loadChapters(bookId), reader.loadProgress(bookId)])
    const savedPosition = savedReadablePosition()
    const initialChapterId = savedPosition?.chapterId || readableChapters.value[0]?.id
    if (initialChapterId) {
      isRestoringPosition.value = true
      await loadChapter(initialChapterId, { restoreScrollRatio: savedPosition?.scrollRatio ?? null, saveProgress: false })
      isRestoringPosition.value = false
    }
    window.addEventListener('scroll', handleReaderScroll, { passive: true })
    document.addEventListener('visibilitychange', handleVisibilityChange)
  } catch (error) {
    isRestoringPosition.value = false
    message.error(error instanceof Error ? error.message : '阅读器加载失败')
  }
})

onBeforeUnmount(() => {
  saveCurrentPosition(true)
  window.removeEventListener('scroll', handleReaderScroll)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

watch(
  () => settings.reader,
  () => {
    document.documentElement.style.setProperty('--reader-font-size', `${settings.reader.font_size}px`)
    document.documentElement.style.setProperty('--reader-line-height', String(settings.reader.line_height))
  },
  { deep: true, immediate: true }
)

watch(chapterDrawer, (visible) => {
  if (visible) scrollActiveChapterIntoView()
})

watch(
  () => reader.activeChapterId,
  () => {
    if (chapterDrawer.value) scrollActiveChapterIntoView()
  }
)
</script>

<template>
  <main class="reader-shell">
    <PageShell v-if="reader.bookMeta" :title="reader.bookMeta.title" :subtitle="reader.bookMeta.author || '在线阅读'">
      <template #actions>
        <n-button secondary @click="openChapterDrawer()">
          <template #icon><List :size="16" /></template>
          目录
        </n-button>
        <n-button secondary @click="settingsDrawer = true">
          <template #icon><Settings2 :size="16" /></template>
          设置
        </n-button>
      </template>
      <div ref="readerTopRef" class="reader-topline surface">
        <div>{{ activeChapter?.title || '未选择章节' }}</div>
        <div class="muted">{{ bookFormatLabel }}</div>
      </div>
      <section class="reader-panel surface">
        <div v-if="activeContent && activeContentType === 'html'" class="reader-content" v-html="activeContent" />
        <div v-else-if="activeContent" class="reader-content reader-content-text" v-text="activeContent" />
        <div v-if="activeContent" class="reader-tap-zones">
          <button type="button" class="reader-tap-zone" aria-label="点击左侧切换上一章" :disabled="chapterIndexById(reader.activeChapterId) <= 0 && !readerMenuOpen" @click="handleReaderTap(prevChapter)" />
          <button type="button" class="reader-tap-zone" aria-label="点击中间打开或关闭阅读菜单" @click="handleReaderTap(openReaderMenu)" />
          <button
            type="button"
            class="reader-tap-zone"
            aria-label="点击右侧切换下一章"
            :disabled="(chapterIndexById(reader.activeChapterId) < 0 || chapterIndexById(reader.activeChapterId) >= readableChapters.length - 1) && !readerMenuOpen"
            @click="handleReaderTap(nextChapter)"
          />
        </div>
        <EmptyState v-else title="暂无章节内容" description="后端未返回正文时可重试加载。">
          <n-button secondary @click="reader.loadChapters(bookId)">重试</n-button>
        </EmptyState>
      </section>
      <div class="reader-bottom toolbar">
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) <= 0" @click="prevChapter()">
          <template #icon><ChevronLeft :size="16" /></template>
          上一章
        </n-button>
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) < 0 || chapterIndexById(reader.activeChapterId) >= readableChapters.length - 1" @click="nextChapter()">
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
        <button
          v-for="chapter in reader.chapters"
          :key="chapter.id"
          :ref="(element) => setChapterItemRef(chapter.id, element as Element | null)"
          type="button"
          class="chapter-item"
          :class="{ active: chapter.id === reader.activeChapterId, volume: chapter.is_volume }"
          :disabled="chapter.is_volume"
          :aria-current="chapter.id === reader.activeChapterId ? 'true' : undefined"
          @click="chapterDrawer = false; loadChapter(chapter.id)"
        >
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

  <Transition name="reader-menu-slide">
    <section v-if="readerMenuOpen" class="reader-menu surface">
      <div class="reader-menu-title">{{ activeChapter?.title || reader.bookMeta?.title || '阅读菜单' }}</div>
      <div class="reader-menu-actions">
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) <= 0" @click="runWithMenuKept(prevChapter)">
          <template #icon><ChevronLeft :size="16" /></template>
          上一章
        </n-button>
        <n-button secondary :disabled="chapterIndexById(reader.activeChapterId) < 0 || chapterIndexById(reader.activeChapterId) >= readableChapters.length - 1" @click="runWithMenuKept(nextChapter)">
          <template #icon><ChevronRight :size="16" /></template>
          下一章
        </n-button>
        <n-button secondary @click="openChapterDrawer()">
          <template #icon><List :size="16" /></template>
          目录
        </n-button>
        <n-button secondary @click="settingsDrawer = true">
          <template #icon><Settings2 :size="16" /></template>
          设置
        </n-button>
      </div>
    </section>
  </Transition>
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
  position: relative;
  min-height: 54vh;
}

.reader-content-text {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.reader-tap-zones {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  pointer-events: none;
}

.reader-tap-zone {
  min-width: 0;
  padding: 0;
  cursor: pointer;
  background: transparent;
  border: 0;
  pointer-events: auto;
  touch-action: pan-y;
}

.reader-bottom {
  justify-content: center;
}

.reader-menu {
  position: fixed;
  right: max(16px, env(safe-area-inset-right));
  bottom: max(16px, env(safe-area-inset-bottom));
  left: max(16px, env(safe-area-inset-left));
  z-index: 20;
  width: min(560px, calc(100vw - 32px));
  margin: 0 auto;
  padding: 16px;
  border: 1px solid var(--color-border);
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.16);
}

.reader-menu-title {
  margin-bottom: 14px;
  color: var(--color-text-main);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.4;
}

.reader-menu-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.reader-menu-slide-enter-active,
.reader-menu-slide-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.reader-menu-slide-enter-from,
.reader-menu-slide-leave-to {
  opacity: 0;
  transform: translateY(18px);
}

.chapter-list {
  display: grid;
  gap: 6px;
}

.chapter-item {
  position: relative;
  display: flex;
  align-items: center;
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
  box-shadow: 0 0 0 1px rgba(24, 160, 88, 0.12);
  font-weight: 700;
}

.chapter-item.volume {
  cursor: default;
  opacity: 0.74;
}

.settings-panel {
  display: grid;
  gap: 18px;
}
</style>
