#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_DIR="${HOME}/.cache/gitea-mcp"
BINARY="$BINARY_DIR/gitea-mcp"

mkdir -p "$BINARY_DIR"

BUILD_NEEDED=false

if [ ! -f "$BINARY" ]; then
    BUILD_NEEDED=true
else
    SOURCE_HASH=$(find "$SCRIPT_DIR" -name "*.go" -newer "$BINARY" 2>/dev/null | head -1)
    if [ -n "$SOURCE_HASH" ]; then
        BUILD_NEEDED=true
    fi
fi

if [ "$BUILD_NEEDED" = true ]; then
    cd "$SCRIPT_DIR"
    go build -v -o "$BINARY" .
fi

exec "$BINARY" "$@"