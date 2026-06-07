ALTER TABLE bookshelves
  DROP COLUMN IF EXISTS personal_description,
  DROP COLUMN IF EXISTS personal_author;
