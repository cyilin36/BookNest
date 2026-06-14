<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { CloudUpload, Library, LockKeyhole } from 'lucide-vue-next'
import CoverUploadField from '@/components/book/CoverUploadField.vue'
import PageShell from '@/components/common/PageShell.vue'
import { useUpload } from '@/composables/useUpload'
import { useTaxonomyOptions } from '@/composables/useTaxonomyOptions'
import { useSystemStore } from '@/stores/system'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const system = useSystemStore()
const { uploading, uploadPrivate, uploadPublic } = useUpload()
const { categoryOptions, tagOptions, loading: taxonomyLoading } = useTaxonomyOptions()
const target = ref(route.query.target === 'public' ? 'public' : 'private')
const file = ref<File | null>(null)
const cover = ref<File | null>(null)
const form = reactive({ title: '', author: '', description: '', category_ids: [] as number[], tag_ids: [] as number[] })
const dragActive = ref(false)

const accept = computed(() => system.supportedFormats.map((format) => `.${format}`).join(','))

function handleFiles(files: FileList | null) {
  file.value = files?.[0] || null
}

async function submit() {
  if (!file.value) {
    message.warning('请选择图书文件')
    return
  }
  try {
    const payload = {
      file: file.value,
      cover: cover.value,
      title: form.title || undefined,
      author: form.author || undefined,
      description: form.description || undefined,
      category_ids: form.category_ids,
      tag_ids: form.tag_ids
    }
    const result = target.value === 'public' ? await uploadPublic(payload) : await uploadPrivate(payload)
    message.success(target.value === 'public' ? '公共图书已上传' : '已上传到书架')
    router.push(target.value === 'public' ? `/library/${result.id}` : `/bookshelf/${result.id}`)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '上传失败')
  }
}
</script>

<template>
  <PageShell title="上传图书" subtitle="支持 EPUB、PDF、TXT，解析和章节生成由后端完成">
    <div class="upload-grid">
      <button class="target-card surface" :class="{ active: target === 'private' }" type="button" @click="target = 'private'">
        <LockKeyhole :size="28" />
        <strong>个人私有书架</strong>
        <span>仅自己可见，占用个人存储配额。</span>
      </button>
      <button class="target-card surface" :class="{ active: target === 'public' }" type="button" @click="target = 'public'">
        <Library :size="28" />
        <strong>全站图书馆</strong>
        <span>其他用户可引用加入，不复制物理文件。</span>
      </button>
    </div>
    <div
      class="drop-zone surface"
      :class="{ active: dragActive }"
      @dragover.prevent="dragActive = true"
      @dragleave.prevent="dragActive = false"
      @drop.prevent="dragActive = false; handleFiles($event.dataTransfer?.files || null)"
    >
      <CloudUpload :size="42" />
      <strong>{{ file?.name || '拖拽文件到这里，或点击选择' }}</strong>
      <span>最大 {{ system.maxUploadSizeMb }} MB · {{ system.supportedFormats.map((f) => f.toUpperCase()).join(' / ') }}</span>
      <input type="file" :accept="accept" @change="handleFiles(($event.target as HTMLInputElement).files)" />
    </div>
    <n-form class="surface upload-form" label-placement="top" @submit.prevent="submit">
      <n-form-item label="标题">
        <n-input v-model:value="form.title" placeholder="留空时使用文件名或解析出的标题" />
      </n-form-item>
      <n-form-item label="作者">
        <n-input v-model:value="form.author" placeholder="可选" />
      </n-form-item>
      <n-form-item label="简介">
        <n-input v-model:value="form.description" type="textarea" placeholder="可选" />
      </n-form-item>
      <n-form-item label="封面">
        <CoverUploadField v-model:file="cover" :disabled="uploading" />
      </n-form-item>
      <n-form-item label="分类">
        <n-select v-model:value="form.category_ids" multiple clearable :loading="taxonomyLoading" :options="categoryOptions" placeholder="可选" />
      </n-form-item>
      <n-form-item label="标签">
        <n-select v-model:value="form.tag_ids" multiple clearable :loading="taxonomyLoading" :options="tagOptions" placeholder="可选" />
      </n-form-item>
      <n-button type="primary" attr-type="submit" :loading="uploading">上传</n-button>
    </n-form>
  </PageShell>
</template>

<style scoped>
.upload-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.target-card {
  display: grid;
  gap: 8px;
  padding: 18px;
  color: var(--color-text-main);
  text-align: left;
  cursor: pointer;
}

.target-card span {
  color: var(--color-text-sec);
}

.target-card.active {
  border-color: var(--color-primary);
  box-shadow: inset 0 0 0 1px var(--color-primary);
}

.drop-zone {
  position: relative;
  display: grid;
  min-height: 190px;
  place-items: center;
  align-content: center;
  gap: 8px;
  border: 2px dashed var(--color-border);
  color: var(--color-text-sec);
}

.drop-zone.active {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.drop-zone input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.upload-form {
  padding: 16px;
}

@media (max-width: 720px) {
  .upload-grid {
    grid-template-columns: 1fr;
  }
}
</style>
