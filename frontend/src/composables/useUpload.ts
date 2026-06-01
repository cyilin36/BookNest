import { ref } from 'vue'
import { bookshelfApi } from '@/api/bookshelf'
import { libraryApi } from '@/api/library'
import type { UploadBookRequest } from '@/api/types'

function toFormData(payload: UploadBookRequest) {
  const form = new FormData()
  form.append('file', payload.file)
  if (payload.title) form.append('title', payload.title)
  if (payload.author) form.append('author', payload.author)
  if (payload.description) form.append('description', payload.description)
  if (payload.category_ids?.length) form.append('category_ids', payload.category_ids.join(','))
  if (payload.tag_ids?.length) form.append('tag_ids', payload.tag_ids.join(','))
  return form
}

export function useUpload() {
  const uploading = ref(false)

  async function uploadPrivate(payload: UploadBookRequest) {
    uploading.value = true
    try {
      return await bookshelfApi.upload(toFormData(payload))
    } finally {
      uploading.value = false
    }
  }

  async function uploadPublic(payload: UploadBookRequest) {
    uploading.value = true
    try {
      return await libraryApi.upload(toFormData(payload))
    } finally {
      uploading.value = false
    }
  }

  return { uploading, uploadPrivate, uploadPublic }
}
