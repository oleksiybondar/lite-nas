#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f -name '*.go' -not -path './vendor/*')

if [ "${#files[@]}" -eq 0 ]; then
  echo "No Go files found."
  exit 0
fi

gofumpt -w "${files[@]}"
goimports -w "${files[@]}"
