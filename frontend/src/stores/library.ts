import { defineStore } from 'pinia'
import { libraryApi, type LibraryQuery } from '@/api/library'
import type { LibraryBook, Pagination } from '@/api/types'

export const useLibraryStore = defineStore('library', {
  state: () => ({
    books: [] as LibraryBook[],
    filters: { page: 1, page_size: 12, sort: 'created_at', order: 'desc' as const } as LibraryQuery,
    pagination: { page: 1, page_size: 12, total: 0 } as Pagination,
    joiningBookId: null as number | null,
    loading: false,
    error: ''
  }),
  actions: {
    async fetchLibraryBooks(extra: LibraryQuery = {}) {
      this.loading = true
      this.error = ''
      try {
        this.filters = { ...this.filters, ...extra }
        const response = await libraryApi.list(this.filters)
        this.books = response.data
        this.pagination = response.pagination
      } catch (error) {
        this.error = error instanceof Error ? error.message : '图书馆加载失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    async addToBookshelf(bookId: number) {
      this.joiningBookId = bookId
      try {
        const shelfItem = await libraryApi.addToBookshelf(bookId)
        const book = this.books.find((item) => item.id === bookId)
        if (book) {
          book.in_bookshelf = true
          book.bookshelf_id = shelfItem.id
        }
        return shelfItem
      } finally {
        this.joiningBookId = null
      }
    },
    resetFilters() {
      this.filters = { page: 1, page_size: 12, sort: 'created_at', order: 'desc' }
    }
  }
})
