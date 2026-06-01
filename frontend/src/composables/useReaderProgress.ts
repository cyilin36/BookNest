import { useThrottleFn } from '@vueuse/core'
import { useReaderStore } from '@/stores/reader'
import type { ProgressType } from '@/api/types'

export function useReaderProgress(bookId: number) {
  const reader = useReaderStore()

  const save = useThrottleFn((progress_type: ProgressType, progress_value: string, percentage: number | null) => {
    reader.saveProgress(bookId, { progress_type, progress_value, percentage }).catch(() => undefined)
  }, 2500)

  return { save }
}
