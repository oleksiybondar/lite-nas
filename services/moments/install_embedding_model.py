"""Install a pinned offline multilingual E5 ONNX model for Moments search."""

import hashlib
import pathlib
import sys

from huggingface_hub import snapshot_download

MODEL_ID = "intfloat/multilingual-e5-small"
MODEL_REVISION = "ccc66d3bcd826f577e26b9a4072cc5fe3a7ad6a3"
FILE_SHA256 = {
    "onnx/model.onnx": "ca456c06b3a9505ddfd9131408916dd79290368331e7d76bb621f1cba6bc8665",
    "onnx/tokenizer.json": "0b44a9d7b51c3c62626640cda0e2c2f70fdacdc25bbbd68038369d14ebdf4c39",
}
FILES = [
    "onnx/model.onnx",
    "onnx/config.json",
    "onnx/tokenizer.json",
    "onnx/tokenizer_config.json",
    "onnx/special_tokens_map.json",
    "onnx/sentencepiece.bpe.model",
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
        if path.stat().st_size > (500 * 1024 * 1024 if name.endswith(".onnx") else 25 * 1024 * 1024):
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
