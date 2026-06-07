<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{
  title: string
}>()

const emit = defineEmits<{
  click: []
}>()

const titleRef = ref<HTMLElement | null>(null)
const displayTitle = ref(props.title)
const isClamped = ref(false)
let resizeObserver: ResizeObserver | null = null
let measureRun = 0

function titleFits(element: HTMLElement) {
  return element.scrollHeight <= element.clientHeight + 1
}

async function updateTitleClamp() {
  const element = titleRef.value
  if (!element) return
  const run = ++measureRun
  const title = props.title

  displayTitle.value = title
  isClamped.value = false
  await nextTick()
  if (run !== measureRun || !titleRef.value) return
  if (titleFits(titleRef.value)) return

  let low = 0
  let high = title.length
  let best = '...'

  while (low <= high) {
    const mid = Math.floor((low + high) / 2)
    displayTitle.value = `${title.slice(0, mid).trimEnd()}...`
    await nextTick()
    if (run !== measureRun || !titleRef.value) return

    if (titleFits(titleRef.value)) {
      best = displayTitle.value
      low = mid + 1
    } else {
      high = mid - 1
    }
  }

  displayTitle.value = best
  isClamped.value = true
}

watch(
  () => props.title,
  () => nextTick(updateTitleClamp)
)

onMounted(() => {
  nextTick(() => {
    updateTitleClamp()
    if (titleRef.value && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(updateTitleClamp)
      resizeObserver.observe(titleRef.value)
    }
  })
  window.addEventListener('resize', updateTitleClamp)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateTitleClamp)
})
</script>

<template>
  <button ref="titleRef" class="book-title-control" :class="{ 'is-clamped': isClamped }" type="button" :title="title" @click="emit('click')">
    {{ displayTitle }}
  </button>
</template>

<style scoped>
.book-title-control {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  height: var(--book-title-height, 42px);
  padding: 0;
  overflow: hidden;
  color: inherit;
  font-size: var(--book-title-font-size, 15px);
  font-weight: var(--book-title-font-weight, 700);
  line-height: var(--book-title-line-height, 21px);
  overflow-wrap: anywhere;
  word-break: break-word;
  white-space: normal;
  text-align: left;
  letter-spacing: 0;
  cursor: pointer;
  background: transparent;
  border: 0;
}
</style>
