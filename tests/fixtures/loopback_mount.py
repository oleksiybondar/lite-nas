"""Test support for isolated ext4 loopback mounts."""

import re
import shlex
import subprocess
import uuid
from collections.abc import Iterator, Sequence
from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

import pytest
from constants import SECURITY_LOGGING_MANAGER_CLI_BINARY
from fixtures.logging_manager_cli import wait_for_logging_manager_event
from hyperiontf import CLIClient, expect

EICAR_SAMPLE = r"X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"
SAFE_MOUNT_PREFIX = re.compile(r"[a-z0-9-]{1,32}")
ALLOWED_TEST_UTILITIES = {"mount", "mkfs.ext4", "truncate", "umount"}
ALLOWED_PRIVILEGED_TEST_UTILITIES = {"mkdir", "mount", "rmdir", "umount"}


@dataclass
class LoopbackMount:
    """Represent one mounted test image and its explicit cleanup operation."""

    image_path: Path
    mount_point: Path
    privileged_cli_client: CLIClient
    mounted: bool = True

    def unmount(self) -> None:
        """Unmount this image once, preserving the mount point for event checks."""
        if not self.mounted:
            return
        must_run_command(
            ["umount", "--", str(self.mount_point)],
            privileged=True,
            cli_client=self.privileged_cli_client,
        )
        self.mounted = False

    def cleanup(self) -> None:
        """Unmount if needed and remove the isolated image and directory."""
        self.unmount()
        must_run_command(
            ["rmdir", "--", str(self.mount_point)],
            privileged=True,
            cli_client=self.privileged_cli_client,
        )
        self.image_path.unlink(missing_ok=True)


@dataclass
class RamfsMount:
    """Represent one temporary ramfs mount and its explicit cleanup operation."""

    mount_point: Path
    privileged_cli_client: CLIClient
    mounted: bool = True

    def unmount(self) -> None:
        """Unmount the ramfs once while keeping its mount directory available."""
        if not self.mounted:
            return
        must_run_command(
            ["umount", "--", str(self.mount_point)],
            privileged=True,
            cli_client=self.privileged_cli_client,
        )
        self.mounted = False

    def cleanup(self) -> None:
        """Unmount the ramfs and remove its unique mount directory."""
        self.unmount()
        must_run_command(
            ["rmdir", "--", str(self.mount_point)],
            privileged=True,
            cli_client=self.privileged_cli_client,
        )


def must_run_command(
    arguments: Sequence[str],
    *,
    privileged: bool = False,
    cli_client: CLIClient | None = None,
) -> None:
    """Run fixture-owned utilities directly or through the restricted sudo test account."""
    allowed_utilities = ALLOWED_PRIVILEGED_TEST_UTILITIES if privileged else ALLOWED_TEST_UTILITIES
    if not arguments or arguments[0] not in allowed_utilities:
        raise ValueError
    if privileged:
        if cli_client is None:
            raise ValueError
        cli_client.execute(shlex.join(["sudo", "-n", "--", *arguments]))
        cli_client.assert_exit_code(0)
        return

    command = list(arguments)
    # Callers use fixture-owned utilities and generated, option-safe path operands.
    subprocess.run(  # noqa: S603
        command,
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
    cli_client: CLIClient,
) -> Iterator[LoopbackMount]:
    """Build, mount, and reliably clean one isolated ext4 test filesystem."""
    if not SAFE_MOUNT_PREFIX.fullmatch(prefix):
        raise ValueError

    mount_root = Path("/mnt")
    must_run_command(
        ["mkdir", "--parents", "--", str(mount_root)],
        privileged=True,
        cli_client=cli_client,
    )
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

    must_run_command(["mkdir", "--", str(mount_point)], privileged=True, cli_client=cli_client)
    resource: LoopbackMount | None = None
    try:
        must_run_command(["truncate", "--size=64M", "--", str(image_path)])
        must_run_command(["mkfs.ext4", "-F", "-d", str(payload_root), "--", str(image_path)])
        must_run_command(
            [
                "mount",
                "-o",
                "loop",
                "-t",
                "ext4",
                "--",
                str(image_path),
                str(mount_point),
            ],
            privileged=True,
            cli_client=cli_client,
        )
        resource = LoopbackMount(
            image_path=image_path,
            mount_point=mount_point,
            privileged_cli_client=cli_client,
        )
        yield resource
    finally:
        if resource is not None:
            resource.cleanup()
        elif mount_point.exists():
            must_run_command(
                ["rmdir", "--", str(mount_point)],
                privileged=True,
                cli_client=cli_client,
            )
        image_path.unlink(missing_ok=True)


@pytest.fixture
def loopback_mount(
    tmp_path_factory: pytest.TempPathFactory,
    testsudo_cli_client: CLIClient,
    security_cli_client: CLIClient,
) -> Iterator[LoopbackMount]:
    """Keep a clean ext4 mount available until Antivirus reports a complete scan."""
    # HyperionTF 0.7.4 reads pytest's temporary-path report during finalization;
    # tmp_path deletes that report before HyperionTF can read it.
    with create_loopback_mount(
        tmp_path_factory.mktemp("security-mount"),
        prefix="litenas-security-test",
        infected=False,
        cli_client=testsudo_cli_client,
    ) as mount:
        event = wait_for_logging_manager_event(
            security_cli_client,
            SECURITY_LOGGING_MANAGER_CLI_BINARY,
            category="Antivirus scan",
            source="antivirus",
            text=f"Antivirus scan clean on {mount.mount_point}",
            status="normal",
        )
        expect(event).not_to_be(None)
        yield mount


@pytest.fixture
def infected_loopback_mount(
    tmp_path_factory: pytest.TempPathFactory, testsudo_cli_client: CLIClient
) -> Iterator[LoopbackMount]:
    """Create an ext4 loopback mount containing the harmless EICAR test signature."""
    with create_loopback_mount(
        tmp_path_factory.mktemp("antivirus-mount"),
        prefix="litenas-antivirus-test",
        infected=True,
        cli_client=testsudo_cli_client,
    ) as mount:
        yield mount


@pytest.fixture
def ramfs_mount(testsudo_cli_client: CLIClient) -> Iterator[RamfsMount]:
    """Create an isolated ramfs mount below `/mnt` for lifecycle alert checks."""
    prefix = "litenas-ramfs-test"
    if not SAFE_MOUNT_PREFIX.fullmatch(prefix):
        raise ValueError
    mount_root = Path("/mnt")
    must_run_command(
        ["mkdir", "--parents", "--", str(mount_root)],
        privileged=True,
        cli_client=testsudo_cli_client,
    )
    mount_point = mount_root / f"{prefix}-{uuid.uuid4().hex[:12]}"
    if not mount_point.is_absolute() or mount_point.name.startswith("-"):
        raise ValueError
    must_run_command(
        ["mkdir", "--", str(mount_point)],
        privileged=True,
        cli_client=testsudo_cli_client,
    )
    resource: RamfsMount | None = None
    try:
        must_run_command(
            [
                "mount",
                "--types=ramfs",
                "--options=size=8m",
                "--",
                "ramfs",
                str(mount_point),
            ],
            privileged=True,
            cli_client=testsudo_cli_client,
        )
        resource = RamfsMount(
            mount_point=mount_point,
            privileged_cli_client=testsudo_cli_client,
        )
        yield resource
    finally:
        if resource is not None:
            resource.cleanup()
        elif mount_point.exists():
            must_run_command(
                ["rmdir", "--", str(mount_point)],
                privileged=True,
                cli_client=testsudo_cli_client,
            )
