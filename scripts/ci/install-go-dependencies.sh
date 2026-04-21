#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*')

if [ "${#modules[@]}" -eq 0 ]; then
  echo "No Go modules found; skipping CI Go dependency installation."
  exit 0
fi

echo "Installing CI Go analysis dependencies"
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
