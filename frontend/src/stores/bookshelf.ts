import { defineStore } from 'pinia'
import { bookshelfApi, type BookshelfQuery } from '@/api/bookshelf'
import type { BookshelfItem, Pagination } from '@/api/types'

export const useBookshelfStore = defineStore('bookshelf', {
  state: () => ({
    items: [] as BookshelfItem[],
    filters: { page: 1, page_size: 12, sort: '', order: 'desc' as const } as BookshelfQuery,
    pagination: { page: 1, page_size: 12, total: 0 } as Pagination,
    loading: false,
    error: ''
  }),
  actions: {
    async fetchBookshelf(extra: BookshelfQuery = {}) {
      this.loading = true
      this.error = ''
      try {
        this.filters = { ...this.filters, ...extra }
        const response = await bookshelfApi.list(this.filters)
        this.items = response.data
        this.pagination = response.pagination
      } catch (error) {
        this.error = error instanceof Error ? error.message : '书架加载失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    async updateBookshelfItem(id: number, payload: Parameters<typeof bookshelfApi.update>[1]) {
      const item = await bookshelfApi.update(id, payload)
      const index = this.items.findIndex((row) => row.id === id)
      if (index >= 0) this.items[index] = item
      return item
    },
    async removeBookshelfItem(id: number) {
      await bookshelfApi.remove(id)
      this.items = this.items.filter((row) => row.id !== id)
      this.pagination.total = Math.max(0, this.pagination.total - 1)
    },
    resetFilters() {
      this.filters = { page: 1, page_size: 12, sort: '', order: 'desc' }
    }
  }
})
