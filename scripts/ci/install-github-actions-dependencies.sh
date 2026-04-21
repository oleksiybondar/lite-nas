#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

log.pushTask "Installing CI GitHub Actions analysis dependencies"
go install github.com/rhysd/actionlint/cmd/actionlint@latest
log.popTask
