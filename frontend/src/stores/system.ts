import { defineStore } from 'pinia'
import { systemApi } from '@/api/system'
import type { BookFormat } from '@/api/types'

export const useSystemStore = defineStore('system', {
  state: () => ({
    siteName: 'BookNest',
    allowRegistration: true,
    libraryReviewRequired: false,
    supportedFormats: ['epub', 'pdf', 'txt'] as BookFormat[],
    maxUploadSizeMb: 100,
    defaultUserStorageQuotaMb: 1024,
    loading: false
  }),
  actions: {
    async fetchSystemInfo() {
      this.loading = true
      try {
        const info = await systemApi.info()
        this.siteName = info.site_name
        this.allowRegistration = info.allow_registration
        this.libraryReviewRequired = info.library_review_required
        this.supportedFormats = info.supported_formats
        this.maxUploadSizeMb = info.max_upload_size_mb
        this.defaultUserStorageQuotaMb = info.default_user_storage_quota_mb
      } finally {
        this.loading = false
      }
    }
  }
})
