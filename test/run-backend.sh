#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="$ROOT_DIR/test/backend-data"

mkdir -p "$DATA_DIR/books" "$DATA_DIR/covers" "$DATA_DIR/temp"

cd "$ROOT_DIR/backend"

export DATABASE_DSN="postgres://book_reader:book_reader_password@localhost:15432/book_reader?sslmode=disable"
export DATA_DIR="$DATA_DIR"
export BOOKS_DIR="$DATA_DIR/books"
export COVERS_DIR="$DATA_DIR/covers"
export TEMP_DIR="$DATA_DIR/temp"
export HTTP_ADDR=":8080"
export JWT_SECRET="test-only-change-me"

exec go run ./cmd/server
