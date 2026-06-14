<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ImagePlus, X } from 'lucide-vue-next'
import BookCover from './BookCover.vue'

const props = defineProps<{
  file: File | null
  currentUrl?: string | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:file': [file: File | null]
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const previewUrl = ref<string | null>(null)
const error = ref('')

const maxBytes = 5 * 1024 * 1024
const allowedExtensions = ['png', 'jpg', 'jpeg', 'webp']
const allowedTypes = ['image/png', 'image/jpeg', 'image/webp']
const accept = '.png,.jpg,.jpeg,.webp,image/png,image/jpeg,image/webp'
const limitText = '支持 PNG、JPG、WebP，最大 5MB。'

const fileLabel = computed(() => props.file?.name || '选择封面图片')

function revokePreview() {
  if (previewUrl.value?.startsWith('blob:')) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = null
}

function resetInput() {
  if (inputRef.value) inputRef.value.value = ''
}

function validate(file: File) {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const typeAllowed = !file.type || allowedTypes.includes(file.type)
  const extensionAllowed = allowedExtensions.includes(extension)
  if (!extensionAllowed || !typeAllowed) return `封面格式不支持，${limitText}`
  if (file.size > maxBytes) return `封面图片过大，最大允许 5MB。当前文件约 ${(file.size / 1024 / 1024).toFixed(2)}MB。`
  return ''
}

function chooseFile(files: FileList | null) {
  const file = files?.[0]
  error.value = ''
  if (!file) return
  const validationError = validate(file)
  if (validationError) {
    error.value = validationError
    emit('update:file', null)
    resetInput()
    return
  }
  emit('update:file', file)
}

function clearFile() {
  error.value = ''
  emit('update:file', null)
  resetInput()
}

watch(
  () => props.file,
  (file) => {
    revokePreview()
    if (file) previewUrl.value = URL.createObjectURL(file)
  },
  { immediate: true }
)

onBeforeUnmount(revokePreview)
</script>

<template>
  <div class="cover-upload-field">
    <div class="cover-preview" :class="{ empty: !previewUrl && !currentUrl }">
      <img v-if="previewUrl" :src="previewUrl" alt="封面预览" />
      <BookCover v-else-if="currentUrl" :src="currentUrl" title="当前封面" />
      <ImagePlus v-else :size="28" />
    </div>
    <div class="cover-upload-main">
      <div class="cover-upload-actions">
        <n-button secondary :disabled="disabled" @click="inputRef?.click()">
          <template #icon><ImagePlus :size="16" /></template>
          {{ fileLabel }}
        </n-button>
        <n-button v-if="file" quaternary circle :disabled="disabled" title="移除已选封面" @click="clearFile">
          <X :size="16" />
        </n-button>
      </div>
      <p>{{ limitText }}</p>
      <p v-if="error" class="cover-upload-error">{{ error }}</p>
      <input ref="inputRef" class="sr-only" type="file" :accept="accept" :disabled="disabled" @change="chooseFile(($event.target as HTMLInputElement).files)" />
    </div>
  </div>
</template>

<style scoped>
.cover-upload-field {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cover-preview {
  display: grid;
  width: 72px;
  aspect-ratio: 3 / 4;
  flex: 0 0 auto;
  place-items: center;
  overflow: hidden;
  color: var(--color-text-sec);
  background: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.cover-preview.empty {
  border-style: dashed;
}

.cover-preview img,
.cover-preview :deep(.book-cover) {
  width: 100%;
  height: 100%;
}

.cover-preview img {
  object-fit: cover;
}

.cover-preview :deep(.book-cover) {
  border: 0;
  border-radius: 0;
}

.cover-upload-main {
  display: grid;
  min-width: 0;
  gap: 6px;
}

.cover-upload-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cover-upload-main p {
  margin: 0;
  color: var(--color-text-sec);
  font-size: 13px;
}

.cover-upload-main .cover-upload-error {
  color: #d03050;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 520px) {
  .cover-upload-field {
    align-items: flex-start;
  }

  .cover-upload-actions {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
