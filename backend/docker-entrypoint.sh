#!/bin/sh
set -eu

PUID="${PUID:-1000}"
PGID="${PGID:-1000}"
DATA_DIR="${DATA_DIR:-/data}"
BOOKS_DIR="${BOOKS_DIR:-$DATA_DIR/books}"
COVERS_DIR="${COVERS_DIR:-$DATA_DIR/covers}"
ASSETS_DIR="${ASSETS_DIR:-$DATA_DIR/assets}"
TEMP_DIR="${TEMP_DIR:-$DATA_DIR/temp}"

case "$PUID" in
  ''|*[!0-9]*) echo "PUID must be a numeric user id" >&2; exit 1 ;;
esac

case "$PGID" in
  ''|*[!0-9]*) echo "PGID must be a numeric group id" >&2; exit 1 ;;
esac

if [ "$PUID" = "0" ] || [ "$PGID" = "0" ]; then
  echo "PUID and PGID must be non-root ids" >&2
  exit 1
fi

if [ "$(id -u)" = "0" ]; then
  GROUP_NAME="$(awk -F: -v gid="$PGID" '$3 == gid { print $1; exit }' /etc/group)"
  if [ -z "$GROUP_NAME" ]; then
    addgroup -g "$PGID" booknest >/dev/null
    GROUP_NAME="booknest"
  fi

  USER_NAME="$(awk -F: -v uid="$PUID" '$3 == uid { print $1; exit }' /etc/passwd)"
  if [ -z "$USER_NAME" ]; then
    adduser -D -H -u "$PUID" -G "$GROUP_NAME" booknest >/dev/null
  fi

  mkdir -p "$DATA_DIR" "$BOOKS_DIR" "$COVERS_DIR" "$ASSETS_DIR" "$TEMP_DIR"
  chown -R "$PUID:$PGID" "$DATA_DIR" "$BOOKS_DIR" "$COVERS_DIR" "$ASSETS_DIR" "$TEMP_DIR"

  exec su-exec "$PUID:$PGID" "$@"
fi

exec "$@"
