<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import CoverUploadField from './CoverUploadField.vue'

export interface BookInfoEditPayload {
  title: string
  author: string | null
  description: string | null
  category_ids?: number[] | null
  tag_ids?: number[] | null
  cover?: File | null
}

const props = defineProps<{
  show: boolean
  title: string
  initial: BookInfoEditPayload | null
  categoryOptions?: { label: string; value: number }[]
  tagOptions?: { label: string; value: number }[]
  taxonomyLoading?: boolean
  singleCategory?: boolean
  saving?: boolean
  allowCover?: boolean
  currentCoverUrl?: string | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  save: [payload: BookInfoEditPayload]
}>()

const cover = ref<File | null>(null)

const form = reactive({
  title: '',
  author: '',
  description: '',
  category_ids: [] as number[],
  tag_ids: [] as number[]
})

const categoryValue = computed({
  get() {
    return props.singleCategory ? (form.category_ids[0] ?? null) : form.category_ids
  },
  set(value: number | number[] | null) {
    form.category_ids = Array.isArray(value) ? value : value ? [value] : []
  }
})

watch(
  () => [props.show, props.initial] as const,
  () => {
    if (!props.show || !props.initial) return
    form.title = props.initial.title
    form.author = props.initial.author || ''
    form.description = props.initial.description || ''
    form.category_ids = props.initial.category_ids || []
    form.tag_ids = props.initial.tag_ids || []
    cover.value = null
  },
  { immediate: true }
)

function close() {
  if (props.saving) return
  emit('update:show', false)
}

function submit() {
  const title = form.title.trim()
  if (!title) return
  emit('save', {
    title,
    author: form.author.trim() || null,
    description: form.description.trim() || null,
    category_ids: props.categoryOptions ? form.category_ids : undefined,
    tag_ids: props.tagOptions ? form.tag_ids : undefined,
    cover: props.allowCover ? cover.value : undefined
  })
}
</script>

<template>
  <n-modal :show="show" preset="card" :title="title" class="book-info-modal" @update:show="emit('update:show', $event)">
    <n-form label-placement="top" class="modal-form" @submit.prevent="submit">
      <n-form-item label="标题" required>
        <n-input v-model:value="form.title" size="large" placeholder="输入标题" maxlength="200" show-count />
      </n-form-item>
      <n-form-item label="作者">
        <n-input v-model:value="form.author" size="large" placeholder="可选" maxlength="120" show-count />
      </n-form-item>
      <n-form-item label="简介">
        <n-input v-model:value="form.description" type="textarea" placeholder="可选" :autosize="{ minRows: 4, maxRows: 8 }" maxlength="2000" show-count />
      </n-form-item>
      <n-form-item v-if="categoryOptions" label="分类">
        <n-select
          v-model:value="categoryValue"
          size="large"
          :multiple="!singleCategory"
          clearable
          :loading="taxonomyLoading"
          :options="categoryOptions"
          placeholder="可选"
        />
      </n-form-item>
      <n-form-item v-if="tagOptions" label="标签">
        <n-select v-model:value="form.tag_ids" size="large" multiple clearable :loading="taxonomyLoading" :options="tagOptions" placeholder="可选" />
      </n-form-item>
      <n-form-item v-if="allowCover" label="封面">
        <CoverUploadField v-model:file="cover" :current-url="currentCoverUrl" :disabled="saving" />
      </n-form-item>
      <n-form-item v-if="allowCover" label="封面">
        <CoverUploadField v-model:file="cover" :current-url="currentCoverUrl" :disabled="saving" />
      </n-form-item>
      <div class="modal-actions">
        <n-button size="large" secondary :disabled="saving" @click="close">取消</n-button>
        <n-button size="large" type="primary" attr-type="submit" :loading="saving" :disabled="!form.title.trim()">保存</n-button>
      </div>
    </n-form>
  </n-modal>
</template>

<style scoped>
.book-info-modal {
  width: min(580px, calc(100vw - 32px));
}

.modal-form :deep(.n-form-item-label) {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-main);
  margin-bottom: var(--spacing-sm);
}

.modal-form :deep(.n-input),
.modal-form :deep(.n-select) {
  border-radius: var(--radius-medium);
}

.modal-form :deep(.n-input:focus-within),
.modal-form :deep(.n-select:focus-within) {
  box-shadow: 0 0 0 4px var(--color-primary-light);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
  margin-top: var(--spacing-xl);
}

.modal-actions :deep(.n-button) {
  min-width: 100px;
  transition: all var(--transition-base);
}

.modal-actions :deep(.n-button:hover:not(:disabled)) {
  transform: translateY(-1px);
}

.modal-actions :deep(.n-button--primary-type) {
  box-shadow: var(--shadow-button-primary);
}

.modal-actions :deep(.n-button--primary-type:hover:not(:disabled)) {
  box-shadow: var(--shadow-button-primary-hover);
}
</style>
