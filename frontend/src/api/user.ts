import { apiClient, unwrap, unwrapPage } from './client'
import type { PageQuery, User } from './types'

export interface UsersQuery extends PageQuery {
  status?: 'active' | 'disabled'
  role?: 'admin' | 'user'
}

export interface UpdateMeRequest {
  email?: string | null
  nickname?: string | null
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface SetDefaultAvatarRequest {
  avatar_name: string
}

export interface AdminUpdateUserRequest {
  email?: string | null
  nickname?: string | null
  storage_quota_bytes?: number | null
}

export const userApi = {
  me() {
    return unwrap<User>(apiClient.get('/users/me'))
  },
  updateMe(payload: UpdateMeRequest) {
    return unwrap<User>(apiClient.patch('/users/me', payload))
  },
  changePassword(payload: ChangePasswordRequest) {
    return unwrap<Record<string, never>>(apiClient.patch('/users/me/password', payload))
  },
  uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return unwrap<User>(apiClient.post('/users/me/avatar/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } }))
  },
  setDefaultAvatar(payload: SetDefaultAvatarRequest) {
    return unwrap<User>(apiClient.post('/users/me/avatar/default', payload))
  },
  list(query: UsersQuery = {}) {
    return unwrapPage<User>(apiClient.get('/admin/users', { params: query }))
  },
  detail(id: number) {
    return unwrap<User>(apiClient.get(`/admin/users/${id}`))
  },
  update(id: number, payload: AdminUpdateUserRequest) {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}`, payload))
  },
  updateStatus(id: number, status: 'active' | 'disabled') {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}/status`, { status }))
  },
  updateRole(id: number, role: 'admin' | 'user') {
    return unwrap<User>(apiClient.patch(`/admin/users/${id}/role`, { role }))
  },
  remove(id: number) {
    return unwrap<Record<string, never>>(apiClient.delete(`/admin/users/${id}`))
  }
}
