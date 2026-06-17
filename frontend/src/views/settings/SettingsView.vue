<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'
import PageShell from '@/components/common/PageShell.vue'
import type { ReaderLineHeight, ReaderMode, ThemeName } from '@/api/types'

const settings = useSettingsStore()

function updateTheme(value: ThemeName) {
  settings.setTheme(value)
}

function updateLineHeight(value: ReaderLineHeight) {
  settings.updateReaderSettings({ line_height: value })
}

function updateReadingMode(value: ReaderMode) {
  settings.updateReaderSettings({ reading_mode: value })
}

function updateFontSize(value: number) {
  settings.updateReaderSettings({ font_size: value })
}

function updateContentWidth(value: number) {
  settings.updateReaderSettings({ content_width: value })
}
</script>

<template>
  <PageShell title="阅读设置" subtitle="主题、字号、行距、宽度和阅读模式">
    <div class="settings-container">
      <div class="setting-card surface">
        <div class="setting-section">
          <div class="setting-header">
            <h3 class="setting-title">主题</h3>
            <p class="setting-description">选择您喜欢的阅读主题</p>
          </div>
          <n-select
            v-model:value="settings.reader.theme"
            size="large"
            :options="[
              { label: '现代', value: 'modern' },
              { label: '纸页', value: 'sepia' },
              { label: '深色', value: 'dark' }
            ]"
            @update:value="updateTheme"
          />
        </div>

        <n-divider />

        <div class="setting-section">
          <div class="setting-header">
            <h3 class="setting-title">阅读模式</h3>
            <p class="setting-description">滚动或分页翻阅</p>
          </div>
          <n-radio-group v-model:value="settings.reader.reading_mode" size="large" @update:value="updateReadingMode">
            <n-space>
              <n-radio value="scroll">滚动</n-radio>
              <n-radio value="page">分页</n-radio>
            </n-space>
          </n-radio-group>
        </div>

        <n-divider />

        <div class="setting-section">
          <div class="setting-header">
            <h3 class="setting-title">字号</h3>
            <p class="setting-description">当前：{{ settings.reader.font_size }}px</p>
          </div>
          <n-slider
            v-model:value="settings.reader.font_size"
            :min="14"
            :max="26"
            :step="1"
            :marks="{ 14: '小', 20: '中', 26: '大' }"
            @update:value="updateFontSize"
          />
        </div>

        <n-divider />

        <div class="setting-section">
          <div class="setting-header">
            <h3 class="setting-title">行距</h3>
            <p class="setting-description">调整文字行间距</p>
          </div>
          <n-select
            v-model:value="settings.reader.line_height"
            size="large"
            :options="[
              { label: '紧凑 1.5', value: 1.5 },
              { label: '适中 1.8', value: 1.8 },
              { label: '宽松 2.2', value: 2.2 }
            ]"
            @update:value="updateLineHeight"
          />
        </div>

        <n-divider />

        <div class="setting-section">
          <div class="setting-header">
            <h3 class="setting-title">阅读宽度</h3>
            <p class="setting-description">{{ settings.reader.content_width === 0 ? '铺满全屏' : `${settings.reader.content_width}px` }}</p>
          </div>
          <n-slider
            v-model:value="settings.reader.content_width"
            :min="0"
            :max="1180"
            :step="20"
            :marks="{ 0: '铺满', 640: '窄', 760: '适中', 920: '宽' }"
            @update:value="updateContentWidth"
          />
          <div class="width-presets">
            <n-button size="medium" secondary @click="updateContentWidth(640)">窄</n-button>
            <n-button size="medium" secondary @click="updateContentWidth(760)">适中</n-button>
            <n-button size="medium" secondary @click="updateContentWidth(920)">宽</n-button>
            <n-button size="medium" secondary @click="updateContentWidth(0)">铺满</n-button>
          </div>
        </div>

        <n-divider />

        <div class="setting-section">
          <n-button size="large" secondary block @click="settings.resetReaderSettings()">
            恢复默认设置
          </n-button>
        </div>
      </div>
    </div>
  </PageShell>
</template>

<style scoped>
.settings-container {
  max-width: 680px;
  margin: 0 auto;
}

.setting-card {
  padding: var(--spacing-3xl);
  border-radius: var(--radius-xlarge);
  box-shadow: var(--shadow-medium);
}

.setting-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.setting-header {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.setting-title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-main);
  letter-spacing: -0.01em;
}

.setting-description {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
}

.setting-section :deep(.n-divider) {
  margin: var(--spacing-2xl) 0;
}

.setting-section :deep(.n-select),
.setting-section :deep(.n-radio-group) {
  width: 100%;
}

.setting-section :deep(.n-slider) {
  margin: var(--spacing-md) 0;
}

.width-presets {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-md);
  margin-top: var(--spacing-sm);
}

.width-presets :deep(.n-button) {
  transition: all var(--transition-base);
}

.width-presets :deep(.n-button:hover:not(:disabled)) {
  transform: translateY(-2px);
}

@media (max-width: 768px) {
  .setting-card {
    padding: var(--spacing-2xl);
  }

  .setting-title {
    font-size: var(--font-size-md);
  }

  .width-presets {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-sm);
  }
}
</style>
