<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useSystemStore } from '@/stores/system'

const props = withDefaults(
  defineProps<{
    size?: number
    fallback?: string
  }>(),
  {
    size: 34,
    fallback: 'BN'
  }
)

const system = useSystemStore()
const failed = ref(false)
const iconSrc = computed(() => (failed.value ? '' : system.siteIconSrc))
const markStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  fontSize: `${Math.max(11, Math.round(props.size * 0.28))}px`
}))

watch(
  () => system.siteIconSrc,
  () => {
    failed.value = false
  }
)
</script>

<template>
  <div class="site-brand-mark" :class="{ 'has-image': iconSrc }" :style="markStyle">
    <img v-if="iconSrc" :src="iconSrc" :alt="system.siteName" @error="failed = true" />
    <span v-else>{{ fallback }}</span>
  </div>
</template>

<style scoped>
.site-brand-mark {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  overflow: hidden;
  color: #fff;
  background: var(--color-primary);
  border-radius: 8px;
  font-weight: 800;
}

.site-brand-mark.has-image {
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
}

.site-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>
