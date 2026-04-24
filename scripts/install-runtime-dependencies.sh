#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/sudo-guard.sh"

runtime_tools=(nats-server openssl)

sudo.guard.requireRoot "scripts/install-runtime-dependencies.sh"

install_apt_packages() {
	if ! command -v apt-get >/dev/null 2>&1; then
		cat <<'MSG' >&2
Missing required runtime dependencies, and apt-get is not available.

Install NATS Server and OpenSSL manually, then re-run this script.
On macOS, use Homebrew:
  brew install nats-server openssl
MSG
		exit 1
	fi

	log.pushTask "Installing Debian/Ubuntu runtime packages"
	apt-get update
	apt-get install -y nats-server openssl
	log.popTask
}

missing_runtime_tools=()
for tool in "${runtime_tools[@]}"; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		missing_runtime_tools+=("$tool")
	fi
done

if [ "${#missing_runtime_tools[@]}" -gt 0 ]; then
	log.info "Missing runtime dependencies: ${missing_runtime_tools[*]}"
	install_apt_packages
fi

log.pushTask "Checking runtime dependencies"
for tool in "${runtime_tools[@]}"; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		log.error "Missing required command after installation attempt: $tool"
		exit 1
	fi
done
log.popTask

log.info "Runtime dependencies are installed."
