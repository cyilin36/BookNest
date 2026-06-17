<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { darkTheme, zhCN, dateZhCN, type GlobalThemeOverrides } from 'naive-ui'
import { RouterView } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { useSystemStore } from '@/stores/system'

const settings = useSettingsStore()
const system = useSystemStore()

onMounted(() => {
  settings.loadLocalSettings()
  system.fetchSystemInfo().catch(() => undefined)
})

const naiveTheme = computed(() => (settings.reader.theme === 'dark' ? darkTheme : null))
const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    primaryColor: settings.reader.theme === 'sepia' ? '#8B5A2B' : settings.reader.theme === 'dark' ? '#10B981' : '#4A90E2',
    primaryColorHover: settings.reader.theme === 'sepia' ? '#A06D3B' : settings.reader.theme === 'dark' ? '#34D399' : '#3A7BC8',
    primaryColorPressed: settings.reader.theme === 'sepia' ? '#7B4A1B' : settings.reader.theme === 'dark' ? '#059669' : '#2A6BB8',
    primaryColorSuppl: settings.reader.theme === 'sepia' ? 'rgba(139, 90, 43, 0.1)' : settings.reader.theme === 'dark' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(74, 144, 226, 0.1)',
    borderRadius: '12px',
    borderRadiusSmall: '8px',
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", "Helvetica Neue", Arial, sans-serif',
    heightMedium: '44px',
    heightSmall: '32px',
    fontSize: '14px',
    fontSizeMedium: '15px',
    lineHeight: '1.6'
  },
  Button: {
    borderRadiusMedium: '10px',
    borderRadiusSmall: '8px',
    heightMedium: '44px',
    heightSmall: '32px',
    paddingMedium: '0 20px',
    paddingSmall: '0 14px',
    fontSizeMedium: '15px',
    fontSizeSmall: '13px',
    fontWeightStrong: '500'
  },
  Input: {
    borderRadius: '12px',
    heightMedium: '44px',
    heightSmall: '36px',
    heightLarge: '52px',
    fontSizeMedium: '16px',
    paddingMedium: '0 20px'
  },
  Select: {
    borderRadius: '12px',
    heightMedium: '44px'
  },
  Card: {
    borderRadius: '12px'
  },
  Drawer: {
    borderRadius: '0px'
  }
}))
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="naiveTheme" :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <RouterView />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
