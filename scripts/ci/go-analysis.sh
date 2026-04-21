#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*')

if [ "${#modules[@]}" -eq 0 ]; then
  echo "No Go modules found."
  exit 0
fi

for module in "${modules[@]}"; do
  dir="$(dirname "$module")"
  echo "Running golangci-lint in ${dir}"
  (cd "$dir" && golangci-lint run)
done
