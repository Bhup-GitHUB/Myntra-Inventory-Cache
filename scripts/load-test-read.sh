#!/usr/bin/env sh
set -eu

URL="${URL:-http://localhost:8080/inventory/101}"

if ! command -v hey >/dev/null 2>&1; then
  echo "hey is required for this script. Install it with: go install github.com/rakyll/hey@latest"
  exit 1
fi

hey -n "${N:-10000}" -c "${C:-100}" "$URL"
