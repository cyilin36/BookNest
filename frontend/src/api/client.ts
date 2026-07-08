import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import router from '@/router'
import type { APIErrorPayload, AuthSession, PaginatedPayload, SuccessPayload } from './types'

const TOKEN_KEY = 'book-nest.access-token'
const LEGACY_TOKEN_KEY = 'book-reader.access-token'

export class AppAPIError extends Error {
  code: string
  requestId?: string
  status?: number
  details?: unknown

  constructor(message: string, code: string, requestId?: string, status?: number, details?: unknown) {
    super(message)
    this.name = 'AppAPIError'
    this.code = code
    this.requestId = requestId
    this.status = status
    this.details = details
  }
}

const legacyToken = sessionStorage.getItem(LEGACY_TOKEN_KEY)
if (!sessionStorage.getItem(TOKEN_KEY) && legacyToken) {
  sessionStorage.setItem(TOKEN_KEY, legacyToken)
  sessionStorage.removeItem(LEGACY_TOKEN_KEY)
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

function headerValue(value: unknown) {
  return Array.isArray(value) ? String(value[0] || '') : String(value || '')
}

async function parseBlobError(payload: Blob, contentTypeHeader: unknown) {
  const contentType = `${payload.type} ${headerValue(contentTypeHeader)}`.toLowerCase()
  if (!contentType.includes('application/json')) return null
  try {
    return JSON.parse(await payload.text()) as APIErrorPayload
  } catch {
    return null
  }
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<APIErrorPayload | Blob>) => {
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

    const responsePayload = error.response?.data
    const payload = responsePayload instanceof Blob ? await parseBlobError(responsePayload, error.response?.headers?.['content-type']) : responsePayload
    if (payload?.error) {
      throw new AppAPIError(payload.error.message, payload.error.code, payload.request_id, error.response?.status, payload.details)
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
