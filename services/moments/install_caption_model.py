"""Download one pinned, safetensors-only BLIP caption model for Moments."""

import errno
import hashlib
import os
import pathlib
import shutil
import stat
import sys

from huggingface_hub import snapshot_download

MODEL_ID = "Salesforce/blip-image-captioning-base"
MODEL_REVISION = "dd2023642d900606c2e56725a1d422485e2a0392"
MODEL_SHA256 = "d5fe21105c145477889dde75d86d0c0460f7679871ca73cc0f265edf9cfdf996"
# Raw files at MODEL_REVISION; the manifest permits offline reuse of an old download.
FILE_SHA256 = {
    "config.json": "cccafa8db5eab4217c2b454e0b74aae6f1764647a8917cca5a04d7e370bfda5e",
    "preprocessor_config.json": "065c10ec97edc5081fbd9655b3d9d25e2647ea6d5dab873325e88eed12d80a7d",
    "special_tokens_map.json": "b6d346be366a7d1d48332dbc9fdf3bf8960b5d879522b7799ddba59e76237ee3",
    "tokenizer_config.json": "a0a0147a736d4a13659e980a8dda47cb1a32550019a6f4751ee622a08e6c873b",
    "vocab.txt": "07eced375cec144d27c900241f3e339478dec958f92fddbc551f295c992038a3",
    "model.safetensors": MODEL_SHA256,
}
FILES = [
    "config.json",
    "generation_config.json",
    "model.safetensors",
    "preprocessor_config.json",
    "special_tokens_map.json",
    "tokenizer_config.json",
    "tokenizer.json",
    "vocab.txt",
]


def valid_model(model_dir: pathlib.Path) -> bool:
    """Accept only the pinned revision and exact files needed by BLIP."""
    if model_dir.is_symlink():
        return False
    marker = model_dir / ".lite-nas-revision"
    if (
        not marker.is_file()
        or marker.is_symlink()
        or marker.stat().st_size > 64
        or marker.read_bytes().strip() != MODEL_REVISION.encode()
    ):
        return False
    for name, expected in FILE_SHA256.items():
        path = model_dir / name
        if not path.is_file() or path.is_symlink():
            return False
        if path.stat().st_size > (
            1100 * 1024 * 1024 if name == "model.safetensors" else 1024 * 1024
        ):
            return False
        with path.open("rb") as handle:
            if hashlib.file_digest(handle, "sha256").hexdigest() != expected:
                return False
    return True


def reuse_legacy_model(model_dir: pathlib.Path, legacy_dir: pathlib.Path) -> bool:
    """Relocate verified old files without another network download."""
    if model_dir.is_dir() and not model_dir.is_symlink() and not any(model_dir.iterdir()):
        model_dir.rmdir()
    if (
        model_dir.exists()
        or model_dir.is_symlink()
        or legacy_dir.is_symlink()
        or not valid_model(legacy_dir)
    ):
        return False
    if any(path.is_symlink() for path in legacy_dir.rglob("*")):
        return False
    try:
        legacy_dir.rename(model_dir)
    except OSError as error:
        if error.errno != errno.EXDEV:
            raise
        # Cross-filesystem upgrades need a one-time local copy of verified files.
        model_dir.mkdir()
        try:
            for name in FILE_SHA256:
                shutil.copy2(legacy_dir / name, model_dir / name)
            shutil.copy2(legacy_dir / ".lite-nas-revision", model_dir / ".lite-nas-revision")
        except OSError:
            shutil.rmtree(model_dir)
            raise
    for path in (model_dir, *model_dir.rglob("*")):
        if path.is_symlink():
            raise RuntimeError("BLIP model contains a link after relocation")
        metadata = path.stat()
        if metadata.st_uid != 0 or metadata.st_gid != 0:
            os.chown(path, 0, 0)
        if metadata.st_mode & 0o022:
            path.chmod(stat.S_IMODE(metadata.st_mode) & ~0o022)
    if not valid_model(model_dir):
        (model_dir / ".lite-nas-revision").unlink(missing_ok=True)
        raise RuntimeError("relocated BLIP model failed checksum verification")
    return True


def main() -> None:
    """Reuse verified weights or download missing files for the pinned revision."""
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: install_caption_model.py ABSOLUTE_MODEL_DIR [LEGACY_MODEL_DIR]")
    model_dir = pathlib.Path(sys.argv[1])
    if not model_dir.is_absolute() or ".." in model_dir.parts:
        raise SystemExit("model directory must be clean and absolute")
    if len(sys.argv) == 3:
        legacy_dir = pathlib.Path(sys.argv[2])
        if not legacy_dir.is_absolute() or ".." in legacy_dir.parts:
            raise SystemExit("legacy model directory must be clean and absolute")
        reuse_legacy_model(model_dir, legacy_dir)
    if valid_model(model_dir):
        return
    model_dir.mkdir(parents=True, exist_ok=True)
    snapshot_download(
        repo_id=MODEL_ID,
        revision=MODEL_REVISION,
        local_dir=model_dir,
        allow_patterns=FILES,
        max_workers=2,
    )
    marker = model_dir / ".lite-nas-revision"
    marker.write_text(MODEL_REVISION + "\n")
    if not valid_model(model_dir):
        marker.unlink(missing_ok=True)
        raise RuntimeError("BLIP model files do not match the pinned revision")


if __name__ == "__main__":
    main()
