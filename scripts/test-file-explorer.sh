#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "${1:-}" == "--with-coverage" ]]; then
	npm --prefix "$SCRIPT_DIR/../apps/file-explorer" run test:unit:coverage
else
	npm --prefix "$SCRIPT_DIR/../apps/file-explorer" run test:unit
fi
