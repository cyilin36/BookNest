UPDATE bookshelves
SET added_at = now()
WHERE added_at < TIMESTAMPTZ '1970-01-01';
