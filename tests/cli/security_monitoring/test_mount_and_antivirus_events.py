"""System CLI tests for mount lifecycle and antivirus security events."""

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    SYSTEM_LOGGING_MANAGER_CLI_BINARY,
)
from fixtures.logging_manager_cli import wait_for_logging_manager_event
from fixtures.loopback_mount import LoopbackMount, RamfsMount
from hyperiontf import CLIClient, expect

pytestmark = [pytest.mark.cli]
EVENT_QUERY_BY_SOURCE = {
    "mount-monitor": (SYSTEM_LOGGING_MANAGER_CLI_BINARY, "Mount lifecycle", "active"),
    "antivirus": (SECURITY_LOGGING_MANAGER_CLI_BINARY, "Malware detection", "normal"),
}


def wait_for_event(
    cli_client: CLIClient,
    *,
    source: str,
    text: str,
) -> dict[str, object] | None:
    """Poll one manager CLI until a matching event reaches the requested state."""
    cli_binary, category, status = EVENT_QUERY_BY_SOURCE[source]
    return wait_for_logging_manager_event(
        cli_client, cli_binary, category=category, source=source, text=text, status=status
    )


@pytest.mark.MountMonitor
@pytest.mark.SystemLoggingManager
def test_mount_event_is_emitted_for_a_new_mount(
    loopback_mount: LoopbackMount,
    operator_cli_client: CLIClient,
) -> None:
    """Test case: a new filesystem mount creates an active mount event.

    Preparation:
    - Mount Monitor and System Logging Manager are running.
    - A unique ext4 loopback image is mounted below `/mnt`.

    Action:
    - Query the System Logging Manager CLI for the event for that mount.

    Expected result:
    - An active `Mount lifecycle` event identifies the mounted filesystem
      with a `sysmnt` counter ID.

    Cleanup:
    - The mount fixture unmounts the loopback image during teardown.

    Requirement: security-monitoring/FR-002, logging-managers/FR-003.
    """
    event = wait_for_event(
        operator_cli_client,
        source="mount-monitor",
        text=str(loopback_mount.mount_point),
    )
    expect(str((event or {}).get("EventID", ""))).to_match(r"^sysmnt_[0-9]{8}$")


@pytest.mark.Antivirus
@pytest.mark.SecurityLoggingManager
def test_clean_mount_scan_reports_informational_result(
    loopback_mount: LoopbackMount,
    security_cli_client: CLIClient,
) -> None:
    """Test case: a clean ext4 mount produces a completed Antivirus scan event.

    Preparation:
    - Mount Monitor, Antivirus, and Security Logging Manager are running.
    - A clean ext4 image is mounted below `/mnt` until the scan completes.

    Action:
    - Query Security Logging Manager for the clean scan result.

    Expected result:
    - An informational, normal Antivirus event identifies the mounted filesystem.

    Requirement: security-monitoring/FR-003.
    """
    event = wait_for_logging_manager_event(
        security_cli_client,
        SECURITY_LOGGING_MANAGER_CLI_BINARY,
        category="Antivirus scan",
        source="antivirus",
        text=f"Antivirus scan clean on {loopback_mount.mount_point}",
        status="normal",
    )
    expect((event or {}).get("Severity")).to_be("info")


@pytest.mark.MountMonitor
@pytest.mark.SystemLoggingManager
def test_ramfs_mount_emits_informational_event_with_filesystem_type(
    ramfs_mount: RamfsMount,
    operator_cli_client: CLIClient,
) -> None:
    """Test case: mounting ramfs produces an informational event naming its filesystem.

    Preparation:
    - Mount Monitor and System Logging Manager are running.
    - A unique `ramfs` filesystem is mounted under `/mnt`.

    Action:
    - Query the System Logging Manager CLI for the new mount event.

    Expected result:
    - An informational `Mount lifecycle` event identifies `ramfs` and the
      unique mount point with a `sysmnt` counter ID.

    Requirement: security-monitoring/FR-002, logging-managers/FR-001.
    """
    event = wait_for_event(
        operator_cli_client,
        source="mount-monitor",
        text=f"(ramfs): {ramfs_mount.mount_point}",
    )
    expect(str((event or {}).get("EventID", ""))).to_match(r"^sysmnt_[0-9]{8}$")
    expect((event or {}).get("Severity")).to_be("info")


@pytest.mark.Antivirus
@pytest.mark.SecurityLoggingManager
def test_eicar_mount_creates_one_resolved_antivirus_event(
    infected_loopback_mount: LoopbackMount,
    security_cli_client: CLIClient,
) -> None:
    """Test case: an EICAR signature on a unique mount creates one resolved malware event.

    Preparation:
    - Mount Monitor, Antivirus, and Security Logging Manager are running.
    - A unique ext4 loopback image under `/mnt` contains the EICAR test signature
      in two files before it is mounted, so the scanner sees both on its initial scan.

    Action:
    - Allow Antivirus to scan the newly observed mount and query events through
      the Security Logging Manager CLI.

    Expected result:
    - One `Malware detection` event from Antivirus is normal after registration
      and has a `secav` counter ID and no occurrence row for this one scan.

    Requirement: security-monitoring/FR-003, logging-managers/FR-001.
    """
    event = wait_for_event(
        security_cli_client,
        source="antivirus",
        text=f"{infected_loopback_mount.mount_point} (2 findings)",
    )
    expect(str((event or {}).get("EventID", ""))).to_match(r"^secav_[0-9]{8}$")
    expect((event or {}).get("LastRecID")).to_be(None)
