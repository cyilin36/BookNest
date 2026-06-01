import { apiClient, unwrap } from './client'
import type { BookMeta, ReaderChapter, ReaderChapterContent, ReadingProgress } from './types'

export interface ProgressPayload {
  progress_type: 'epub_cfi' | 'pdf_page' | 'txt_offset'
  progress_value: string
  percentage?: number | null
}

export const readerApi = {
  meta(bookId: number) {
    return unwrap<BookMeta>(apiClient.get(`/reader/books/${bookId}/meta`))
  },
  chapters(bookId: number) {
    return unwrap<ReaderChapter[]>(apiClient.get(`/reader/books/${bookId}/chapters`))
  },
  chapterContent(bookId: number, chapterId: number) {
    return unwrap<ReaderChapterContent>(apiClient.get(`/reader/books/${bookId}/chapters/${chapterId}/content`))
  },
  progress(bookId: number) {
    return unwrap<ReadingProgress | null>(apiClient.get(`/reader/books/${bookId}/progress`))
  },
  saveProgress(bookId: number, payload: ProgressPayload) {
    return unwrap<ReadingProgress>(apiClient.put(`/reader/books/${bookId}/progress`, payload))
  },
  fileUrl(bookId: number) {
    return `${import.meta.env.VITE_API_BASE || '/api/v1'}/reader/books/${bookId}/file`
  },
  coverUrl(bookId: number) {
    return `${import.meta.env.VITE_API_BASE || '/api/v1'}/reader/books/${bookId}/cover`
  }
}
