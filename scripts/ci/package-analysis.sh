#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

package_roots=(
	"packaging/debian/lite-nas"
	"packaging/debian/lite-nas-system-metrics"
)
tmp_dir=""
tmp_package_dir=""
maintainer_scripts=()
for package_root in "${package_roots[@]}"; do
	for maintainer_script in config postinst postrm prerm; do
		if [ -f "$package_root/DEBIAN/$maintainer_script" ]; then
			maintainer_scripts+=("$package_root/DEBIAN/$maintainer_script")
		fi
	done
done

log.pushTask "Running shellcheck for Debian maintainer scripts"
log.requireCommand "shellcheck" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-shell-dependencies.sh."
shellcheck -s sh "${maintainer_scripts[@]}"
log.popTask

if command -v nats-server >/dev/null 2>&1 && command -v openssl >/dev/null 2>&1; then
	log.pushTask "Validating NATS configuration"
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_package_dir" "${tmp_dir:-}"' EXIT
	cp -a configs/etc/. "$tmp_dir/"
	openssl req \
		-x509 \
		-newkey rsa:2048 \
		-sha256 \
		-days 1 \
		-nodes \
		-keyout "$tmp_dir/server.key" \
		-out "$tmp_dir/server.crt" \
		-subj "/CN=lite-nas-nats-server" >/dev/null 2>&1
	cp "$tmp_dir/server.crt" "$tmp_dir/root-ca.crt"
	sed \
		-e "s|/etc/nats-server/certificates/server.crt|$tmp_dir/server.crt|g" \
		-e "s|/etc/nats-server/certificates/server.key|$tmp_dir/server.key|g" \
		-e "s|/etc/nats-server/certificates/root-ca.crt|$tmp_dir/root-ca.crt|g" \
		configs/etc/nats-server/tls.conf >"$tmp_dir/nats-server/tls.conf"
	nats-server -t -c "$tmp_dir/nats-server.conf"
	log.popTask
fi

if command -v systemd-analyze >/dev/null 2>&1; then
	log.pushTask "Validating systemd unit template"
	if [ -z "${tmp_dir:-}" ]; then
		tmp_dir="$(mktemp -d)"
		trap 'rm -rf "$tmp_package_dir" "${tmp_dir:-}"' EXIT
	fi
	sed \
		-e 's|@SYSTEM_METRICS_BINARY@|/bin/true|g' \
		-e 's|@SYSTEM_METRICS_CONFIG_DIR@|/etc/liteNAS|g' \
		-e 's|@SYSTEM_METRICS_CONFIG_GROUP@|lite-nas|g' \
		-e 's|@SYSTEM_METRICS_LOG_FILE@|/var/log/liteNAS/system-metrics.log|g' \
		-e 's|@SYSTEM_METRICS_RUNTIME_GROUP@|lite-nas-system-metrics|g' \
		-e 's|@SYSTEM_METRICS_RUNTIME_USER@|lite-nas-system-metrics|g' \
		configs/systemd/system/lite-nas-system-metrics.service >"$tmp_dir/lite-nas-system-metrics.service"
	systemd-analyze verify "$tmp_dir/lite-nas-system-metrics.service"
	log.popTask
fi

log.pushTask "Building Debian package for analysis"
tmp_package_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_package_dir" "${tmp_dir:-}"' EXIT
./scripts/package/build-all-debs.sh \
	--version=0.0.0+ci \
	--output-dir="$tmp_package_dir"
log.popTask

if command -v lintian >/dev/null 2>&1; then
	log.pushTask "Running lintian on analysis packages"
	lintian --fail-on error --display-experimental --pedantic \
		"$tmp_package_dir"/lite-nas*_0.0.0+ci_*.deb
	log.popTask
fi
