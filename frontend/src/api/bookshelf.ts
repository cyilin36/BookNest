import { apiClient, unwrap, unwrapPage } from './client'
import type { BookshelfItem, PageQuery } from './types'

export interface BookshelfQuery extends PageQuery {
  format?: 'epub' | 'pdf' | 'txt'
  category_id?: number
  tag_id?: number
  source_type?: 'uploaded' | 'library'
  favorite?: boolean
}

export interface UpdateBookshelfItemRequest {
  personal_title?: string | null
  personal_category_id?: number | null
  favorite?: boolean
  pinned?: boolean
  tag_ids?: number[]
}

export const bookshelfApi = {
  upload(formData: FormData) {
    return unwrap<BookshelfItem>(apiClient.post('/bookshelf/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } }))
  },
  list(query: BookshelfQuery = {}) {
    return unwrapPage<BookshelfItem>(apiClient.get('/bookshelf', { params: query }))
  },
  detail(id: number) {
    return unwrap<BookshelfItem>(apiClient.get(`/bookshelf/${id}`))
  },
  update(id: number, payload: UpdateBookshelfItemRequest) {
    return unwrap<BookshelfItem>(apiClient.patch(`/bookshelf/${id}`, payload))
  },
  remove(id: number) {
    return unwrap<Record<string, never>>(apiClient.delete(`/bookshelf/${id}`))
  },
  addLibraryBook(bookId: number) {
    return unwrap<BookshelfItem>(apiClient.post(`/bookshelf/from-library/${bookId}`))
  }
}
