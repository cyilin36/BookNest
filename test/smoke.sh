#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$ROOT_DIR/test/tmp"
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

wait_for "$BASE_URL/health"

echo "health:"
curl -fsS "$BASE_URL/health"
echo

echo "system info:"
curl -fsS "$BASE_URL/system/info"
echo

REGISTER_PAYLOAD='{"username":"admin_test","email":"admin_test@example.com","password":"test123456","nickname":"Admin Test"}'
REGISTER_RESPONSE="$(curl -sS -X POST "$BASE_URL/auth/register" -H 'Content-Type: application/json' -d "$REGISTER_PAYLOAD")"

if echo "$REGISTER_RESPONSE" | grep -Eq '"(username_exists|email_exists)"'; then
  LOGIN_PAYLOAD='{"login":"admin_test","password":"test123456"}'
  SESSION="$(curl -fsS -X POST "$BASE_URL/auth/login" -H 'Content-Type: application/json' -d "$LOGIN_PAYLOAD")"
else
  SESSION="$REGISTER_RESPONSE"
fi

ACCESS_TOKEN="$(printf '%s' "$SESSION" | json_field access_token)"

echo "auth me:"
curl -fsS "$BASE_URL/auth/me" -H "Authorization: Bearer $ACCESS_TOKEN"
echo

TMP_BOOK="$(mktemp "$TMP_DIR/book-reader-smoke.XXXXXX")"
trap 'rm -f "$TMP_BOOK"' EXIT
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

echo "smoke ok"
