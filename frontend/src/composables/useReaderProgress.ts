import { useThrottleFn } from '@vueuse/core'
import { useReaderStore } from '@/stores/reader'
import type { ProgressType } from '@/api/types'

export function useReaderProgress(bookId: number) {
  const reader = useReaderStore()

  const saveNow = (progress_type: ProgressType, progress_value: string, percentage: number | null) => reader.saveProgress(bookId, { progress_type, progress_value, percentage }).catch(() => undefined)
  const save = useThrottleFn(saveNow, 2500)

  return { save, saveNow }
}
