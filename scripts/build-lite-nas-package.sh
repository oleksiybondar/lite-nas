#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/build-all.sh"
"$SCRIPT_DIR/package/build-lite-nas-deb-from-source.sh" \
	--source-layout=build-all \
	--source-root="$SCRIPT_DIR/../.build/all" \
	"$@"
