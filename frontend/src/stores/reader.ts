import { defineStore } from 'pinia'
import { readerApi, type ProgressPayload } from '@/api/reader'
import type { BookMeta, ReaderChapter, ReaderChapterContent, ReadingProgress } from '@/api/types'

export const useReaderStore = defineStore('reader', {
  state: () => ({
    bookMeta: null as BookMeta | null,
    chapters: [] as ReaderChapter[],
    activeChapterId: null as number | null,
    chapterContentCache: new Map<number, ReaderChapterContent>(),
    progress: null as ReadingProgress | null,
    loading: false,
    error: ''
  }),
  actions: {
    clearReaderState() {
      this.bookMeta = null
      this.chapters = []
      this.activeChapterId = null
      this.chapterContentCache = new Map()
      this.progress = null
      this.error = ''
    },
    async loadMeta(bookId: number) {
      this.bookMeta = await readerApi.meta(bookId)
      return this.bookMeta
    },
    async loadChapters(bookId: number) {
      this.chapters = await readerApi.chapters(bookId)
      if (!this.activeChapterId && this.chapters.length) this.activeChapterId = this.chapters[0].id
      return this.chapters
    },
    async loadChapterContent(bookId: number, chapterId: number) {
      const cached = this.chapterContentCache.get(chapterId)
      if (cached) return cached
      const content = await readerApi.chapterContent(bookId, chapterId)
      const next = new Map(this.chapterContentCache)
      next.set(chapterId, content)
      while (next.size > 5) {
        const firstKey = next.keys().next().value
        next.delete(firstKey)
      }
      this.chapterContentCache = next
      return content
    },
    async loadProgress(bookId: number) {
      this.progress = await readerApi.progress(bookId)
      return this.progress
    },
    async saveProgress(bookId: number, payload: ProgressPayload) {
      this.progress = await readerApi.saveProgress(bookId, payload)
      return this.progress
    }
  }
})
