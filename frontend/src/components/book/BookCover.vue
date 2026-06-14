<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { BookOpenText } from 'lucide-vue-next'
import { apiClient } from '@/api/client'

const props = defineProps<{
  src?: string | null
  title: string
  revision?: number | string
}>()

const imageUrl = ref<string | null>(null)
const failed = ref(false)

function revokeObjectUrl() {
  if (imageUrl.value?.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value)
  }
}

function apiPathFromURL(rawURL: string) {
  const url = new URL(rawURL, window.location.origin)
  const base = import.meta.env.VITE_API_BASE || '/api/v1'
  if (url.pathname.startsWith(base)) {
    return `${url.pathname.slice(base.length) || '/'}${url.search}`
  }
  return `${url.pathname}${url.search}`
}

async function loadProtectedImage(src: string) {
  revokeObjectUrl()
  failed.value = false
  imageUrl.value = null
  try {
    const response = await apiClient.get<Blob>(apiPathFromURL(src), { responseType: 'blob' })
    imageUrl.value = URL.createObjectURL(response.data)
  } catch {
    failed.value = true
  }
}

const shouldShowFallback = computed(() => !props.src || failed.value)

watch(
  () => [props.src, props.revision] as const,
  ([src]) => {
    if (src) {
      loadProtectedImage(src)
    } else {
      revokeObjectUrl()
      imageUrl.value = null
      failed.value = false
    }
  },
  { immediate: true }
)

onBeforeUnmount(revokeObjectUrl)
</script>

<template>
  <div class="book-cover">
    <img v-if="imageUrl && !failed" :src="imageUrl" :alt="title" @error="failed = true" />
    <div v-if="shouldShowFallback" class="cover-fallback">
      <BookOpenText :size="32" />
      <span>{{ title.slice(0, 12) }}</span>
    </div>
  </div>
</template>

<style scoped>
.book-cover {
  position: relative;
  aspect-ratio: 3 / 4;
  overflow: hidden;
  background: linear-gradient(145deg, rgba(24, 160, 88, 0.18), rgba(31, 41, 55, 0.08));
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.book-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  display: grid;
  height: 100%;
  place-items: center;
  align-content: center;
  gap: 10px;
  padding: 14px;
  color: var(--color-text-sec);
  text-align: center;
}

.cover-fallback span {
  color: var(--color-text-main);
  font-weight: 700;
  line-height: 1.35;
}
</style>
