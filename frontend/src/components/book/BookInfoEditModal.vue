<script setup lang="ts">
import { reactive, watch } from 'vue'

export interface BookInfoEditPayload {
  title: string
  author: string | null
  description: string | null
}

const props = defineProps<{
  show: boolean
  title: string
  initial: BookInfoEditPayload | null
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  save: [payload: BookInfoEditPayload]
}>()

const form = reactive({
  title: '',
  author: '',
  description: ''
})

watch(
  () => [props.show, props.initial] as const,
  () => {
    if (!props.show || !props.initial) return
    form.title = props.initial.title
    form.author = props.initial.author || ''
    form.description = props.initial.description || ''
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
    description: form.description.trim() || null
  })
}
</script>

<template>
  <n-modal :show="show" preset="card" :title="title" class="book-info-modal" @update:show="emit('update:show', $event)">
    <n-form label-placement="top" @submit.prevent="submit">
      <n-form-item label="标题" required>
        <n-input v-model:value="form.title" placeholder="输入标题" maxlength="200" show-count />
      </n-form-item>
      <n-form-item label="作者">
        <n-input v-model:value="form.author" placeholder="可选" maxlength="120" show-count />
      </n-form-item>
      <n-form-item label="简介">
        <n-input v-model:value="form.description" type="textarea" placeholder="可选" :autosize="{ minRows: 4, maxRows: 8 }" maxlength="2000" show-count />
      </n-form-item>
      <div class="modal-actions">
        <n-button secondary :disabled="saving" @click="close">取消</n-button>
        <n-button type="primary" attr-type="submit" :loading="saving" :disabled="!form.title.trim()">保存</n-button>
      </div>
    </n-form>
  </n-modal>
</template>

<style scoped>
.book-info-modal {
  width: min(520px, calc(100vw - 32px));
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
