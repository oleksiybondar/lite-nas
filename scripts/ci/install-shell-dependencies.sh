#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo "Installing CI shell analysis dependencies"
sudo apt-get update
sudo apt-get install -y shellcheck
go install mvdan.cc/sh/v3/cmd/shfmt@latest
