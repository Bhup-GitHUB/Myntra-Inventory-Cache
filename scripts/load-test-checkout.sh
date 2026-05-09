#!/usr/bin/env sh
set -eu

URL="${URL:-http://localhost:8080/checkout}"

if ! command -v hey >/dev/null 2>&1; then
  echo "hey is required for this script. Install it with: go install github.com/rakyll/hey@latest"
  exit 1
fi

hey -n "${N:-100}" -c "${C:-10}" -m POST \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"product_id":101,"quantity":1}' \
  "$URL"
