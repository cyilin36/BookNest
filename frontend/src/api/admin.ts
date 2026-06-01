import { apiClient, unwrap, unwrapPage } from './client'
import type { AdminStorageStats, LibraryStatus, PageQuery, SystemSettings } from './types'
import type { LibraryBook } from './types'

export interface AdminLibraryQuery extends PageQuery {
  status?: LibraryStatus
  mine?: boolean
  format?: 'epub' | 'pdf' | 'txt'
  category_id?: number
  tag_id?: number
}

export const adminApi = {
  library(query: AdminLibraryQuery = {}) {
    return unwrapPage<LibraryBook>(apiClient.get('/admin/library/books', { params: query }))
  },
  updateLibraryStatus(id: number, payload: { status: string; reason?: string | null }) {
    return unwrap<LibraryBook>(apiClient.patch(`/admin/library/books/${id}/status`, payload))
  },
  deleteLibraryBook(id: number, deleteFile = false) {
    return unwrap<Record<string, never>>(apiClient.delete(`/admin/library/books/${id}`, { params: { delete_file: deleteFile } }))
  },
  storage() {
    return unwrap<AdminStorageStats>(apiClient.get('/admin/system/storage'))
  },
  settings() {
    return unwrap<SystemSettings>(apiClient.get('/admin/system/settings'))
  },
  updateSettings(payload: SystemSettings) {
    return unwrap<SystemSettings>(apiClient.put('/admin/system/settings', payload))
  }
}
