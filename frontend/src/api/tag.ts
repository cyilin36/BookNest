import { apiClient, unwrap } from './client'
import type { Tag } from './types'

export const tagApi = {
  list() {
    return unwrap<Tag[]>(apiClient.get('/tags'))
  },
  create(payload: { name: string; description?: string | null }) {
    return unwrap<Tag>(apiClient.post('/admin/tags', payload))
  },
  update(id: number, payload: { name: string; description?: string | null }) {
    return unwrap<Tag>(apiClient.patch(`/admin/tags/${id}`, payload))
  },
  remove(id: number) {
    return unwrap<Record<string, never>>(apiClient.delete(`/admin/tags/${id}`))
  }
}
