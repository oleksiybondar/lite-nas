#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_FILE_EXPLORER_SHARE_ROOT="/usr/share/lite-nas/file-explorer"
readonly LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_TARGET="/usr/share/lite-nas/file-explorer/nginx/file-explorer.conf"
readonly LITE_NAS_FILE_EXPLORER_NGINX_ACTIVE_LINK="/usr/share/lite-nas/nginx/file-explorer.conf"
readonly LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_SOURCE="$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas-file-explorer/nginx/file-explorer.conf"

deploy.fileExplorer.deploy() {
	local assets_source="$1"

	if [ ! -f "$assets_source/index.html" ] || [ ! -d "$assets_source/assets" ]; then
		log.error "Missing File Explorer Vite assets in: $assets_source"
		exit 1
	fi
	if [ ! -f "$LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_SOURCE" ]; then
		log.error "Missing File Explorer nginx snippet: $LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_SOURCE"
		exit 1
	fi

	install -d -m 0755 /usr/share/lite-nas /usr/share/lite-nas/nginx
	rm -rf -- "$LITE_NAS_FILE_EXPLORER_SHARE_ROOT"
	install -d -m 0755 "$LITE_NAS_FILE_EXPLORER_SHARE_ROOT"
	cp -R "$assets_source/." "$LITE_NAS_FILE_EXPLORER_SHARE_ROOT/"
	find "$LITE_NAS_FILE_EXPLORER_SHARE_ROOT" -type d -exec chmod 0755 {} +
	find "$LITE_NAS_FILE_EXPLORER_SHARE_ROOT" -type f -exec chmod 0644 {} +
	install -D -m 0644 "$LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_SOURCE" \
		"$LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_TARGET"
	ln -sfn "$LITE_NAS_FILE_EXPLORER_NGINX_SNIPPET_TARGET" \
		"$LITE_NAS_FILE_EXPLORER_NGINX_ACTIVE_LINK"
}
