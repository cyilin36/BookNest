import { apiClient, unwrap, unwrapPage } from './client'
import type { BookFormat, BookshelfItem, LibraryBook, LibraryStatus, PageQuery } from './types'

export interface LibraryQuery extends PageQuery {
  format?: BookFormat
  category_id?: number
  tag_id?: number
  mine?: boolean
  status?: Extract<LibraryStatus, 'approved' | 'hidden'>
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
  hide(id: number) {
    return unwrap<LibraryBook>(apiClient.post(`/library/books/${id}/hide`))
  },
  show(id: number) {
    return unwrap<LibraryBook>(apiClient.post(`/library/books/${id}/show`))
  },
  addToBookshelf(id: number) {
    return unwrap<BookshelfItem>(apiClient.post(`/library/books/${id}/add-to-bookshelf`))
  }
}
