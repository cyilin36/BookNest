<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'
import PageShell from '@/components/common/PageShell.vue'
import type { ReaderLineHeight, ThemeName } from '@/api/types'

const settings = useSettingsStore()

function updateTheme(value: ThemeName) {
  settings.setTheme(value)
}

function updateLineHeight(value: ReaderLineHeight) {
  settings.updateReaderSettings({ line_height: value })
}

function updateFontSize(value: number) {
  settings.updateReaderSettings({ font_size: value })
}
</script>

<template>
  <PageShell title="阅读设置" subtitle="本地保存主题、字号、行距和字体">
    <div class="surface settings-card">
      <n-select v-model:value="settings.reader.theme" :options="[{ label: 'Modern', value: 'modern' }, { label: 'Sepia', value: 'sepia' }, { label: 'Dark', value: 'dark' }]" @update:value="updateTheme" />
      <n-slider v-model:value="settings.reader.font_size" :min="14" :max="26" :step="1" @update:value="updateFontSize" />
      <n-select
        v-model:value="settings.reader.line_height"
        :options="[
          { label: '紧凑 1.5', value: 1.5 },
          { label: '适中 1.8', value: 1.8 },
          { label: '宽松 2.2', value: 2.2 }
        ]"
        @update:value="updateLineHeight"
      />
      <n-button secondary @click="settings.resetReaderSettings()">恢复默认</n-button>
    </div>
  </PageShell>
</template>

<style scoped>
.settings-card {
  display: grid;
  gap: 18px;
  max-width: 540px;
  padding: 18px;
}
</style>
