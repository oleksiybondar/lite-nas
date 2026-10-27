"""Test support for isolated ext4 loopback mounts."""

import os
import re
import subprocess
import uuid
from collections.abc import Iterator, Sequence
from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

import pytest

EICAR_SAMPLE = r"X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"
SAFE_MOUNT_PREFIX = re.compile(r"[a-z0-9-]{1,32}")
ALLOWED_TEST_UTILITIES = {"mount", "mkfs.ext4", "truncate", "umount"}


@dataclass
class LoopbackMount:
    """Represent one mounted test image and its explicit cleanup operation."""

    image_path: Path
    mount_point: Path
    mounted: bool = True

    def unmount(self) -> None:
        """Unmount this image once, preserving the mount point for event checks."""
        if not self.mounted:
            return
        must_run_command(["umount", "--", str(self.mount_point)])
        self.mounted = False

    def cleanup(self) -> None:
        """Unmount if needed and remove the isolated image and directory."""
        self.unmount()
        self.mount_point.rmdir()
        self.image_path.unlink(missing_ok=True)


@dataclass
class RamfsMount:
    """Represent one temporary ramfs mount and its explicit cleanup operation."""

    mount_point: Path
    mounted: bool = True

    def unmount(self) -> None:
        """Unmount the ramfs once while keeping its mount directory available."""
        if not self.mounted:
            return
        must_run_command(["umount", "--", str(self.mount_point)])
        self.mounted = False

    def cleanup(self) -> None:
        """Unmount the ramfs and remove its unique mount directory."""
        self.unmount()
        self.mount_point.rmdir()


def must_run_command(arguments: Sequence[str]) -> None:
    """Run a fixed system utility with separate, validated arguments."""
    if not arguments or arguments[0] not in ALLOWED_TEST_UTILITIES:
        raise ValueError
    # Callers use fixture-owned utilities and generated, option-safe path operands.
    subprocess.run(  # noqa: S603
        list(arguments),
        check=True,
        capture_output=True,
        text=True,
        timeout=60,
    )


@contextmanager
def create_loopback_mount(
    tmp_path: Path,
    *,
    prefix: str,
    infected: bool,
) -> Iterator[LoopbackMount]:
    """Build, mount, and reliably clean one isolated ext4 test filesystem."""
    if os.geteuid() != 0:
        pytest.skip("loopback mount system tests require root privileges")
    if not SAFE_MOUNT_PREFIX.fullmatch(prefix):
        raise ValueError

    mount_root = Path("/mnt")
    mount_root.mkdir(parents=True, exist_ok=True)
    mount_point = mount_root / f"{prefix}-{uuid.uuid4().hex[:12]}"
    image_path = tmp_path / f"{prefix}.img"
    payload_root = tmp_path / "filesystem-root"
    payload_root.mkdir()
    if infected:
        (payload_root / "eicar-one.com").write_text(EICAR_SAMPLE, encoding="ascii")
        (payload_root / "eicar-two.com").write_text(EICAR_SAMPLE, encoding="ascii")
    if not image_path.is_absolute() or image_path.name.startswith("-"):
        raise ValueError
    if not mount_point.is_absolute() or mount_point.name.startswith("-"):
        raise ValueError

    mount_point.mkdir()
    resource: LoopbackMount | None = None
    try:
        must_run_command(["truncate", "--size=64M", "--", str(image_path)])
        must_run_command(["mkfs.ext4", "-F", "-d", str(payload_root), "--", str(image_path)])
        must_run_command(
            ["mount", "-o", "loop", "-t", "ext4", "--", str(image_path), str(mount_point)]
        )
        resource = LoopbackMount(image_path=image_path, mount_point=mount_point)
        yield resource
    finally:
        if resource is not None:
            resource.cleanup()
        elif mount_point.exists():
            mount_point.rmdir()
        image_path.unlink(missing_ok=True)


@pytest.fixture
def loopback_mount(tmp_path: Path) -> Iterator[LoopbackMount]:
    """Create an isolated ext4 loopback mount under the antivirus scan prefix."""
    with create_loopback_mount(tmp_path, prefix="litenas-security-test", infected=False) as mount:
        yield mount


@pytest.fixture
def infected_loopback_mount(tmp_path: Path) -> Iterator[LoopbackMount]:
    """Create an ext4 loopback mount containing the harmless EICAR test signature."""
    with create_loopback_mount(tmp_path, prefix="litenas-antivirus-test", infected=True) as mount:
        yield mount


@pytest.fixture
def ramfs_mount() -> Iterator[RamfsMount]:
    """Create an isolated ramfs mount below `/mnt` for lifecycle alert checks."""
    if os.geteuid() != 0:
        pytest.skip("ramfs mount system tests require root privileges")

    prefix = "litenas-ramfs-test"
    if not SAFE_MOUNT_PREFIX.fullmatch(prefix):
        raise ValueError
    mount_root = Path("/mnt")
    mount_root.mkdir(parents=True, exist_ok=True)
    mount_point = mount_root / f"{prefix}-{uuid.uuid4().hex[:12]}"
    if not mount_point.is_absolute() or mount_point.name.startswith("-"):
        raise ValueError
    mount_point.mkdir()
    resource: RamfsMount | None = None
    try:
        must_run_command(
            ["mount", "--types=ramfs", "--options=size=8m", "--", "ramfs", str(mount_point)]
        )
        resource = RamfsMount(mount_point=mount_point)
        yield resource
    finally:
        if resource is not None:
            resource.cleanup()
        elif mount_point.exists():
            mount_point.rmdir()
