#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '\n==> %s\n' "$1"
}

run_as_user() {
  if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
    local user_home
    user_home="$(getent passwd "$SUDO_USER" | cut -d: -f6)"
    sudo -H -u "$SUDO_USER" env HOME="$user_home" "$@"
  else
    "$@"
  fi
}

install_apt_packages() {
  if ! command -v apt-get >/dev/null 2>&1; then
    cat <<'MSG' >&2
Missing required base tooling, and apt-get is not available.

Install Node.js, npm, Go, and shellcheck manually, then re-run this script.
On macOS, use Homebrew:
  brew install node go shellcheck
MSG
    exit 1
  fi

  local apt_prefix=()
  if [ "$(id -u)" -ne 0 ]; then
    if ! command -v sudo >/dev/null 2>&1; then
      printf 'Missing sudo; install Node.js, npm, Go, and shellcheck manually.\n' >&2
      exit 1
    fi
    apt_prefix=(sudo)
  fi

  log "Installing Debian/Ubuntu base packages"
  "${apt_prefix[@]}" apt-get update
  "${apt_prefix[@]}" apt-get install -y git nodejs npm golang-go shellcheck
}

missing_base_tools=()
for tool in node npm go shellcheck; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    missing_base_tools+=("$tool")
  fi
done

if [ "${#missing_base_tools[@]}" -gt 0 ]; then
  printf 'Missing base tooling: %s\n' "${missing_base_tools[*]}"
  install_apt_packages
fi

log "Checking base tooling"
for tool in node npm go shellcheck; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    printf 'Missing required command after installation attempt: %s\n' "$tool" >&2
    exit 1
  fi
done

log "Installing Node developer dependencies"
run_as_user npm install

log "Installing Go developer tools"
run_as_user go install mvdan.cc/gofumpt@latest
run_as_user go install golang.org/x/tools/cmd/goimports@latest
run_as_user go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
run_as_user go install mvdan.cc/sh/v3/cmd/shfmt@latest

log "Installing Git hooks"
run_as_user npx lefthook install

cat <<'MSG'

Developer tooling is installed.

Ensure your Go bin directory is on PATH, for example:
  export PATH="$(go env GOPATH)/bin:$PATH"
MSG
