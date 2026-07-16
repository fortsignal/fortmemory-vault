#!/usr/bin/env bash
# Placeholder demo once serve + write are implemented.
set -euo pipefail
BASE="${FORTMEMORY_URL:-http://127.0.0.1:7432}"
TOKEN="${FORTMEMORY_TOKEN:?set FORTMEMORY_TOKEN}"

curl -sS "$BASE/v1/health" | jq .
echo "write demo not wired yet — implement memory.Write first"
