import { computed, onMounted, ref } from 'vue'
import { categoryApi } from '@/api/category'
import { tagApi } from '@/api/tag'
import type { Category, Tag } from '@/api/types'

export function useTaxonomyOptions() {
  const categories = ref<Category[]>([])
  const tags = ref<Tag[]>([])
  const loading = ref(false)

  const categoryOptions = computed(() => categories.value.map((item) => ({ label: item.name, value: item.id })))
  const tagOptions = computed(() => tags.value.map((item) => ({ label: item.name, value: item.id })))

  async function loadTaxonomies() {
    loading.value = true
    try {
      const [categoryRows, tagRows] = await Promise.all([categoryApi.list(), tagApi.list()])
      categories.value = categoryRows
      tags.value = tagRows
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadTaxonomies().catch(() => undefined)
  })

  return { categories, tags, categoryOptions, tagOptions, loading, loadTaxonomies }
}
