#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f -name '*.md' \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
  echo "No Markdown files found."
  exit 0
fi

npx --no-install markdownlint-cli2 "**/*.md" "#node_modules" "#dist" "#build"
