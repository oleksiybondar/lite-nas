#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

package_roots=(
	"packaging/debian/lite-nas"
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

log.pushTask "Validating Debian package structure (Lintian)"
for package_root in "${package_roots[@]}"; do
	if [ -d "$package_root" ] && command -v lintian >/dev/null 2>&1; then
		# Lintian expects a .deb or a directory that looks like an unpacked package (with DEBIAN/control).
		# However, our template uses control.in and placeholders.
		# We create a temporary unpacked package structure for lintian to analyze.
		analysis_dir="$(mktemp -d)"
		mkdir -p "$analysis_dir/DEBIAN"
		sed \
			-e "s|@PACKAGE_ARCH@|amd64|g" \
			-e "s|@PACKAGE_VERSION@|0.0.0+ci|g" \
			"$package_root/DEBIAN/control.in" >"$analysis_dir/DEBIAN/control"

		# Copy other DEBIAN files if they exist
		for f in config templates postinst postrm prerm; do
			if [ -f "$package_root/DEBIAN/$f" ]; then
				cp "$package_root/DEBIAN/$f" "$analysis_dir/DEBIAN/$f"
			fi
		done

		# Copy the payload structure
		if [ -d "$package_root/usr" ]; then
			cp -a "$package_root/usr" "$analysis_dir/"
		fi

		# Compress changelog to satisfy Lintian
		if [ -f "$analysis_dir/usr/share/doc/lite-nas/changelog.Debian" ]; then
			gzip -n -9 "$analysis_dir/usr/share/doc/lite-nas/changelog.Debian"
			mv "$analysis_dir/usr/share/doc/lite-nas/changelog.Debian.gz" \
				"$analysis_dir/usr/share/doc/lite-nas/changelog.gz"
		fi

		# Normalize permissions for Lintian (simulate dpkg-deb behavior)
		find "$analysis_dir" -type d -exec chmod 0755 {} +
		find "$analysis_dir" -type f -exec chmod 0644 {} +
		if [ -d "$analysis_dir/DEBIAN" ]; then
			# Maintainer scripts must be executable, control and templates must NOT
			find "$analysis_dir/DEBIAN" -type f -exec chmod 0755 {} +
			chmod 0644 "$analysis_dir/DEBIAN/control"
			if [ -f "$analysis_dir/DEBIAN/templates" ]; then
				chmod 0644 "$analysis_dir/DEBIAN/templates"
			fi
		fi

		# Create a minimal .deb for lintian if it doesn't support directory analysis.
		# Some lintian versions strictly require a .deb file.
		# We build a lightweight package without heavy binaries to keep analysis fast.
		log.requireCommand "dpkg-deb" "Install dpkg-deb for package analysis."
		analysis_deb="${analysis_dir}.deb"
		dpkg-deb --root-owner-group --build "$analysis_dir" "$analysis_deb" >/dev/null

		# Run lintian on the prepared .deb.
		lintian --fail-on error --display-experimental --pedantic "$analysis_deb" || {
			res=$?
			rm -f "$analysis_deb"
			rm -rf "$analysis_dir"
			exit $res
		}
		rm -f "$analysis_deb"
		rm -rf "$analysis_dir"
	fi
done
log.popTask
