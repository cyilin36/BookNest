UPDATE books
SET library_status = 'approved', updated_at = now()
WHERE library_status = 'pending';

UPDATE books
SET library_status = 'hidden', updated_at = now()
WHERE library_status = 'rejected';

ALTER TABLE books DROP CONSTRAINT IF EXISTS chk_books_library_status;

ALTER TABLE books ADD CONSTRAINT chk_books_library_status CHECK (
  library_status IS NULL OR library_status IN ('approved', 'hidden', 'deleted')
);
