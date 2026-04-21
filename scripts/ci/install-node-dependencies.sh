#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo "Installing CI Node dependencies"
if [ -f package-lock.json ]; then
  npm ci
else
  npm install
fi
