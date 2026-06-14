ALTER TABLE bookshelves
  ADD COLUMN IF NOT EXISTS personal_cover_path TEXT;
