#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( -name '*.sh' -o -name '*.bash' -o -name '*.zsh' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
  echo "No shell files found."
  exit 0
fi

shfmt -w "${files[@]}"
