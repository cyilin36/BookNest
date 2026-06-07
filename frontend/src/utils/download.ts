import type { AxiosResponse } from 'axios'
import type { BookFormat } from '@/api/types'

const ILLEGAL_FILENAME_CHARS = /[<>:"/\\|?*\x00-\x1F]/g

function decodeFilename(value: string) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

export function filenameFromDisposition(disposition: string | undefined) {
  if (!disposition) return null
  const encodedMatch = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (encodedMatch?.[1]) return decodeFilename(encodedMatch[1].trim().replace(/^"|"$/g, ''))

  const plainMatch = disposition.match(/filename="?([^";]+)"?/i)
  return plainMatch?.[1]?.trim() || null
}

export function createBookDownloadName(title: string, format: BookFormat) {
  const safeTitle = title.replace(ILLEGAL_FILENAME_CHARS, '_').trim() || 'book'
  const extension = `.${format}`
  return safeTitle.toLowerCase().endsWith(extension) ? safeTitle : `${safeTitle}${extension}`
}

export function downloadBlob(response: AxiosResponse<Blob>, fallbackName: string) {
  const headerName = response.headers['content-disposition']
  const filename = filenameFromDisposition(Array.isArray(headerName) ? headerName[0] : headerName) || fallbackName
  const url = URL.createObjectURL(response.data)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 0)
}
