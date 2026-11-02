#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

if [ "$#" -ne 1 ] || [ ! -f "${1:-}" ]; then
	log.error "Usage: scripts/ci/run-openscap-release-delta.sh <lite-nas.deb>"
	exit 64
fi

package_path="$(realpath -- "$1")"
if [ "$(dpkg-deb -f "$package_path" Package)" != "lite-nas" ] ||
	[ "$(dpkg-deb -f "$package_path" Architecture)" != "amd64" ]; then
	log.error "Expected an amd64 lite-nas Debian package: $package_path"
	exit 64
fi

repo_root="$(cd "$SCRIPT_DIR/../.." && pwd)"
report_dir="$repo_root/.build/openscap-release-delta"
content_version="0.1.82"
profile="xccdf_org.ssgproject.content_profile_cis_level2_server"
content_url="https://github.com/ComplianceAsCode/content/releases/download/v${content_version}/scap-security-guide-${content_version}.zip"
mkdir -p "$report_dir"

log.pushTask "Installing OpenSCAP assessment tools"
sudo apt-get update
sudo apt-get install -y openscap-scanner
log.popTask

log.pushTask "Downloading verified Ubuntu 24.04 benchmark content"
curl --fail --location --retry 3 --output "$report_dir/content.zip" "$content_url"
curl --fail --location --retry 3 --output "$report_dir/content.zip.sha512" "${content_url}.sha512"
python3 - "$report_dir" <<'PY'
"""Check the publisher's digest and extract only the Ubuntu 24.04 data stream."""
import hashlib
import re
import sys
import zipfile
from pathlib import Path

directory = Path(sys.argv[1])
archive = directory / "content.zip"
digest_file = directory / "content.zip.sha512"
digest = digest_file.read_text(encoding="ascii").split()[0]
if not re.fullmatch(r"[0-9a-fA-F]{128}", digest):
    raise SystemExit("Invalid benchmark SHA-512 digest")
if hashlib.sha512(archive.read_bytes()).hexdigest().lower() != digest.lower():
    raise SystemExit("Benchmark archive checksum mismatch")
with zipfile.ZipFile(archive) as content:
    matches = [name for name in content.namelist() if name.rsplit("/", 1)[-1] == "ssg-ubuntu2404-ds.xml"]
    if len(matches) != 1:
        raise SystemExit(f"Expected one Ubuntu 24.04 data stream, found {len(matches)}")
    (directory / "ssg-ubuntu2404-ds.xml").write_bytes(content.read(matches[0]))
PY
benchmark="$report_dir/ssg-ubuntu2404-ds.xml"
rm -f "$report_dir/content.zip"
oscap info "$benchmark" >"$report_dir/benchmark-info.txt"
if ! grep -Fq "$profile" "$report_dir/benchmark-info.txt"; then
	log.error "Benchmark does not contain the selected profile: $profile"
	exit 1
fi
log.popTask

run_scan() {
	local phase="$1"
	local result=0
	log.pushTask "OpenSCAP ${phase} evaluation"
	# The runner owns the report directory, so its shell writes the scan log.
	# shellcheck disable=SC2024
	if sudo oscap xccdf eval \
		--profile "$profile" \
		--results "$report_dir/${phase}.xml" \
		--results-arf "$report_dir/${phase}.arf.xml" \
		--report "$report_dir/${phase}.html" \
		"$benchmark" >"$report_dir/${phase}.log" 2>&1; then
		result=0
	else
		result=$?
	fi
	if [ "$result" -ne 0 ] && [ "$result" -ne 2 ]; then
		cat "$report_dir/${phase}.log" >&2
		log.error "OpenSCAP ${phase} evaluation failed with status $result"
		return "$result"
	fi
	for output in "$report_dir/${phase}.xml" "$report_dir/${phase}.arf.xml" "$report_dir/${phase}.html"; do
		if [ ! -s "$output" ]; then
			log.error "OpenSCAP did not produce $output"
			return 1
		fi
	done
	log.popTask
}

run_scan before

log.pushTask "Installing the LiteNAS release package"
# The runner owns the report directory, so its shell can write these logs.
# shellcheck disable=SC2024
sudo env DEBIAN_FRONTEND=noninteractive \
	"$repo_root/scripts/package/install-lite-nas-deb.sh" \
	--package "$package_path" --no-install-recommends \
	>"$report_dir/install.log" 2>&1 || {
	cat "$report_dir/install.log" >&2
	exit 1
}
log.popTask

run_scan after

python3 "$SCRIPT_DIR/scap_delta_report.py" \
	--benchmark "$benchmark" \
	--before "$report_dir/before.xml" \
	--after "$report_dir/after.xml" \
	--json "$report_dir/delta.json" \
	--html "$report_dir/delta.html"

{
	printf 'Content version: %s\n' "$content_version"
	printf 'Profile: %s\n' "$profile"
	printf 'Runner image: %s\n' "${ImageVersion:-unknown}"
	printf 'Package SHA-256: '
	sha256sum "$package_path"
	oscap --version
} >"$report_dir/scan-metadata.txt"

log.info "OpenSCAP before, after, and comparison reports: $report_dir"
