#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

log.pushTask "Installing CI shell analysis dependencies"
sudo apt-get update
sudo apt-get install -y shellcheck
go install mvdan.cc/sh/v3/cmd/shfmt@latest
log.popTask
