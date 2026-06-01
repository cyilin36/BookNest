import { apiClient, unwrap } from './client'
import type { AuthSession, User } from './types'

export interface LoginRequest {
  login: string
  password: string
}

export interface RegisterRequest {
  username: string
  email?: string | null
  password: string
  nickname?: string | null
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UpdateMeRequest {
  email?: string | null
  nickname?: string | null
}

export interface RefreshRequest {
  refresh_token?: string
}

export const authApi = {
  login(payload: LoginRequest) {
    return unwrap<AuthSession>(apiClient.post('/auth/login', payload))
  },
  register(payload: RegisterRequest) {
    return unwrap<AuthSession>(apiClient.post('/auth/register', payload))
  },
  refresh(payload?: RefreshRequest) {
    return unwrap<AuthSession>(apiClient.post('/auth/refresh', payload, { skipAuthRefresh: true }))
  },
  logout() {
    return unwrap<Record<string, never>>(apiClient.post('/auth/logout', undefined, { skipAuthRefresh: true }))
  },
  me() {
    return unwrap<User>(apiClient.get('/auth/me'))
  },
  updateMe(payload: UpdateMeRequest) {
    return unwrap<User>(apiClient.patch('/users/me', payload))
  },
  changePassword(payload: ChangePasswordRequest) {
    return unwrap<Record<string, never>>(apiClient.patch('/users/me/password', payload))
  }
}
