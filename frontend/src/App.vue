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
    primaryColor: settings.reader.theme === 'sepia' ? '#8b5a2b' : settings.reader.theme === 'dark' ? '#10b981' : '#18a058',
    primaryColorHover: settings.reader.theme === 'sepia' ? '#a06d3b' : settings.reader.theme === 'dark' ? '#34d399' : '#36ad6a',
    borderRadius: '8px',
    fontFamily: 'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
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
