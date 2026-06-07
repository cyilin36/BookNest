ALTER TABLE books DROP CONSTRAINT IF EXISTS chk_books_library_status;

ALTER TABLE books ADD CONSTRAINT chk_books_library_status CHECK (
  library_status IS NULL OR library_status IN ('pending', 'approved', 'rejected', 'hidden', 'deleted')
);
