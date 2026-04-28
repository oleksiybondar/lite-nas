#!/usr/bin/env bash

if [ -n "${LITE_NAS_PACKAGING_HELPERS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_PACKAGING_HELPERS_LOADED=1

package.renderTemplate() {
	local package_arch="$1"
	local package_version="$2"
	local template_path="$3"
	local destination_path="$4"

	sed \
		-e "s|@PACKAGE_ARCH@|$package_arch|g" \
		-e "s|@PACKAGE_VERSION@|$package_version|g" \
		"$template_path" >"$destination_path"
}

package.copyTree() {
	local source_dir="$1"
	local destination_dir="$2"

	mkdir -p "$destination_dir"
	cp -a "$source_dir/." "$destination_dir/"
}

package.prepareRoot() {
	local package_root="$1"

	rm -rf "$package_root"
	mkdir -p "$package_root/DEBIAN"
}

package.installMaintainerScript() {
	local template_dir="$1"
	local package_root="$2"
	local script_name="$3"

	if [ ! -f "$template_dir/DEBIAN/$script_name" ]; then
		return 0
	fi

	cp "$template_dir/DEBIAN/$script_name" "$package_root/DEBIAN/$script_name"
	chmod 0755 "$package_root/DEBIAN/$script_name"
}

package.requireBuildCommands() {
	log.requireCommand "dpkg-deb" "Install dpkg-deb and retry."
	log.requireCommand "gzip" "Install gzip and retry."
}

package.resolveArchitecture() {
	dpkg --print-architecture
}

package.prepareMetadata() {
	local package_arch="$1"
	local package_version="$2"
	local template_dir="$3"
	local package_root="$4"

	package.prepareRoot "$package_root"
	package.renderTemplate \
		"$package_arch" \
		"$package_version" \
		"$template_dir/DEBIAN/control.in" \
		"$package_root/DEBIAN/control"
}

package.copyDocTreeAndCompressChangelog() {
	local source_dir="$1"
	local destination_dir="$2"
	local changelog_path="$3"

	package.copyTree "$source_dir" "$destination_dir"
	gzip -n -9 "$changelog_path"
}

package.writeMd5sums() {
	local package_root="$1"

	(
		cd "$package_root" || exit 1
		while IFS= read -r -d '' file; do
			printf '%s  %s\n' \
				"$(md5sum "$file" | awk '{print $1}')" \
				"${file#./}"
		done < <(find . -path './DEBIAN' -prune -o -type f -print0 | LC_ALL=C sort -z) >DEBIAN/md5sums
	)
}
