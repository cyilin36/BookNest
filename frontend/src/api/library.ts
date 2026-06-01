import { apiClient, unwrap, unwrapPage } from './client'
import type { BookshelfItem, LibraryBook, PageQuery } from './types'

export interface LibraryQuery extends PageQuery {
  format?: 'epub' | 'pdf' | 'txt'
  category_id?: number
  tag_id?: number
  mine?: boolean
  status?: 'pending' | 'approved' | 'rejected' | 'hidden' | 'deleted'
}

export const libraryApi = {
  list(query: LibraryQuery = {}) {
    return unwrapPage<LibraryBook>(apiClient.get('/library/books', { params: query }))
  },
  detail(id: number) {
    return unwrap<LibraryBook>(apiClient.get(`/library/books/${id}`))
  },
  upload(formData: FormData) {
    return unwrap<LibraryBook>(apiClient.post('/library/books/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } }))
  },
  addToBookshelf(id: number) {
    return unwrap<BookshelfItem>(apiClient.post(`/library/books/${id}/add-to-bookshelf`))
  }
}
