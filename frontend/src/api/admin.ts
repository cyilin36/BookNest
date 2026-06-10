import { apiClient, unwrap, unwrapPage } from './client'
import type { AdminStorageStats, BookFormat, LibraryStatus, PageQuery, SystemIconUploadResult, SystemSettings, UpdateSystemSettingsRequest } from './types'
import type { LibraryBook } from './types'

export type AdminLibraryEditableStatus = Extract<LibraryStatus, 'approved' | 'hidden'>

export interface AdminLibraryQuery extends PageQuery {
  status?: LibraryStatus
  mine?: boolean
  format?: BookFormat
  category_id?: number
  tag_id?: number
}

export const adminApi = {
  library(query: AdminLibraryQuery = {}) {
    return unwrapPage<LibraryBook>(apiClient.get('/admin/library/books', { params: query }))
  },
  updateLibraryStatus(id: number, payload: { status: AdminLibraryEditableStatus; reason?: string | null }) {
    return unwrap<LibraryBook>(apiClient.patch(`/admin/library/books/${id}/status`, payload))
  },
  deleteLibraryBook(id: number) {
    return unwrap<Record<string, never>>(apiClient.delete(`/admin/library/books/${id}`))
  },
  storage() {
    return unwrap<AdminStorageStats>(apiClient.get('/admin/system/storage'))
  },
  settings() {
    return unwrap<SystemSettings>(apiClient.get('/admin/system/settings'))
  },
  updateSettings(payload: UpdateSystemSettingsRequest) {
    return unwrap<SystemSettings>(apiClient.put('/admin/system/settings', payload))
  },
  uploadSystemIcon(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return unwrap<SystemIconUploadResult>(apiClient.post('/admin/system/icon', formData, { headers: { 'Content-Type': 'multipart/form-data' } }))
  },
  deleteSystemIcon() {
    return unwrap<{ site_icon_url: null }>(apiClient.delete('/admin/system/icon'))
  }
}
