ALTER TABLE bookshelves
  ADD COLUMN IF NOT EXISTS personal_author VARCHAR(512),
  ADD COLUMN IF NOT EXISTS personal_description TEXT;

UPDATE bookshelves bs
SET
  personal_title = COALESCE(bs.personal_title, b.title),
  personal_author = COALESCE(bs.personal_author, b.author),
  personal_description = COALESCE(bs.personal_description, b.description)
FROM books b
WHERE bs.book_id = b.id
  AND bs.source_type = 'library';
