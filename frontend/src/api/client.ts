import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import router from '@/router'
import type { APIErrorPayload, AuthSession, PaginatedPayload, SuccessPayload } from './types'

const TOKEN_KEY = 'book-reader.access-token'

export class AppAPIError extends Error {
  code: string
  requestId?: string
  status?: number

  constructor(message: string, code: string, requestId?: string, status?: number) {
    super(message)
    this.name = 'AppAPIError'
    this.code = code
    this.requestId = requestId
    this.status = status
  }
}

let accessToken = sessionStorage.getItem(TOKEN_KEY)
let refreshPromise: Promise<string | null> | null = null
let authFailureHandler: (() => void) | null = null

export function getAccessToken() {
  return accessToken
}

export function setAccessToken(token: string | null) {
  accessToken = token
  if (token) {
    sessionStorage.setItem(TOKEN_KEY, token)
  } else {
    sessionStorage.removeItem(TOKEN_KEY)
  }
}

export function setAuthFailureHandler(handler: () => void) {
  authFailureHandler = handler
}

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
  withCredentials: true,
  timeout: 30000
})

apiClient.interceptors.request.use((config) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})

async function refreshAccessToken() {
  if (!refreshPromise) {
    refreshPromise = apiClient
      .post<SuccessPayload<AuthSession>>('/auth/refresh', undefined, { skipAuthRefresh: true } as AxiosRequestConfig)
      .then((response) => {
        const token = response.data.data.access_token
        setAccessToken(token)
        return token
      })
      .catch(() => {
        setAccessToken(null)
        authFailureHandler?.()
        return null
      })
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<APIErrorPayload>) => {
    const original = error.config as (AxiosRequestConfig & { _retry?: boolean; skipAuthRefresh?: boolean }) | undefined
    if (error.response?.status === 401 && original && !original._retry && !original.skipAuthRefresh) {
      original._retry = true
      const token = await refreshAccessToken()
      if (token) {
        original.headers = { ...(original.headers || {}), Authorization: `Bearer ${token}` }
        return apiClient(original)
      }
      if (router.currentRoute.value.meta.requiresAuth) {
        router.replace({ path: '/login', query: { redirect: router.currentRoute.value.fullPath } })
      }
    }

    const payload = error.response?.data
    if (payload?.error) {
      throw new AppAPIError(payload.error.message, payload.error.code, payload.request_id, error.response?.status)
    }
    throw new AppAPIError(error.message || '网络请求失败', 'network_error', undefined, error.response?.status)
  }
)

export async function unwrap<T>(request: Promise<{ data: SuccessPayload<T> }>) {
  const response = await request
  return response.data.data
}

export async function unwrapPage<T>(request: Promise<{ data: PaginatedPayload<T> }>) {
  const response = await request
  return response.data
}

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipAuthRefresh?: boolean
  }
}
