export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'disabled'
export type BookFormat = 'epub' | 'pdf' | 'txt'
export type BookVisibility = 'private' | 'public'
export type LibraryStatus = 'approved' | 'hidden' | 'deleted'
export type BookshelfSourceType = 'uploaded' | 'library'
export type BookshelfStatus = 'active' | 'removed' | 'unavailable'
export type ProgressType = 'epub_cfi' | 'pdf_page' | 'txt_offset'
export type ParseStatus = 'parsed' | 'partial' | 'failed'
export type ThemeName = 'modern' | 'sepia' | 'dark'
export type ReaderLineHeight = 1.5 | 1.8 | 2.2
export type ReaderFontFamily = 'sans' | 'serif'
export type ReaderMode = 'scroll' | 'page'

export interface APIErrorBody {
  code: string
  message: string
}

export interface APIErrorPayload {
  error: APIErrorBody
  request_id: string
}

export interface SuccessPayload<T> {
  data: T
  request_id: string
}

export interface Pagination {
  page: number
  page_size: number
  total: number
}

export interface PaginatedPayload<T> {
  data: T[]
  pagination: Pagination
  request_id: string
}

export interface PageQuery {
  page?: number
  page_size?: number
  keyword?: string
  sort?: string
  order?: 'asc' | 'desc'
}

export interface User {
  id: number
  username: string
  email: string | null
  nickname: string | null
  avatar_url: string | null
  role: UserRole
  status: UserStatus
  storage_quota_bytes: number | null
  storage_used_bytes: number
  effective_storage_quota_bytes: number | null
  created_at?: string
  last_login_at?: string | null
}

export interface AuthSession {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
  user: User
}

export interface SystemInfo {
  site_name: string
  site_icon_url: string | null
  login_background_url: string | null
  allow_registration: boolean
  library_review_required: boolean
  supported_formats: BookFormat[]
  max_upload_size_mb: number
  default_user_storage_quota_mb: number
}

export interface SystemSettings {
  site_name: string
  site_icon_url: string | null
  login_background_url: string | null
  allow_registration: boolean
  library_review_required: boolean
  max_upload_size_mb: number
  default_user_storage_quota_mb: number
}

export type UpdateSystemSettingsRequest = Omit<SystemSettings, 'site_icon_url' | 'login_background_url'>

export interface SystemIconUploadResult {
  site_icon_url: string
  content_type: 'image/png' | 'image/jpeg' | 'image/webp' | 'image/svg+xml' | 'image/x-icon'
}

export interface LoginBackgroundUploadResult {
  login_background_url: string
  content_type: string
}

export interface BookMeta {
  id: number
  title: string
  author: string | null
  description: string | null
  format: BookFormat
  cover_url: string | null
  file_size: number
  visibility: BookVisibility
  library_status: LibraryStatus | null
  parse_status?: ParseStatus
  created_at?: string
  updated_at?: string
}

export interface BookshelfItem {
  id: number
  book_id: number
  title: string
  author: string | null
  description: string | null
  format: BookFormat
  cover_url: string | null
  source_type: BookshelfSourceType
  visibility: BookVisibility
  library_status: LibraryStatus | null
  favorite: boolean
  pinned: boolean
  personal_category_id: number | null
  tag_ids: number[]
  last_read_at: string | null
  added_at: string
  readable: boolean
  unreadable_reason: 'library_hidden' | 'library_deleted' | 'file_missing' | 'permission_denied' | null
  progress_percentage: number | null
}

export interface LibraryBook {
  id: number
  title: string
  author: string | null
  description: string | null
  format: BookFormat
  cover_url: string | null
  file_size: number
  library_status: LibraryStatus
  owner_user_id: number
  owner_username?: string | null
  category_ids: number[]
  tag_ids: number[]
  in_bookshelf: boolean
  bookshelf_id: number | null
  created_at: string
  updated_at?: string
}

export interface Category {
  id: number
  name: string
  description: string | null
  scope: 'system'
  created_at?: string
}

export interface Tag {
  id: number
  name: string
  description: string | null
  scope: 'system'
  created_at?: string
}

export interface ReaderChapter {
  id: number
  chapter_index: number
  title: string
  is_volume: boolean
  word_count: number | null
}

export interface ReaderChapterContent {
  id: number
  chapter_index: number
  title: string
  content_type: 'html' | 'text'
  content: string
}

export interface ReadingProgress {
  progress_type: ProgressType
  progress_value: string
  percentage: number | null
  updated_at: string
}

export interface ReaderSettings {
  theme: ThemeName
  reading_mode: ReaderMode
  font_size: number
  line_height: ReaderLineHeight
  content_width: number
  font_family: ReaderFontFamily
}

export interface UploadBookRequest {
  file: File
  cover?: File | null
  title?: string
  author?: string
  description?: string
  category_ids?: number[]
  tag_ids?: number[]
}

export interface AdminStorageStats {
  private_books_bytes: number
  public_books_bytes: number
  covers_bytes: number
  total_books: number
  private_books: number
  public_books: number
}
