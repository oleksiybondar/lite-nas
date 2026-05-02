#!/usr/bin/env bash

if [ -n "${LITE_NAS_ADMIN_PANEL_ASSETS_HELPERS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_ADMIN_PANEL_ASSETS_HELPERS_LOADED=1

adminPanelAssets.sourceFile() {
	local source_dir="$1"
	local file_name="$2"

	if [ -f "$source_dir/$file_name" ]; then
		printf '%s\n' "$source_dir/$file_name"
		return 0
	fi

	if [ -f "$source_dir/assets/$file_name" ]; then
		printf '%s\n' "$source_dir/assets/$file_name"
		return 0
	fi

	return 1
}

adminPanelAssets.validateBuildOutput() {
	local source_dir="$1"
	local required_file
	local required_files=(
		index.html
		index.css
		index.js
	)

	if [ ! -d "$source_dir" ]; then
		log.error "Missing admin-panel assets source: $source_dir"
		exit 1
	fi

	for required_file in "${required_files[@]}"; do
		if adminPanelAssets.sourceFile "$source_dir" "$required_file" >/dev/null; then
			continue
		fi

		log.error "Missing admin-panel asset: $required_file in $source_dir"
		exit 1
	done
}

adminPanelAssets.installFlat() {
	local source_dir="$1"
	local target_dir="$2"
	local file_name
	local source_file
	local files=(
		index.html
		index.css
		index.js
	)

	adminPanelAssets.validateBuildOutput "$source_dir"

	rm -rf "$target_dir"
	install -d -m 0755 "$target_dir"

	for file_name in "${files[@]}"; do
		source_file="$(adminPanelAssets.sourceFile "$source_dir" "$file_name")"
		install -m 0644 "$source_file" "$target_dir/$file_name"
	done

	if source_file="$(adminPanelAssets.sourceFile "$source_dir" favicon.ico)"; then
		install -m 0644 "$source_file" "$target_dir/favicon.ico"
	else
		log.warn "favicon.ico is not present in $source_dir; /favicon.ico will return 404 until it is added."
	fi
}
