import { defineStore } from 'pinia'
import { authApi, type LoginRequest, type RegisterRequest } from '@/api/auth'
import { setAccessToken, getAccessToken, setAuthFailureHandler } from '@/api/client'
import type { User } from '@/api/types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: getAccessToken() as string | null,
    user: null as User | null,
    authReady: false,
    loading: false,
    error: ''
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.accessToken && state.user),
    isAdmin: (state) => state.user?.role === 'admin'
  },
  actions: {
    applySession(session: { access_token: string; user: User }) {
      this.accessToken = session.access_token
      this.user = session.user
      setAccessToken(session.access_token)
    },
    clearAuth() {
      this.accessToken = null
      this.user = null
      setAccessToken(null)
    },
    async register(payload: RegisterRequest) {
      this.loading = true
      this.error = ''
      try {
        this.applySession(await authApi.register(payload))
      } finally {
        this.loading = false
      }
    },
    async login(payload: LoginRequest) {
      this.loading = true
      this.error = ''
      try {
        this.applySession(await authApi.login(payload))
      } finally {
        this.loading = false
      }
    },
    async fetchMe() {
      this.user = await authApi.me()
      return this.user
    },
    async refresh() {
      const session = await authApi.refresh()
      this.applySession(session)
      return session
    },
    async restore() {
      if (this.authReady) return
      setAuthFailureHandler(() => this.clearAuth())
      try {
        if (this.accessToken) {
          try {
            await this.fetchMe()
          } catch {
            await this.refresh()
          }
        } else {
          await this.refresh()
        }
      } catch {
        this.clearAuth()
      } finally {
        this.authReady = true
      }
    },
    async logout() {
      try {
        await authApi.logout()
      } finally {
        this.clearAuth()
      }
    }
  }
})
