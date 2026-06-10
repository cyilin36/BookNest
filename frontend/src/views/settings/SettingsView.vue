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

function updateContentWidth(value: number) {
  settings.updateReaderSettings({ content_width: value })
}
</script>

<template>
  <PageShell title="阅读设置" subtitle="本地保存主题、字号、行距、宽度和字体">
    <div class="surface settings-card">
      <n-select v-model:value="settings.reader.theme" :options="[{ label: '现代', value: 'modern' }, { label: '纸页', value: 'sepia' }, { label: '深色', value: 'dark' }]" @update:value="updateTheme" />
      <n-slider v-model:value="settings.reader.font_size" :min="14" :max="26" :step="1" :tooltip="false" @update:value="updateFontSize" />
      <div class="setting-group">
        <div class="setting-label">
          <span>阅读宽度</span>
        </div>
        <n-slider v-model:value="settings.reader.content_width" :min="0" :max="1180" :step="20" :tooltip="false" @update:value="updateContentWidth" />
        <div class="width-presets">
          <n-button size="small" secondary @click="updateContentWidth(640)">窄</n-button>
          <n-button size="small" secondary @click="updateContentWidth(760)">适中</n-button>
          <n-button size="small" secondary @click="updateContentWidth(920)">宽</n-button>
          <n-button size="small" secondary @click="updateContentWidth(0)">铺满</n-button>
        </div>
      </div>
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

.setting-group {
  display: grid;
  gap: 8px;
}

.setting-label {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--color-text-sec);
  font-size: 13px;
}

.setting-label strong {
  color: var(--color-text-main);
}

.width-presets {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 6px;
}
</style>
