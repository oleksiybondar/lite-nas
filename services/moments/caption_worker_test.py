"""Local unit tests for the optional Moments caption worker inputs and weights."""

import base64
import binascii
import errno
import hashlib
import importlib.util
import pathlib
import sys
import tempfile
import types
import unittest
from unittest.mock import patch

import caption_image


def load_installer():
    """Load the optional installer with a controlled Hugging Face dependency."""
    hub = types.SimpleNamespace(snapshot_download=lambda **_kwargs: None)
    script = pathlib.Path(__file__).with_name("install_caption_model.py")
    spec = importlib.util.spec_from_file_location("caption_installer_for_test", script)
    if spec is None or spec.loader is None:
        raise RuntimeError("caption installer could not be loaded")
    with patch.dict(sys.modules, {"huggingface_hub": hub}):
        installer = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(installer)
    return installer


def make_model_fixture(model_dir, installer):
    """Create a tiny complete model with replacement checksums for reuse tests."""
    model_dir.mkdir()
    files = {name: name.encode() for name in installer.FILE_SHA256}
    for name, data in files.items():
        (model_dir / name).write_bytes(data)
    (model_dir / ".lite-nas-revision").write_text(installer.MODEL_REVISION)
    return {name: hashlib.sha256(data).hexdigest() for name, data in files.items()}


class OversizeImage:
    """Expose dimensions that exceed the worker's pixel limit before decoding."""

    width = 10_000
    height = 10_000

    def __enter__(self):
        """Return this controlled image from Pillow's context manager."""
        return self

    def __exit__(self, *_args):
        """Leave the controlled image without suppressing errors."""
        return False


class CaptionWorkerTests(unittest.TestCase):
    """Verify that untrusted image input is rejected before inference."""

    def test_rejects_invalid_base64(self):
        """Malformed encoded bytes must not reach the image decoder."""
        image_module = types.SimpleNamespace(DecompressionBombWarning=Warning)
        with patch.dict(sys.modules, {"PIL": types.SimpleNamespace(Image=image_module)}):
            with self.assertRaises(binascii.Error):
                caption_image.decode_image("!")

    def test_rejects_excessive_image_pixels(self):
        """A small compressed input cannot expand beyond the pixel limit."""
        image_module = types.SimpleNamespace(
            DecompressionBombWarning=Warning,
            open=lambda *_args, **_kwargs: OversizeImage(),
        )
        with patch.dict(sys.modules, {"PIL": types.SimpleNamespace(Image=image_module)}):
            with self.assertRaisesRegex(ValueError, "image exceeds pixel limit"):
                caption_image.decode_image(base64.b64encode(b"image").decode())


class CaptionInstallerTests(unittest.TestCase):
    """Verify pinned model weights without importing the optional hub package."""

    def test_rejects_different_safetensors_bytes(self):
        """A local weights file with a different digest must fail validation."""
        installer = load_installer()
        with tempfile.TemporaryDirectory() as directory:
            model = pathlib.Path(directory, "model")
            hashes = make_model_fixture(model, installer)
            (model / "model.safetensors").write_bytes(b"different weights")
            with patch.object(installer, "FILE_SHA256", hashes):
                self.assertFalse(installer.valid_model(model))

    def test_reuses_verified_legacy_model_without_download(self):
        """An unchanged pinned model moves to protected storage without a network call."""
        installer = load_installer()
        with tempfile.TemporaryDirectory() as directory:
            legacy = pathlib.Path(directory, "legacy")
            target = pathlib.Path(directory, "protected")
            target.mkdir()
            hashes = make_model_fixture(legacy, installer)
            with (
                patch.object(installer, "FILE_SHA256", hashes),
                patch.object(installer.os, "chown"),
                patch.object(
                    installer, "snapshot_download", side_effect=AssertionError("downloaded")
                ),
                patch.object(sys, "argv", ["install_caption_model.py", str(target), str(legacy)]),
            ):
                installer.main()
                self.assertTrue(installer.valid_model(target))

    def test_reuses_installed_model_on_subsequent_deploy(self):
        """A verified protected model never asks the hub to download again."""
        installer = load_installer()
        with tempfile.TemporaryDirectory() as directory:
            target = pathlib.Path(directory, "protected")
            hashes = make_model_fixture(target, installer)
            with (
                patch.object(installer, "FILE_SHA256", hashes),
                patch.object(
                    installer, "snapshot_download", side_effect=AssertionError("downloaded")
                ),
                patch.object(sys, "argv", ["install_caption_model.py", str(target)]),
            ):
                installer.main()
                self.assertTrue(installer.valid_model(target))

    def test_rejects_changed_legacy_configuration(self):
        """A stale or altered model configuration cannot be migrated as pinned."""
        installer = load_installer()
        with tempfile.TemporaryDirectory() as directory:
            legacy = pathlib.Path(directory, "legacy")
            target = pathlib.Path(directory, "protected")
            hashes = make_model_fixture(legacy, installer)
            (legacy / "config.json").write_bytes(b"changed")
            with patch.object(installer, "FILE_SHA256", hashes):
                self.assertFalse(installer.reuse_legacy_model(target, legacy))

    def test_copies_legacy_model_across_filesystems_without_download(self):
        """A verified model can be reused when rename cannot cross mounts."""
        installer = load_installer()
        with tempfile.TemporaryDirectory() as directory:
            legacy = pathlib.Path(directory, "legacy")
            target = pathlib.Path(directory, "protected")
            hashes = make_model_fixture(legacy, installer)
            with (
                patch.object(installer, "FILE_SHA256", hashes),
                patch.object(
                    pathlib.Path, "rename", side_effect=OSError(errno.EXDEV, "cross-device")
                ),
                patch.object(installer.os, "chown"),
            ):
                installer.reuse_legacy_model(target, legacy)
                self.assertTrue(installer.valid_model(target))


if __name__ == "__main__":
    unittest.main()
