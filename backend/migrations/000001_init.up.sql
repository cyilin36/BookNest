CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  email VARCHAR(255),
  password_hash TEXT NOT NULL,
  nickname VARCHAR(128),
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  storage_quota_bytes BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ,
  CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user')),
  CONSTRAINT chk_users_status CHECK (status IN ('active', 'disabled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique_not_null ON users(email) WHERE email IS NOT NULL;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  user_agent TEXT,
  ip_address VARCHAR(64),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT,
  scope VARCHAR(32) NOT NULL DEFAULT 'system',
  owner_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_categories_scope CHECK (scope IN ('system', 'user'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_system_slug ON categories(slug) WHERE scope = 'system';

CREATE TABLE IF NOT EXISTS tags (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT,
  scope VARCHAR(32) NOT NULL DEFAULT 'system',
  owner_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_tags_scope CHECK (scope IN ('system', 'user'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_system_slug ON tags(slug) WHERE scope = 'system';

CREATE TABLE IF NOT EXISTS books (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(512) NOT NULL,
  author VARCHAR(512),
  description TEXT,
  original_filename VARCHAR(512),
  format VARCHAR(32) NOT NULL,
  file_size BIGINT NOT NULL,
  file_hash VARCHAR(128) NOT NULL,
  file_path TEXT NOT NULL,
  cover_path TEXT,
  charset VARCHAR(64),
  visibility VARCHAR(32) NOT NULL,
  owner_user_id BIGINT NOT NULL REFERENCES users(id),
  library_status VARCHAR(32),
  parse_status VARCHAR(32) NOT NULL DEFAULT 'parsed',
  parse_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT chk_books_format CHECK (format IN ('epub', 'pdf', 'txt')),
  CONSTRAINT chk_books_visibility CHECK (visibility IN ('private', 'public')),
  CONSTRAINT chk_books_library_status CHECK (
    library_status IS NULL OR library_status IN ('approved', 'hidden', 'deleted')
  ),
  CONSTRAINT chk_books_parse_status CHECK (parse_status IN ('parsed', 'partial', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_books_visibility ON books(visibility);
CREATE INDEX IF NOT EXISTS idx_books_owner_user_id ON books(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_books_file_hash ON books(file_hash);
CREATE INDEX IF NOT EXISTS idx_books_library_status ON books(library_status);
CREATE INDEX IF NOT EXISTS idx_books_deleted_at ON books(deleted_at);

CREATE TABLE IF NOT EXISTS bookshelves (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  source_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  favorite BOOLEAN NOT NULL DEFAULT false,
  pinned BOOLEAN NOT NULL DEFAULT false,
  personal_title VARCHAR(512),
  personal_author VARCHAR(512),
  personal_description TEXT,
  personal_cover_path TEXT,
  personal_category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
  added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_read_at TIMESTAMPTZ,
  removed_at TIMESTAMPTZ,
  UNIQUE(user_id, book_id),
  CONSTRAINT chk_bookshelves_source_type CHECK (source_type IN ('uploaded', 'library')),
  CONSTRAINT chk_bookshelves_status CHECK (status IN ('active', 'removed', 'unavailable'))
);

CREATE INDEX IF NOT EXISTS idx_bookshelves_user_id ON bookshelves(user_id);
CREATE INDEX IF NOT EXISTS idx_bookshelves_book_id ON bookshelves(book_id);
CREATE INDEX IF NOT EXISTS idx_bookshelves_last_read_at ON bookshelves(last_read_at);

CREATE TABLE IF NOT EXISTS book_categories (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, category_id)
);

CREATE TABLE IF NOT EXISTS book_tags (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(book_id, tag_id)
);

CREATE TABLE IF NOT EXISTS bookshelf_tags (
  bookshelf_id BIGINT NOT NULL REFERENCES bookshelves(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY(bookshelf_id, tag_id)
);

CREATE TABLE IF NOT EXISTS book_chapters (
  id BIGSERIAL PRIMARY KEY,
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  chapter_index INTEGER NOT NULL,
  title VARCHAR(512) NOT NULL,
  locator TEXT NOT NULL,
  start_offset BIGINT,
  end_offset BIGINT,
  href TEXT,
  is_volume BOOLEAN NOT NULL DEFAULT false,
  word_count BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(book_id, chapter_index)
);

CREATE INDEX IF NOT EXISTS idx_book_chapters_book_id ON book_chapters(book_id);

CREATE TABLE IF NOT EXISTS reading_progress (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE SET NULL,
  format VARCHAR(32) NOT NULL,
  progress_type VARCHAR(32) NOT NULL,
  progress_value TEXT NOT NULL,
  percentage NUMERIC(6, 3),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, book_id),
  CONSTRAINT chk_reading_progress_type CHECK (progress_type IN ('epub_cfi', 'pdf_page', 'txt_offset')),
  CONSTRAINT chk_reading_progress_percentage CHECK (percentage IS NULL OR (percentage >= 0 AND percentage <= 100))
);

CREATE INDEX IF NOT EXISTS idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_book_id ON reading_progress(book_id);

CREATE TABLE IF NOT EXISTS bookmarks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id),
  bookshelf_id BIGINT REFERENCES bookshelves(id) ON DELETE SET NULL,
  title VARCHAR(255),
  position_type VARCHAR(32) NOT NULL,
  position_value TEXT NOT NULL,
  percentage NUMERIC(6, 3),
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_bookmarks_position_type CHECK (position_type IN ('epub_cfi', 'pdf_page', 'txt_offset'))
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_book ON bookmarks(user_id, book_id);

CREATE TABLE IF NOT EXISTS system_settings (
  key VARCHAR(128) PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO system_settings(key, value) VALUES
  ('allow_registration', 'true'),
  ('library_review_required', 'false'),
  ('site_name', 'BookNest'),
  ('max_upload_size_mb', '100'),
  ('default_user_storage_quota_mb', '10240')
ON CONFLICT (key) DO NOTHING;
