#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/nginx.sh"

# Installs the loopback Moments API and its independently served browser assets.
deploy.moments.deploy() {
	local binary_path="$1"
	local assets_source="$2"
	local should_start="${3:-1}"
	binary_path="$(realpath -- "$binary_path")"
	assets_source="$(realpath -- "$assets_source")"

	if [ ! -f "$binary_path" ] || [ ! -f "$assets_source/index.html" ] || [ ! -d "$assets_source/assets" ]; then
		log.error "Missing Moments binary or frontend assets."
		exit 1
	fi

	deploy.ensureSystemGroup lite-nas-moments "Moments service group"
	if ! id lite-nas-moments >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin --gid lite-nas-moments --groups lite-nas lite-nas-moments
	fi

	if [ "$(realpath -- "$binary_path")" = "$(realpath -m -- /usr/libexec/lite-nas/moments)" ]; then
		chmod 0755 /usr/libexec/lite-nas/moments
	else
		install -D -m 0755 "$binary_path" /usr/libexec/lite-nas/moments
	fi
	deploy.installMergedConfig "$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/moments.conf" /etc/lite-nas/moments.conf \
		lite-nas-moments lite-nas
	install -d -m 0750 -o lite-nas-moments -g lite-nas-moments /var/lib/lite-nas/moments
	install -d -m 0750 -o lite-nas-moments -g lite-nas-moments /var/lib/lite-nas/moments/thumbnails
	install -d -m 0751 /var/log/lite-nas
	if [ ! -f /var/log/lite-nas/moments.log ]; then
		install -D -m 0640 -o lite-nas-moments -g lite-nas-moments /dev/null /var/log/lite-nas/moments.log
	fi
	install -D -m 0644 "$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-moments.service" /etc/systemd/system/lite-nas-moments.service

	install -d -m 0755 /usr/share/lite-nas/moments /usr/share/lite-nas/nginx
	cp -R "$assets_source/." /usr/share/lite-nas/moments/
	if [ -f "$LITE_NAS_REPO_ROOT/services/moments/caption_image.py" ]; then
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/caption_image.py" /usr/share/lite-nas/moments/caption_image.py
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/install_caption_model.py" /usr/share/lite-nas/moments/install_caption_model.py
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/embed_text.py" /usr/share/lite-nas/moments/embed_text.py
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/install_embedding_model.py" /usr/share/lite-nas/moments/install_embedding_model.py
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/install_e5_model.py" /usr/share/lite-nas/moments/install_e5_model.py
	elif [ ! -f /usr/share/lite-nas/moments/caption_image.py ] || [ ! -f /usr/share/lite-nas/moments/install_caption_model.py ]; then
		log.error "Moments caption scripts are missing"
		return 1
	fi
	local place_data_source="$LITE_NAS_REPO_ROOT/services/moments/data/cities1000.zip"
	if [ -f "$place_data_source" ]; then
		install -m 0644 "$place_data_source" /usr/share/lite-nas/moments/cities1000.zip
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/data/README.md" /usr/share/lite-nas/moments/GEONAMES-ATTRIBUTION.md
	elif [ ! -f /usr/share/lite-nas/moments/cities1000.zip ]; then
		log.error "Moments offline city data is missing"
		return 1
	fi
	find /usr/share/lite-nas/moments -type d -exec chmod 0755 {} +
	find /usr/share/lite-nas/moments -type f -exec chmod 0644 {} +
	if [ "${LITE_NAS_PACKAGE_VALIDATE:-0}" != "1" ]; then
		if ! deploy.moments.installCaptionModel; then
			log.warn "BLIP caption setup failed; Moments will keep scanning media without generated captions."
		fi
		if ! deploy.moments.installEmbeddingModel; then
			log.warn "Semantic model setup failed; Moments will continue using text search."
		fi
	fi
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nginx-certificates.sh" --if-missing
	deploy.nginx.deploy 0
	if [ "$should_start" = "1" ]; then
		deploy.nginx.ensureCertificateAssets
		deploy.nginx.validateConfig
		deploy.moments.enableAndStart
		deploy.nginx.enableAndStart
	fi
}

# Installs the Python caption stack and downloads pinned model weights only once.
deploy.moments.installCaptionModel() {
	local venv_dir=/usr/local/lib/lite-nas/moments/caption-venv
	local model_dir=/usr/local/lib/lite-nas/moments/blip-model
	local legacy_model_dir=/var/lib/lite-nas/moments/blip-model
	local model_installer=/usr/share/lite-nas/moments/install_caption_model.py

	install -d -m 0755 -o root -g root /usr/local/lib/lite-nas/moments || return 1

	if ! python3 -c 'import PIL, torch, venv' >/dev/null 2>&1; then
		log.warn "BLIP requires python3-pil, python3-torch, and python3-venv; install the LiteNAS package through APT to resolve these dependencies."
		return 1
	fi
	if [ ! -x "$venv_dir/bin/python3" ] || ! "$venv_dir/bin/python3" -m pip --version >/dev/null 2>&1; then
		python3 -m venv --system-site-packages "$venv_dir" || return 1
	fi
	if ! "$venv_dir/bin/python3" -c 'import importlib.metadata; assert importlib.metadata.version("transformers") == "4.57.6"' >/dev/null 2>&1; then
		"$venv_dir/bin/python3" -m pip install --no-cache-dir --disable-pip-version-check 'transformers==4.57.6' || return 1
	fi
	"$venv_dir/bin/python3" -c 'import PIL, torch; from transformers import BlipForConditionalGeneration, BlipProcessor' || return 1
	"$venv_dir/bin/python3" "$model_installer" "$model_dir" "$legacy_model_dir" || return 1
	chown -R root:lite-nas-moments "$model_dir" || return 1
	find "$model_dir" -type d -exec chmod 0750 {} + || return 1
	find "$model_dir" -type f -exec chmod 0640 {} + || return 1
}

