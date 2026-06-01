<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { Plus, RefreshCw, Save, Trash2 } from 'lucide-vue-next'
import type { Category, Tag } from '@/api/types'

type TaxonomyItem = Category | Tag

const props = defineProps<{
  title: string
  description: string
  loadItems: () => Promise<TaxonomyItem[]>
  createItem: (payload: { name: string; description?: string | null }) => Promise<TaxonomyItem>
  updateItem: (id: number, payload: { name: string; description?: string | null }) => Promise<TaxonomyItem>
  removeItem: (id: number) => Promise<Record<string, never>>
}>()

const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const saving = ref(false)
const items = ref<TaxonomyItem[]>([])
const editingId = ref<number | null>(null)
const form = reactive({ name: '', description: '' })

const editing = computed(() => items.value.find((item) => item.id === editingId.value) || null)

function resetForm() {
  editingId.value = null
  form.name = ''
  form.description = ''
}

function edit(item: TaxonomyItem) {
  editingId.value = item.id
  form.name = item.name
  form.description = item.description || ''
}

async function refresh() {
  loading.value = true
  try {
    items.value = await props.loadItems()
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.name.trim()) {
    message.warning('名称不能为空')
    return
  }
  saving.value = true
  try {
    const payload = { name: form.name.trim(), description: form.description.trim() || null }
    if (editing.value) {
      const updated = await props.updateItem(editing.value.id, payload)
      const index = items.value.findIndex((item) => item.id === updated.id)
      if (index >= 0) items.value[index] = updated
      message.success('已更新')
    } else {
      items.value.unshift(await props.createItem(payload))
      message.success('已创建')
    }
    resetForm()
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function confirmRemove(item: TaxonomyItem) {
  dialog.warning({
    title: `删除「${item.name}」`,
    content: '已被图书或书架使用的项目不能删除。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await props.removeItem(item.id)
        items.value = items.value.filter((row) => row.id !== item.id)
        if (editingId.value === item.id) resetForm()
        message.success('已删除')
      } catch (error) {
        message.error(error instanceof Error ? error.message : '删除失败')
      }
    }
  })
}

onMounted(refresh)
</script>

<template>
  <div class="taxonomy-grid">
    <section class="surface taxonomy-form">
      <div class="form-title">
        <div>
          <h2>{{ editing ? '编辑' : '新建' }}{{ title }}</h2>
          <p>{{ description }}</p>
        </div>
        <n-button quaternary circle title="刷新" @click="refresh">
          <RefreshCw :size="16" />
        </n-button>
      </div>
      <n-form label-placement="top" @submit.prevent="submit">
        <n-form-item label="名称">
          <n-input v-model:value="form.name" placeholder="输入名称" />
        </n-form-item>
        <n-form-item label="描述">
          <n-input v-model:value="form.description" type="textarea" placeholder="可选" />
        </n-form-item>
        <div class="toolbar">
          <n-button type="primary" attr-type="submit" :loading="saving">
            <template #icon><Save :size="16" /></template>
            保存
          </n-button>
          <n-button secondary @click="resetForm">
            <template #icon><Plus :size="16" /></template>
            新建
          </n-button>
        </div>
      </n-form>
    </section>
    <section class="surface taxonomy-list">
      <n-spin :show="loading">
        <div v-if="items.length" class="chip-list">
          <button v-for="item in items" :key="item.id" type="button" class="taxonomy-chip" :class="{ active: item.id === editingId }" @click="edit(item)">
            <span>{{ item.name }}</span>
            <small>{{ item.description || '无描述' }}</small>
            <n-button quaternary circle size="tiny" type="error" @click.stop="confirmRemove(item)">
              <Trash2 :size="14" />
            </n-button>
          </button>
        </div>
        <div v-else class="empty-taxonomy">暂无{{ title }}</div>
      </n-spin>
    </section>
  </div>
</template>

<style scoped>
.taxonomy-grid {
  display: grid;
  grid-template-columns: 360px minmax(0, 1fr);
  gap: 16px;
}

.taxonomy-form,
.taxonomy-list {
  padding: 16px;
}

.form-title {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.form-title h2 {
  margin: 0;
  font-size: 18px;
}

.form-title p {
  margin: 4px 0 0;
  color: var(--color-text-sec);
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.taxonomy-chip {
  display: grid;
  grid-template-columns: minmax(70px, auto) minmax(0, 1fr) 26px;
  align-items: center;
  gap: 8px;
  max-width: 360px;
  padding: 8px 8px 8px 12px;
  color: var(--color-text-main);
  background: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: 999px;
  cursor: pointer;
}

.taxonomy-chip.active {
  color: var(--color-primary);
  border-color: var(--color-primary);
  background: var(--color-primary-suppl);
}

.taxonomy-chip span {
  font-weight: 700;
}

.taxonomy-chip small {
  overflow: hidden;
  color: var(--color-text-sec);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-taxonomy {
  display: grid;
  min-height: 180px;
  place-items: center;
  color: var(--color-text-sec);
}

@media (max-width: 860px) {
  .taxonomy-grid {
    grid-template-columns: 1fr;
  }
}
</style>
