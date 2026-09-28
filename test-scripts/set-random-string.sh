#!/usr/bin/env bash
# Sets common.RandomString (appended to every backend /hello response).
# Usage: set-random-string.sh [value]   (default: a new random 8-char string)
# Pass --rebuild as the last argument to also rebuild the example namespace.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
file="${repo_root}/example/projects/common/random.go"

rebuild=false
if [ "${!#:-}" = "--rebuild" ]; then
  rebuild=true
  set -- "${@:1:$(($# - 1))}"
fi

value="${1:-$(openssl rand -hex 4)}"
if ! [[ "$value" =~ ^[A-Za-z0-9._-]+$ ]]; then
  echo "value may only contain letters, digits, '.', '_' and '-'" >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
sed -E "s/^const RandomString = \".*\"$/const RandomString = \"${value}\"/" "$file" > "$tmp"
if ! grep -qx "const RandomString = \"${value}\"" "$tmp"; then
  echo "could not find RandomString constant in $file" >&2
  exit 1
fi
cat "$tmp" > "$file"
echo "RandomString set to ${value}"

if [ "$rebuild" = true ]; then
  "${repo_root}/test-scripts/rebuild.sh" example
fi
