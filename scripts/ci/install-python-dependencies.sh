#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

log.pushTask "Setting up Python virtual environment"
if [ ! -f .venv/bin/pip ]; then
	rm -rf .venv
	python3 -m venv .venv
fi
log.popTask

log.pushTask "Installing CI Python dependencies"
.venv/bin/pip install --quiet -r tests/requirements.txt
log.popTask

log.pushTask "Installing Playwright runtime dependencies"
if command -v sudo >/dev/null 2>&1; then
	sudo .venv/bin/playwright install-deps
else
	.venv/bin/playwright install-deps
fi
.venv/bin/playwright install chromium
log.popTask
