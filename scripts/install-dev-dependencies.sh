#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/sudo-guard.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

sudo.guard.requireRoot "scripts/install-dev-dependencies.sh"

run_as_user() {
	if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
		local user_home
		user_home="$(getent passwd "$SUDO_USER" | cut -d: -f6)"
		sudo -H -u "$SUDO_USER" env HOME="$user_home" GOBIN="$GOBIN" PATH="$PATH" "$@"
	else
		"$@"
	fi
}

ensure_user_writable_directory() {
	local directory="$1"

	mkdir -p "$directory"

	if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
		chown "$SUDO_USER:$(id -gn "$SUDO_USER")" "$directory"
	fi
}

install_go_tool_if_missing() {
	local binary_name="$1"
	local package_spec="$2"

	if [ -x "$GOBIN/$binary_name" ]; then
		log.info "Keeping existing Go tool: $binary_name"
		return 0
	fi

	run_as_user go install "$package_spec"
}

install_apt_packages() {
	if ! command -v apt-get >/dev/null 2>&1; then
		cat <<'MSG' >&2
Missing required base tooling, and apt-get is not available.

Install Node.js, npm, Go, shellcheck, shfmt, and Debian packaging tools manually, then re-run this script.
On macOS, use Homebrew:
  brew install node go shellcheck shfmt actionlint
MSG
		exit 1
	fi

	log.pushTask "Installing Debian/Ubuntu base packages"
	apt-get update
	if ! apt-get install -y git nodejs npm golang-go shellcheck shfmt lintian debconf gcc libc6-dev libpam0g-dev; then
		log.warn "Could not install shfmt with apt-get; it will be installed with go install instead."
		apt-get install -y git nodejs npm golang-go shellcheck lintian debconf gcc libc6-dev libpam0g-dev
	fi
	log.popTask
}

has_pam_development_headers() {
	if ! command -v gcc >/dev/null 2>&1; then
		return 1
	fi

	printf '#include <security/pam_appl.h>\n' |
		gcc -x c -fsyntax-only - >/dev/null 2>&1
}

missing_base_tools=()
for tool in node npm go shellcheck shfmt gcc; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		missing_base_tools+=("$tool")
	fi
done

missing_dev_components=()

if ! has_pam_development_headers; then
	missing_dev_components+=("libpam0g-dev")
fi

if [ "${#missing_base_tools[@]}" -gt 0 ] || [ "${#missing_dev_components[@]}" -gt 0 ]; then
	log.info "Missing base tooling: ${missing_base_tools[*]}"
	if [ "${#missing_dev_components[@]}" -gt 0 ]; then
		log.info "Missing development components: ${missing_dev_components[*]}"
	fi
	install_apt_packages
fi

log.pushTask "Checking base tooling"
for tool in node npm go shellcheck shfmt gcc; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		log.error "Missing required command after installation attempt: $tool"
		exit 1
	fi
done
log.popTask

log.pushTask "Checking PAM build dependencies"
if ! has_pam_development_headers; then
	log.error "Missing PAM development headers after installation attempt: security/pam_appl.h"
	exit 1
fi
log.popTask

log.pushTask "Installing Node developer dependencies"
run_as_user npm install
log.popTask

log.pushTask "Installing Go developer tools"
ensure_user_writable_directory "$GOBIN"
install_go_tool_if_missing "gofumpt" "mvdan.cc/gofumpt@latest"
install_go_tool_if_missing "goimports" "golang.org/x/tools/cmd/goimports@latest"
install_go_tool_if_missing "golangci-lint" "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
install_go_tool_if_missing "actionlint" "github.com/rhysd/actionlint/cmd/actionlint@latest"
install_go_tool_if_missing "shfmt" "mvdan.cc/sh/v3/cmd/shfmt@latest"
log.popTask

log.pushTask "Installing Git hooks"
run_as_user npx lefthook install
log.popTask

log.info "Developer tooling is installed."
log.info "Repo-local Go tools are installed in ${GOBIN}."
log.info "Add ${GOBIN} to PATH for direct shell usage, or use the repository scripts."
