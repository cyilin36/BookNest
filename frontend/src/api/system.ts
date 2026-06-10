import { apiClient, unwrap } from './client'
import type { SystemInfo, SystemSettings, UpdateSystemSettingsRequest } from './types'

export const systemApi = {
  info() {
    return unwrap<SystemInfo>(apiClient.get('/system/info'))
  },
  adminSettings() {
    return unwrap<SystemSettings>(apiClient.get('/admin/system/settings'))
  },
  updateAdminSettings(payload: UpdateSystemSettingsRequest) {
    return unwrap<SystemSettings>(apiClient.put('/admin/system/settings', payload))
  }
}
