import { apiClient, unwrap } from './client'
import type { Category } from './types'

export const categoryApi = {
  list() {
    return unwrap<Category[]>(apiClient.get('/categories'))
  },
  create(payload: { name: string; description?: string | null }) {
    return unwrap<Category>(apiClient.post('/admin/categories', payload))
  },
  update(id: number, payload: { name: string; description?: string | null }) {
    return unwrap<Category>(apiClient.patch(`/admin/categories/${id}`, payload))
  },
  remove(id: number) {
    return unwrap<Record<string, never>>(apiClient.delete(`/admin/categories/${id}`))
  }
}