# Installs a distinct local text-model runtime; verified weights are reused.
deploy.moments.installEmbeddingModel() {
	local venv_dir=/usr/local/lib/lite-nas/moments/embedding-venv
	local model_dir=/usr/local/lib/lite-nas/moments/embeddinggemma-model
	local e5_dir=/usr/local/lib/lite-nas/moments/e5-model
	local model_installer=/usr/share/lite-nas/moments/install_embedding_model.py
	local e5_installer=/usr/share/lite-nas/moments/install_e5_model.py

	install -d -m 0755 -o root -g root /usr/local/lib/lite-nas/moments || return 1
	if ! python3 -c 'import venv' >/dev/null 2>&1; then
		return 1
	fi
	if [ ! -x "$venv_dir/bin/python3" ] || ! "$venv_dir/bin/python3" -m pip --version >/dev/null 2>&1; then
		python3 -m venv --copies --system-site-packages "$venv_dir" || return 1
	fi
	if ! "$venv_dir/bin/python3" -c 'import importlib.metadata; assert importlib.metadata.version("transformers") == "4.57.6"' >/dev/null 2>&1; then
		"$venv_dir/bin/python3" -m pip install --no-cache-dir --disable-pip-version-check 'transformers==4.57.6' || return 1
	fi
	if ! "$venv_dir/bin/python3" -c 'import onnxruntime, numpy' >/dev/null 2>&1; then
		local onnxruntime_version
		onnxruntime_version="$(deploy.moments.onnxRuntimeVersion "$venv_dir/bin/python3")" || return 1
		"$venv_dir/bin/python3" -m pip install --no-cache-dir --disable-pip-version-check "onnxruntime==$onnxruntime_version" || return 1
	fi
	local directory installer model
	for model in e5 embeddinggemma; do
		if [ "$model" = e5 ]; then
			directory="$e5_dir"
			installer="$e5_installer"
		else
			directory="$model_dir"
			installer="$model_installer"
		fi
		"$venv_dir/bin/python3" "$installer" "$directory" || return 1
		chown -R root:lite-nas-moments "$directory" || return 1
		find "$directory" -type d -exec chmod 0750 {} + || return 1
		find "$directory" -type f -exec chmod 0640 {} + || return 1
		deploy.moments.smokeTestEmbedding "$venv_dir/bin/python3" "$model" || return 1
	done
}

# Prints the ONNX Runtime version to install for the given interpreter.
# 1.23.2 has no wheels for CPython 3.14 (first available is 1.24.1, verified on
# aarch64); older interpreters keep the previously tested 1.23.2.
deploy.moments.onnxRuntimeVersion() {
	local python_bin="$1"
	if "$python_bin" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 14) else 1)'; then
		echo "1.24.1"
	else
		echo "1.23.2"
	fi
}

# Runs real inference through the service's embed_text.py so a bare import
# success cannot hide a model that fails to load. Failure is reported to the
# caller, which downgrades it to a warning and keeps text search.
deploy.moments.smokeTestEmbedding() {
	local python_bin="$1"
	local model="$2"
	local size=384
	local query_prefix='query: '
	local passage_prefix='passage: '
	if [ "$model" = embeddinggemma ]; then
		size=768
		query_prefix='task: search result | query: '
		passage_prefix='title: none | text: '
	fi
	local response
	response="$(printf '{"model":"%s","texts":["%sмаленька дівчинка і квіти","%sa little girl holding flowers","%sa mountain road"]}\n' "$model" "$query_prefix" "$passage_prefix" "$passage_prefix" |
		"$python_bin" /usr/share/lite-nas/moments/embed_text.py)" || return 1
	printf '%s' "$response" | "$python_bin" -c '
import base64, json, math, struct, sys
vectors = []
for encoded in json.load(sys.stdin)["vectors"]:
    raw = base64.b64decode(encoded)
    vec = struct.unpack("<%df" % (len(raw) // 4), raw)
    assert len(vec) == int(sys.argv[1]) and abs(math.sqrt(sum(v * v for v in vec)) - 1.0) < 1e-3
    vectors.append(vec)
assert len(vectors) == 3
score = lambda left, right: sum(a * b for a, b in zip(left, right))
assert score(vectors[0], vectors[1]) > score(vectors[0], vectors[2])
 ' "$size" || return 1
}

# Print a systemd quota equal to 35% of all online CPU capacity.
deploy.moments.cpuQuota() {
	local online_cpus
	online_cpus="$(getconf _NPROCESSORS_ONLN)" || return 1
	case "$online_cpus" in
	'' | *[!0-9]*) return 1 ;;
	esac
	if [ "$online_cpus" -lt 1 ] || [ "$online_cpus" -gt 4096 ]; then
		return 1
	fi
	printf '%s%%\n' "$((35 * online_cpus))"
}

# Enables the Moments service with a quota shared by the API and model workers.
deploy.moments.enableAndStart() {
	if deploy.hasUsableSystemd; then
		local quota dropin_dir
		quota="$(deploy.moments.cpuQuota)" || quota=35%
		dropin_dir=/etc/systemd/system/lite-nas-moments.service.d
		install -d -m 0755 "$dropin_dir" || return 1
		printf '[Service]\nCPUQuota=%s\n' "$quota" >"$dropin_dir/35-percent-total-cpu.conf" || return 1
		chmod 0644 "$dropin_dir/35-percent-total-cpu.conf" || return 1
		systemctl daemon-reload || return 1
	fi
	deploy.enableAndRefreshService lite-nas-moments.service
}
