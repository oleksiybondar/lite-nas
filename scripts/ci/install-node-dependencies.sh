#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

log.pushTask "Installing CI Node dependencies"
if [ -f package-lock.json ]; then
	npm ci
else
	npm install
fi
log.popTask
