"""Administrator CLI and cleanup fixtures for backup system tests."""

import os
import re
import shlex
from collections.abc import Generator
from http import HTTPStatus
from pathlib import PurePosixPath
from typing import Any
from uuid import uuid4

import pytest
from constants import BACKUP_RESTORE_CLI_BINARY
from fixtures.api_client import RedactTestPassword
from fixtures.cli_client import authenticate_cli_client_as_user
from hyperiontf import CLIClient, RESTClient

COOKIE = "lite-nas-at"
BASE = "/api/backup-restore"
APPLICATION_PATTERN = re.compile(r"[A-Za-z0-9][A-Za-z0-9_-]{0,63}")
VERSION_PATTERN = re.compile(r"[A-Za-z0-9][A-Za-z0-9_.-]{0,63}")


def backup_command(arguments: list[str]) -> str:
    """Quote a configured binary and validated CLI operands for Hyperion's shell."""
    binary = PurePosixPath(BACKUP_RESTORE_CLI_BINARY)
    if (
        not binary.is_absolute()
        or ".." in binary.parts
        or not re.fullmatch(r"/[A-Za-z0-9_./-]+", str(binary))
    ):
        raise ValueError("invalid backup CLI binary path")  # noqa: TRY003
    return shlex.join([str(binary), *arguments])


def must_api_data(api_client: RESTClient, cookie: str, path: str) -> Any:
    """Return one successful gateway data field or fail fixture preparation."""
    response = api_client.get(path=path, cookies={COOKIE: cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"backup API read failed at {path}: HTTP {response.status}")
    return response.body["data"]


def must_applications(api_client: RESTClient, cookie: str) -> list[str]:
    """Read validated configured application IDs before building API paths."""
    data = must_api_data(api_client, cookie, f"{BASE}/applications")
    if not isinstance(data, list) or not all(
        isinstance(item, str) and APPLICATION_PATTERN.fullmatch(item) for item in data
    ):
        pytest.fail("backup API returned malformed application IDs")
    return data


def must_versions(api_client: RESTClient, cookie: str, application: str) -> set[str]:
    """Read validated version IDs for one configured application."""
    data = must_api_data(api_client, cookie, f"{BASE}/applications/{application}/versions")
    if not isinstance(data, list) or not all(
        isinstance(item, str) and VERSION_PATTERN.fullmatch(item) for item in data
    ):
        pytest.fail(f"backup API returned malformed versions for {application}")
    return set(data)


def optional_label(value: Any) -> str | None:
    """Validate one optional archive label from the external metadata response."""
    if value is not None and not isinstance(value, str):
        pytest.fail("backup API returned a malformed archive label")
    return value


def must_metadata_labels(
    api_client: RESTClient, cookie: str, application: str, version: str
) -> tuple[str | None, str | None]:
    """Read validated optional labels for one archive named by CLI output."""
    if not APPLICATION_PATTERN.fullmatch(application) or not VERSION_PATTERN.fullmatch(version):
        pytest.fail("backup CLI returned an invalid application or version")
    metadata = must_api_data(
        api_client, cookie, f"{BASE}/applications/{application}/versions/{version}"
    )
    if not isinstance(metadata, dict):
        pytest.fail("backup API returned malformed archive metadata")
    return optional_label(metadata.get("name")), optional_label(metadata.get("description"))


@pytest.fixture
def admin_cli_client(cli_client: CLIClient) -> Generator[CLIClient, None, None]:
    """Log in as the dedicated administrator with a fresh, private CLI token."""
    login = os.getenv("LITENAS_FILE_SERVICE_ADMIN_LOGIN")
    password = os.getenv("LITENAS_FILE_SERVICE_ADMIN_PASSWORD")
    if not login or not password:
        pytest.fail("Set the administrator login and password system-test environment variables")
    redactor = RedactTestPassword(password)
    cli_client.logger.addFilter(redactor)
    token_path: str | None = None
    try:
        authenticate_cli_client_as_user(cli_client, login, password)
        cli_client.execute("id -u")
        uid = cli_client.output.strip()
        if not re.fullmatch(r"[0-9]+", uid):
            pytest.fail("administrator CLI returned an invalid UID")
        runtime_dir = f"/run/user/{uid}"
        cli_client.execute(f"test -d {runtime_dir}")
        if cli_client.exit_code != 0:
            pytest.fail(f"administrator runtime directory is missing: {runtime_dir}")
        token_path = f"{runtime_dir}/lite-nas/backup-restore-cli.token"
        cli_client.execute(f"rm -f -- {token_path}")
        cli_client.assert_exit_code(0)
        cli_client.exec_interactive(backup_command(["applications"]))
        cli_client.wait("Username:", timeout=10)
        cli_client.send_keys(login)
        cli_client.wait("Password:", timeout=10)
        cli_client.send_keys(password)
        cli_client.wait(timeout=30)
        cli_client.assert_exit_code(0)
        yield cli_client
    finally:
        try:
            if token_path is not None:
                cli_client.execute(f"rm -f -- {token_path}")
                cli_client.assert_exit_code(0)
        finally:
            cli_client.logger.removeFilter(redactor)


@pytest.fixture
def backup_artifacts(
    api_client: RESTClient, admin_cookie: str
) -> Generator[tuple[str, set[tuple[str, str]]], None, None]:
    """Track unique CLI archives and delete only matching newly created versions."""
    name = f"cli-hyperion-{uuid4().hex}"
    reported: set[tuple[str, str]] = set()
    applications = must_applications(api_client, admin_cookie)
    before = {
        application: must_versions(api_client, admin_cookie, application)
        for application in applications
    }
    try:
        yield name, reported
    finally:
        for application in applications:
            created = must_versions(api_client, admin_cookie, application) - before[application]
            for version in created:
                path = f"{BASE}/applications/{application}/versions/{version}"
                if (application, version) not in reported:
                    metadata = must_api_data(api_client, admin_cookie, path)
                    if not isinstance(metadata, dict):
                        pytest.fail(
                            f"backup API returned malformed metadata for {application}/{version}"
                        )
                    if metadata.get("name") != name:
                        continue
                response = api_client.delete(path=path, cookies={COOKIE: admin_cookie})
                if response.status != HTTPStatus.OK:
                    pytest.fail(f"CLI backup cleanup failed for {application}/{version}")
