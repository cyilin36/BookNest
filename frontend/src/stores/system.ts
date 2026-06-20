import { defineStore } from 'pinia'
import { systemApi } from '@/api/system'
import type { BookFormat, SystemInfo, SystemSettings } from '@/api/types'

function faviconLink() {
  const existing = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (existing) return existing
  const link = document.createElement('link')
  link.rel = 'icon'
  document.head.appendChild(link)
  return link
}

function withCacheVersion(url: string, version: number) {
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}v=${version}`
}

export const useSystemStore = defineStore('system', {
  state: () => ({
    siteName: 'BookNest',
    siteIconUrl: null as string | null,
    siteIconVersion: 0,
    loginBackgroundUrl: null as string | null,
    loginBackgroundVersion: 0,
    allowRegistration: true,
    libraryReviewRequired: false,
    supportedFormats: ['epub', 'pdf', 'txt'] as BookFormat[],
    maxUploadSizeMb: 100,
    defaultUserStorageQuotaMb: 1024,
    loading: false
  }),
  getters: {
    siteIconSrc(state) {
      return state.siteIconUrl ? withCacheVersion(state.siteIconUrl, state.siteIconVersion) : ''
    },
    loginBackgroundSrc(state) {
      return state.loginBackgroundUrl ? withCacheVersion(state.loginBackgroundUrl, state.loginBackgroundVersion) : null
    }
  },
  actions: {
    syncFavicon() {
      faviconLink().href = this.siteIconSrc || '/favicon.ico'
    },
    applySystemInfo(info: SystemInfo | SystemSettings) {
      this.siteName = info.site_name
      this.siteIconUrl = info.site_icon_url
      this.loginBackgroundUrl = info.login_background_url
      this.allowRegistration = info.allow_registration
      this.libraryReviewRequired = info.library_review_required
      this.maxUploadSizeMb = info.max_upload_size_mb
      this.defaultUserStorageQuotaMb = info.default_user_storage_quota_mb
      if ('supported_formats' in info) {
        this.supportedFormats = info.supported_formats
      }
      this.syncFavicon()
    },
    setSiteIconUrl(url: string | null) {
      this.siteIconUrl = url
      this.siteIconVersion = Date.now()
      this.syncFavicon()
    },
    setLoginBackgroundUrl(url: string | null) {
      this.loginBackgroundUrl = url
      this.loginBackgroundVersion = Date.now()
    },
    async fetchSystemInfo() {
      this.loading = true
      try {
        const info = await systemApi.info()
        this.applySystemInfo(info)
      } finally {
        this.loading = false
      }
    }
  }
})
