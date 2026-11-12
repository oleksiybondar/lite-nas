"""Install a pinned offline Google EmbeddingGemma ONNX model for Moments."""

import hashlib
import pathlib
import sys

from huggingface_hub import snapshot_download

MODEL_ID = "onnx-community/embeddinggemma-300m-ONNX"
MODEL_REVISION = "75a84c732f1884df76bec365346230e32f582c82"
FILE_SHA256 = {
    "onnx/model_q4.onnx": "ad1dfee81a70f7944b9b9d1cc6e48075b832881cf33fab2f2b248be78f3f0043",
    "onnx/model_q4.onnx_data": "599962c3143b040de2dd05e5975be3e9091dd067cacc6a8f7186e3203bab9e02",
    "tokenizer.json": "4dda02faaf32bc91031dc8c88457ac272b00c1016cc679757d1c441b248b9c47",
}
FILES = [
    "onnx/model_q4.onnx",
    "onnx/model_q4.onnx_data",
    "config.json",
    "tokenizer.json",
    "tokenizer_config.json",
    "special_tokens_map.json",
    "added_tokens.json",
    "tokenizer.model",
]


def valid_model(directory: pathlib.Path) -> bool:
    """Accept only the expected revision with verified model and tokenizer."""
    marker = directory / ".lite-nas-revision"
    if directory.is_symlink() or not marker.is_file() or marker.is_symlink():
        return False
    if marker.stat().st_size > 64 or marker.read_text().strip() != MODEL_REVISION:
        return False
    for name in FILES:
        path = directory / name
        if not path.is_file() or path.is_symlink():
            return False
        limit = 300 * 1024 * 1024 if name.endswith(".onnx_data") else 25 * 1024 * 1024
        if path.stat().st_size > limit:
            return False
    for name, expected in FILE_SHA256.items():
        with (directory / name).open("rb") as handle:
            if hashlib.file_digest(handle, "sha256").hexdigest() != expected:
                return False
    return True


def main() -> None:
    """Reuse verified local files or download the same pinned revision once."""
    if len(sys.argv) != 2:
        raise SystemExit("usage: install_embedding_model.py ABSOLUTE_MODEL_DIR")
    directory = pathlib.Path(sys.argv[1])
    if not directory.is_absolute() or ".." in directory.parts or directory.is_symlink():
        raise SystemExit("model directory must be clean and absolute")
    if valid_model(directory):
        return
    directory.mkdir(parents=True, exist_ok=True)
    snapshot_download(
        repo_id=MODEL_ID,
        revision=MODEL_REVISION,
        local_dir=directory,
        allow_patterns=FILES,
        max_workers=2,
    )
    marker = directory / ".lite-nas-revision"
    marker.write_text(MODEL_REVISION + "\n")
    if not valid_model(directory):
        marker.unlink(missing_ok=True)
        raise RuntimeError("embedding model files do not match the pinned revision")


if __name__ == "__main__":
    main()
