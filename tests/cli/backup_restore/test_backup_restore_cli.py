"""Installed backup and restore CLI happy paths through HyperionTF."""

import re
from uuid import uuid4

import pytest
from cli.backup_restore.conftest import (
    backup_command,
    must_applications,
    must_metadata_labels,
)
from hyperiontf import CLIClient, RESTClient, expect

pytestmark = [pytest.mark.cli, pytest.mark.BackupRestore]

VERSION_LINE = re.compile(r"[0-9]{8}T[0-9]{6}\.[0-9]+Z")


def must_run_backup_cli(client: CLIClient, arguments: list[str], timeout: int = 120) -> str:
    """Run an expected successful CLI action and return its output."""
    client.execute(backup_command(arguments), timeout=timeout)
    client.assert_exit_code(0)
    return str(client.output)


def parse_backup_all_outcomes(
    output: str, applications: list[str], reported: set[tuple[str, str]]
) -> dict[str, tuple[str, str]]:
    """Collect application outcomes and register CLI-reported archives for cleanup."""
    outcomes: dict[str, tuple[str, str]] = {}
    for line in output.splitlines():
        parts = line.split()
        if len(parts) < 2 or parts[0] not in applications:  # noqa: PLR2004
            continue
        if parts[1] not in {"completed", "skipped", "failed"}:
            continue
        version = parts[2] if len(parts) >= 3 else ""  # noqa: PLR2004
        outcomes[parts[0]] = (parts[1], version)
        if parts[1] == "completed" and VERSION_LINE.fullmatch(version):
            reported.add((parts[0], version))
    return outcomes


# Requirements: backup-restore/IR-003.
def test_cli_lists_configured_applications(admin_cli_client: CLIClient) -> None:
    """Test case: the administrator CLI lists installed backup applications.

    Preparation:
    - The backup service and NATS are running, and the dedicated administrator
      has completed the CLI login flow.

    Action:
    - Run the applications inventory command.

    Expected result:
    - The installed lite-nas application appears in the CLI output.
    """
    output = must_run_backup_cli(admin_cli_client, ["applications"])
    expect("lite-nas" in output.splitlines()).to_be(True)


# Requirements: backup-restore/FR-001, backup-restore/IR-003.
def test_cli_creates_named_backup(
    admin_cli_client: CLIClient,
    api_client: RESTClient,
    admin_cookie: str,
    backup_artifacts: tuple[str, set[tuple[str, str]]],
) -> None:
    """Test case: CLI labels are stored in a new lite-nas archive.

    Preparation:
    - A dedicated administrator is authenticated in the CLI and gateway.
    - A unique name identifies the disposable archive for cleanup.

    Action:
    - Create a lite-nas backup with a name and description through the CLI.

    Expected result:
    - The archive metadata contains the submitted labels.
    """
    backup_marker, reported = backup_artifacts
    description = f"cli-description-{uuid4().hex}"
    output = must_run_backup_cli(
        admin_cli_client,
        ["backup", "lite-nas", "--name", backup_marker, "--description", description],
    )
    version = next((line for line in output.splitlines() if VERSION_LINE.fullmatch(line)), None)
    if version is None:
        pytest.fail(f"backup CLI did not print a version: {output}")
    reported.add(("lite-nas", version))
    labels = must_metadata_labels(api_client, admin_cookie, "lite-nas", version)
    expect(labels).to_be((backup_marker, description))


# Requirements: backup-restore/FR-008, backup-restore/IR-003.
def test_cli_creates_all_backups(
    admin_cli_client: CLIClient,
    api_client: RESTClient,
    admin_cookie: str,
    backup_artifacts: tuple[str, set[tuple[str, str]]],
) -> None:
    """Test case: backup all visits every configured application.

    Preparation:
    - The backup service, configured application plugins, and NATS are running.
    - A unique name identifies every new archive for cleanup.

    Action:
    - Run backup all with a name and description through the CLI.

    Expected result:
    - Every configured application has an outcome, with at least one completed backup.
    """
    backup_marker, reported = backup_artifacts
    applications = must_applications(api_client, admin_cookie)
    description = "Hyperion CLI backup all"
    output = must_run_backup_cli(
        admin_cli_client,
        ["backup", "all", "--name", backup_marker, "--description", description],
        timeout=300,
    )
    outcomes = parse_backup_all_outcomes(output, applications, reported)
    completed = [
        (app, version) for app, (status, version) in outcomes.items() if status == "completed"
    ]
    labels_match = all(
        must_metadata_labels(api_client, admin_cookie, app, version) == (backup_marker, description)
        for app, version in completed
    )
    expect((set(outcomes), bool(completed), labels_match)).to_be((set(applications), True, True))
