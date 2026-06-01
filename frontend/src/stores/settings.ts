import { defineStore } from 'pinia'
import type { ReaderSettings, ReaderLineHeight, ThemeName } from '@/api/types'

const KEY = 'book-reader.settings'

const defaults: ReaderSettings = {
  theme: 'modern',
  font_size: window.matchMedia('(max-width: 720px)').matches ? 16 : 18,
  line_height: 1.8,
  font_family: 'sans'
}

function readSettings(): ReaderSettings {
  try {
    return { ...defaults, ...JSON.parse(localStorage.getItem(KEY) || '{}') }
  } catch {
    return defaults
  }
}

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    reader: readSettings()
  }),
  getters: {
    themeClass: (state) => `theme-${state.reader.theme}`
  },
  actions: {
    persist() {
      localStorage.setItem(KEY, JSON.stringify(this.reader))
      const root = document.documentElement
      root.classList.remove('theme-modern', 'theme-sepia', 'theme-dark')
      root.classList.add(`theme-${this.reader.theme}`)
    },
    loadLocalSettings() {
      this.reader = readSettings()
      this.persist()
    },
    setTheme(theme: ThemeName) {
      this.reader.theme = theme
      this.persist()
    },
    updateReaderSettings(partial: Partial<ReaderSettings>) {
      this.reader = { ...this.reader, ...partial }
      this.persist()
    },
    resetReaderSettings() {
      this.reader = { ...defaults }
      this.persist()
    }
  }
})
