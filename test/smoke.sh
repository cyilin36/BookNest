#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$ROOT_DIR/test/tmp"
SMOKE_USERNAME="${SMOKE_USERNAME:-admin_test}"
SMOKE_EMAIL="${SMOKE_EMAIL:-admin_test@example.com}"
SMOKE_PASSWORD="${SMOKE_PASSWORD:-test123456}"
SMOKE_NICKNAME="${SMOKE_NICKNAME:-Admin Test}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-book-reader-test-postgres}"
POSTGRES_USER="${POSTGRES_USER:-book_reader}"
POSTGRES_DB="${POSTGRES_DB:-book_reader}"
TMP_BOOK=""
mkdir -p "$TMP_DIR"

wait_for() {
  local url="$1"
  for _ in $(seq 1 60); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "Timed out waiting for $url" >&2
  return 1
}

json_field() {
  python3 -c 'import json,sys; data=json.load(sys.stdin); print(data["data"][sys.argv[1]])' "$1"
}

json_expr() {
  python3 -c 'import json,sys; data=json.load(sys.stdin); print(eval(sys.argv[1], {}, {"data": data}))' "$1"
}

cleanup_smoke_user() {
  if ! docker ps --format '{{.Names}}' | grep -qx "$POSTGRES_CONTAINER"; then
    return 0
  fi
  while IFS= read -r file_path; do
    if [ -n "$file_path" ] && [[ "$file_path" != /* ]] && [[ "$file_path" != *..* ]]; then
      rm -f "$ROOT_DIR/test/backend-data/books/$file_path"
    fi
  done < <(
    docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At \
      -v smoke_username="$SMOKE_USERNAME" <<'SQL'
SELECT file_path FROM books WHERE owner_user_id IN (SELECT id FROM users WHERE username = :'smoke_username');
SQL
  )
  docker exec -i "$POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    -v smoke_username="$SMOKE_USERNAME" <<'SQL'
BEGIN;
CREATE TEMP TABLE smoke_user_ids ON COMMIT DROP AS
  SELECT id FROM users WHERE username = :'smoke_username';
CREATE TEMP TABLE smoke_book_ids ON COMMIT DROP AS
  SELECT id FROM books WHERE owner_user_id IN (SELECT id FROM smoke_user_ids);
CREATE TEMP TABLE smoke_bookshelf_ids ON COMMIT DROP AS
  SELECT id FROM bookshelves
  WHERE user_id IN (SELECT id FROM smoke_user_ids)
     OR book_id IN (SELECT id FROM smoke_book_ids);

DELETE FROM bookshelf_tags WHERE bookshelf_id IN (SELECT id FROM smoke_bookshelf_ids);
DELETE FROM bookmarks
WHERE user_id IN (SELECT id FROM smoke_user_ids)
   OR bookshelf_id IN (SELECT id FROM smoke_bookshelf_ids)
   OR book_id IN (SELECT id FROM smoke_book_ids);
DELETE FROM reading_progress
WHERE user_id IN (SELECT id FROM smoke_user_ids)
   OR bookshelf_id IN (SELECT id FROM smoke_bookshelf_ids)
   OR book_id IN (SELECT id FROM smoke_book_ids);
DELETE FROM bookshelves WHERE id IN (SELECT id FROM smoke_bookshelf_ids);
DELETE FROM book_categories WHERE book_id IN (SELECT id FROM smoke_book_ids);
DELETE FROM book_tags WHERE book_id IN (SELECT id FROM smoke_book_ids);
DELETE FROM book_chapters WHERE book_id IN (SELECT id FROM smoke_book_ids);
DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM smoke_user_ids);
DELETE FROM books WHERE id IN (SELECT id FROM smoke_book_ids);
DELETE FROM users WHERE id IN (SELECT id FROM smoke_user_ids);
COMMIT;
SQL
}

cleanup_on_exit() {
  if [ -n "$TMP_BOOK" ]; then
    rm -f "$TMP_BOOK"
  fi
  cleanup_smoke_user
}

trap cleanup_on_exit EXIT

cleanup_smoke_user

wait_for "$BASE_URL/health"

echo "health:"
curl -fsS "$BASE_URL/health"
echo

echo "system info:"
curl -fsS "$BASE_URL/system/info"
echo

REGISTER_PAYLOAD="$(python3 -c 'import json,sys; print(json.dumps({"username": sys.argv[1], "email": sys.argv[2], "password": sys.argv[3], "nickname": sys.argv[4]}))' "$SMOKE_USERNAME" "$SMOKE_EMAIL" "$SMOKE_PASSWORD" "$SMOKE_NICKNAME")"
REGISTER_RESPONSE="$(curl -sS -X POST "$BASE_URL/auth/register" -H 'Content-Type: application/json' -d "$REGISTER_PAYLOAD")"

if echo "$REGISTER_RESPONSE" | grep -Eq '"(username_exists|email_exists)"'; then
  LOGIN_PAYLOAD="$(python3 -c 'import json,sys; print(json.dumps({"login": sys.argv[1], "password": sys.argv[2]}))' "$SMOKE_USERNAME" "$SMOKE_PASSWORD")"
  SESSION="$(curl -fsS -X POST "$BASE_URL/auth/login" -H 'Content-Type: application/json' -d "$LOGIN_PAYLOAD")"
else
  SESSION="$REGISTER_RESPONSE"
fi

ACCESS_TOKEN="$(printf '%s' "$SESSION" | json_field access_token)"

echo "auth me:"
curl -fsS "$BASE_URL/auth/me" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

TMP_BOOK="$(mktemp "$TMP_DIR/book-reader-smoke.XXXXXX")"
cat > "$TMP_BOOK" <<'BOOK'
第一章 初见
这是第一章的正文。

第二章 继续
这是第二章的正文。
BOOK

echo "upload private txt:"
UPLOAD_RESPONSE="$(curl -fsS -X POST "$BASE_URL/bookshelf/upload" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "file=@$TMP_BOOK;filename=smoke.txt" \
  -F "title=Smoke TXT")"
printf '%s\n' "$UPLOAD_RESPONSE"

BOOK_ID="$(printf '%s' "$UPLOAD_RESPONSE" | json_expr 'data["data"]["book_id"]')"
BOOKSHELF_ID="$(printf '%s' "$UPLOAD_RESPONSE" | json_expr 'data["data"]["id"]')"

echo "bookshelf:"
curl -fsS "$BASE_URL/bookshelf" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "reader meta:"
curl -fsS "$BASE_URL/reader/books/$BOOK_ID/meta" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "reader chapters:"
CHAPTERS_RESPONSE="$(curl -fsS "$BASE_URL/reader/books/$BOOK_ID/chapters" -H "Authorization: Bearer $ACCESS_TOKEN")"
printf '%s\n' "$CHAPTERS_RESPONSE"
CHAPTER_ID="$(printf '%s' "$CHAPTERS_RESPONSE" | json_expr 'data["data"][0]["id"]')"

echo "reader chapter content:"
curl -fsS "$BASE_URL/reader/books/$BOOK_ID/chapters/$CHAPTER_ID/content" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "save progress:"
curl -fsS -X PUT "$BASE_URL/reader/books/$BOOK_ID/progress" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"progress_type":"txt_offset","progress_value":"0","percentage":12.5}'
echo

echo "reader progress:"
curl -fsS "$BASE_URL/reader/books/$BOOK_ID/progress" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "bookshelf after progress:"
curl -fsS "$BASE_URL/bookshelf" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "cleanup bookshelf item:"
curl -fsS -X DELETE "$BASE_URL/bookshelf/$BOOKSHELF_ID" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

echo "cleanup smoke user:"
cleanup_smoke_user

echo "smoke ok"
