import { apiClient, unwrap, unwrapPage } from './client'
import type { PageQuery, User } from './types'

export interface UsersQuery extends PageQuery {
  status?: 'active' | 'disabled'
  role?: 'admin' | 'user'
}

export const userApi = {
  me() {
    return unwrap<User>(apiClient.get('/users/me'))
  },
  list(query: UsersQuery = {}) {
    return unwrapPage<User>(apiClient.get('/admin/users', { params: query }))
  },
  detail(id: number) {
    return unwrap<User>(apiClient.get(`/admin/users/${id}`))
  },
  update(id: number, payload: Partial<User>) {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}`, payload))
  },
  updateStatus(id: number, status: 'active' | 'disabled') {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}/status`, { status }))
  },
  updateRole(id: number, role: 'admin' | 'user') {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}/role`, { role }))
  }
}
